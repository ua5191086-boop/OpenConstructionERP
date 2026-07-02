"""E-01 Schedule: Primavera XER import onto V014 schema + schedule browsing.

XER format: tab-delimited lines — %T <table>, %F <fields...>, %R <values...>,
%E end. We consume PROJECT, PROJWBS, TASK, TASKPRED and CALENDAR (name only).
"""
import io
import uuid
from datetime import datetime
from typing import Optional

from fastapi import APIRouter, HTTPException, Query, UploadFile, File
from psycopg.rows import dict_row

router = APIRouter()

TASK_TYPE = {"TT_Task": "task", "TT_Mile": "start_milestone",
             "TT_FinMile": "finish_milestone", "TT_LOE": "level_of_effort",
             "TT_WBS": "wbs_summary", "TT_Rsrc": "task"}
STATUS = {"TK_NotStart": "not_started", "TK_Active": "in_progress",
          "TK_Complete": "completed"}
REL = {"PR_FS": "FS", "PR_SS": "SS", "PR_FF": "FF", "PR_SF": "SF"}


def _pool():
    from app.main import get_pool
    return get_pool()


def parse_xer(text: str) -> dict[str, list[dict]]:
    tables: dict[str, list[dict]] = {}
    cur, fields = None, []
    for line in text.splitlines():
        if not line:
            continue
        parts = line.split("\t")
        tag = parts[0]
        if tag == "%T":
            cur = parts[1]
            tables[cur] = []
        elif tag == "%F":
            fields = parts[1:]
        elif tag == "%R" and cur:
            row = dict(zip(fields, parts[1:]))
            tables[cur].append(row)
        elif tag == "%E":
            break
    return tables


def _d(v: Optional[str]):
    if not v:
        return None
    try:
        return datetime.strptime(v.split(" ")[0], "%Y-%m-%d").date()
    except ValueError:
        return None


def _hours_to_days(v: Optional[str]) -> int:
    try:
        return max(round(float(v or 0) / 8.0), 0)
    except ValueError:
        return 0


@router.post("/{project_id}/schedules/import-xer", status_code=201)
async def import_xer(project_id: uuid.UUID, file: UploadFile = File(...),
                     schedule_code: Optional[str] = None):
    if not file.filename.lower().endswith(".xer"):
        raise HTTPException(400, "Only .xer files are supported")
    raw = await file.read()
    for enc in ("utf-8", "cp1251", "latin-1"):
        try:
            text = raw.decode(enc)
            break
        except UnicodeDecodeError:
            continue
    else:
        raise HTTPException(400, "Cannot decode file")
    t = parse_xer(text)
    tasks = t.get("TASK", [])
    if not tasks:
        raise HTTPException(400, "No TASK table found — is this a valid XER export?")
    proj_rows = t.get("PROJECT", [])
    wbs = {w.get("wbs_id"): w.get("wbs_short_name") or w.get("wbs_name", "")
           for w in t.get("PROJWBS", [])}
    p6_name = (proj_rows[0].get("proj_short_name") if proj_rows else None) or "P6 import"
    code = schedule_code or f"XER-{datetime.now():%Y%m%d-%H%M}"

    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        proj = await (await conn.execute(
            "SELECT id FROM projects WHERE id=%s", (project_id,))).fetchone()
        if not proj:
            raise HTTPException(404, "Project not found")
        try:
            sch = await (await conn.execute(
                """INSERT INTO schedules (project_id, schedule_code, schedule_name,
                                          schedule_type, data_date, created_by)
                   VALUES (%s,%s,%s,'baseline',%s,'xer-import') RETURNING id""",
                (project_id, code, p6_name,
                 _d(proj_rows[0].get("last_recalc_date")) if proj_rows else None),
            )).fetchone()
        except Exception:
            raise HTTPException(409, f"Schedule code {code} already exists")
        sch_id = sch["id"]

        id_map: dict[str, uuid.UUID] = {}   # P6 task_id -> our UUID
        imported = skipped = 0
        for task in tasks:
            start = _d(task.get("target_start_date")) or _d(task.get("act_start_date"))
            finish = _d(task.get("target_end_date")) or _d(task.get("act_end_date")) or start
            if not start:
                skipped += 1
                continue
            row = await (await conn.execute(
                """INSERT INTO schedule_activities
                     (schedule_id, activity_id, wbs_code, activity_name,
                      activity_type, status, original_duration,
                      remaining_duration, percent_complete,
                      early_start, early_finish, actual_start, actual_finish,
                      start_date, finish_date, float_total)
                   VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)
                   RETURNING id""",
                (sch_id, task.get("task_code", f"T{imported}"),
                 wbs.get(task.get("wbs_id"), ""),
                 (task.get("task_name") or "")[:500],
                 TASK_TYPE.get(task.get("task_type"), "task"),
                 STATUS.get(task.get("status_code"), "not_started"),
                 _hours_to_days(task.get("target_drtn_hr_cnt")),
                 _hours_to_days(task.get("remain_drtn_hr_cnt")),
                 float(task.get("phys_complete_pct") or 0),
                 _d(task.get("early_start_date")), _d(task.get("early_end_date")),
                 _d(task.get("act_start_date")), _d(task.get("act_end_date")),
                 start, finish,
                 _hours_to_days(task.get("total_float_hr_cnt"))))).fetchone()
            id_map[task.get("task_id")] = row["id"]
            imported += 1

        rels = 0
        for pred in t.get("TASKPRED", []):
            a = id_map.get(pred.get("pred_task_id"))
            b = id_map.get(pred.get("task_id"))
            if not a or not b:
                continue
            await conn.execute(
                """INSERT INTO schedule_relationships
                     (schedule_id, predecessor_id, successor_id, relation_type, lag_days)
                   VALUES (%s,%s,%s,%s,%s) ON CONFLICT DO NOTHING""",
                (sch_id, a, b, REL.get(pred.get("pred_type"), "FS"),
                 _hours_to_days(pred.get("lag_hr_cnt"))))
            rels += 1

    return {"schedule_id": sch_id, "schedule_code": code, "p6_project": p6_name,
            "activities_imported": imported, "activities_skipped": skipped,
            "relationships": rels}


@router.get("/{project_id}/schedules")
async def list_schedules(project_id: uuid.UUID):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        return await (await conn.execute(
            """SELECT s.*,
                      (SELECT count(*) FROM schedule_activities a
                        WHERE a.schedule_id=s.id) AS activities,
                      (SELECT min(start_date) FROM schedule_activities a
                        WHERE a.schedule_id=s.id) AS starts,
                      (SELECT max(finish_date) FROM schedule_activities a
                        WHERE a.schedule_id=s.id) AS finishes
               FROM schedules s WHERE s.project_id=%s
               ORDER BY s.created_at DESC""",
            (project_id,))).fetchall()


@router.get("/{project_id}/schedules/{schedule_id}/activities")
async def list_activities(project_id: uuid.UUID, schedule_id: uuid.UUID,
                          q: Optional[str] = None, critical: bool = False,
                          limit: int = Query(200, le=2000), offset: int = 0):
    sql = """SELECT a.* FROM schedule_activities a
             JOIN schedules s ON s.id = a.schedule_id
             WHERE s.project_id=%s AND a.schedule_id=%s"""
    args: list = [project_id, schedule_id]
    if q:
        sql += " AND (a.activity_name ILIKE %s OR a.activity_id ILIKE %s)"
        args += [f"%{q}%", f"%{q}%"]
    if critical:
        sql += " AND COALESCE(a.float_total,0) <= 0 AND a.activity_type='task'"
    sql += " ORDER BY a.start_date, a.activity_id LIMIT %s OFFSET %s"
    args += [limit, offset]
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        return await (await conn.execute(sql, args)).fetchall()


@router.get("/{project_id}/schedules/{schedule_id}/summary")
async def schedule_summary(project_id: uuid.UUID, schedule_id: uuid.UUID):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        s = await (await conn.execute(
            """SELECT count(*) AS activities,
                      count(*) FILTER (WHERE activity_type LIKE '%%milestone%%') AS milestones,
                      count(*) FILTER (WHERE COALESCE(float_total,0) <= 0
                                       AND activity_type='task') AS critical,
                      count(*) FILTER (WHERE status='completed') AS completed,
                      min(start_date) AS starts, max(finish_date) AS finishes,
                      round(avg(percent_complete),1) AS avg_pct
               FROM schedule_activities WHERE schedule_id=%s""",
            (schedule_id,))).fetchone()
        rels = await (await conn.execute(
            "SELECT count(*) AS n FROM schedule_relationships WHERE schedule_id=%s",
            (schedule_id,))).fetchone()
    return {**s, "relationships": rels["n"],
            "duration_days": (s["finishes"] - s["starts"]).days
            if s["starts"] and s["finishes"] else None}
