"""Site documents vertical: D-03 RFI + C-05 Daily Reports.
Daily work entries feed physical progress against the BOQ."""
import uuid
from datetime import date, datetime
from typing import Optional

from fastapi import APIRouter, HTTPException, Query
from pydantic import BaseModel, Field
from psycopg.rows import dict_row

from app.routers.projects import mirror_object, link_objects

router = APIRouter()


def _pool():
    from app.main import get_pool
    return get_pool()


# ------------------------------------------------------------------ RFI ----

class RfiIn(BaseModel):
    subject: str = Field(min_length=3, max_length=255)
    question: str
    discipline: Optional[str] = None
    raised_by: Optional[str] = None
    assigned_to: Optional[str] = None
    due_date: Optional[date] = None


class RfiAnswer(BaseModel):
    answer: str
    answered_by: Optional[str] = None


@router.get("/{project_id}/rfis")
async def list_rfis(project_id: uuid.UUID, status: Optional[str] = None):
    sql = """SELECT *,
                    (status = 'open' AND due_date IS NOT NULL
                     AND due_date < CURRENT_DATE) AS overdue,
                    CASE WHEN status = 'open' AND due_date IS NOT NULL
                         THEN due_date - CURRENT_DATE END AS days_to_due
             FROM rfis WHERE project_id = %s"""
    args: list = [project_id]
    if status:
        sql += " AND status = %s"; args.append(status)
    sql += " ORDER BY number DESC"
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        return await (await conn.execute(sql, args)).fetchall()


@router.post("/{project_id}/rfis", status_code=201)
async def create_rfi(project_id: uuid.UUID, r: RfiIn):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        row = await (await conn.execute(
            """INSERT INTO rfis (project_id, number, code, subject, question,
                                 discipline, raised_by, assigned_to, due_date)
               VALUES (%(p)s,
                       COALESCE((SELECT max(number) FROM rfis WHERE project_id=%(p)s),0)+1,
                       'RFI-' || lpad((COALESCE((SELECT max(number) FROM rfis
                                       WHERE project_id=%(p)s),0)+1)::text, 4, '0'),
                       %(su)s, %(q)s, %(d)s, %(rb)s, %(at)s, %(dd)s)
               RETURNING *""",
            {"p": project_id, "su": r.subject, "q": r.question, "d": r.discipline,
             "rb": r.raised_by, "at": r.assigned_to, "dd": r.due_date},
        )).fetchone()
        obj = await mirror_object(conn, "rfi", project_id,
                                  {"code": row["code"], "subject": r.subject,
                                   "status": "open"}, "rfis", row["id"])
        proj_obj = await (await conn.execute(
            """SELECT id FROM objects WHERE source_table='projects' AND source_id=%s""",
            (project_id,),
        )).fetchone()
        if proj_obj:
            await link_objects(conn, "belongs_to", obj, proj_obj["id"])
    return row


@router.post("/{project_id}/rfis/{rfi_id}/answer")
async def answer_rfi(project_id: uuid.UUID, rfi_id: uuid.UUID, a: RfiAnswer):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        row = await (await conn.execute(
            """UPDATE rfis SET answer=%s, status='answered',
                   answered_at=NOW(), updated_at=NOW()
               WHERE id=%s AND project_id=%s AND status='open' RETURNING *""",
            (a.answer, rfi_id, project_id),
        )).fetchone()
    if not row:
        raise HTTPException(409, "RFI not found or not open")
    return row


@router.post("/{project_id}/rfis/{rfi_id}/close")
async def close_rfi(project_id: uuid.UUID, rfi_id: uuid.UUID):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        row = await (await conn.execute(
            """UPDATE rfis SET status='closed', closed_at=NOW(), updated_at=NOW()
               WHERE id=%s AND project_id=%s AND status='answered' RETURNING *""",
            (rfi_id, project_id),
        )).fetchone()
    if not row:
        raise HTTPException(409, "RFI must be answered before closing")
    return row


# -------------------------------------------------------- Daily reports ----

class WorkEntry(BaseModel):
    boq_item_code: str
    qty_done: float = Field(gt=0)
    location: Optional[str] = None
    notes: Optional[str] = None


class DailyReportIn(BaseModel):
    report_date: date
    shift: str = "day"
    weather: Optional[str] = None
    temp_c: Optional[float] = None
    manpower_total: Optional[int] = None
    equipment_total: Optional[int] = None
    narrative: Optional[str] = None
    hse_notes: Optional[str] = None
    delays: Optional[str] = None
    author: Optional[str] = None
    entries: list[WorkEntry] = []


@router.get("/{project_id}/daily-reports")
async def list_daily_reports(project_id: uuid.UUID,
                             limit: int = Query(30, le=200), offset: int = 0):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        return await (await conn.execute(
            """SELECT d.*,
                      (SELECT count(*) FROM daily_work_entries e
                        WHERE e.report_id = d.id) AS entries_count
               FROM daily_reports d WHERE d.project_id = %s
               ORDER BY d.report_date DESC, d.shift LIMIT %s OFFSET %s""",
            (project_id, limit, offset),
        )).fetchall()


@router.post("/{project_id}/daily-reports", status_code=201)
async def create_daily_report(project_id: uuid.UUID, d: DailyReportIn):
    unresolved = []
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        try:
            rep = await (await conn.execute(
                """INSERT INTO daily_reports
                     (project_id, report_date, shift, weather, temp_c,
                      manpower_total, equipment_total, narrative, hse_notes,
                      delays, author, status)
                   VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,'submitted')
                   RETURNING *""",
                (project_id, d.report_date, d.shift, d.weather, d.temp_c,
                 d.manpower_total, d.equipment_total, d.narrative,
                 d.hse_notes, d.delays, d.author),
            )).fetchone()
        except Exception:
            raise HTTPException(409,
                f"Daily report for {d.report_date}/{d.shift} already exists")
        for e in d.entries:
            item = await (await conn.execute(
                "SELECT id FROM boq_items WHERE project_id=%s AND code=%s",
                (project_id, e.boq_item_code),
            )).fetchone()
            if not item:
                unresolved.append(e.boq_item_code)
                continue
            await conn.execute(
                """INSERT INTO daily_work_entries
                     (report_id, boq_item_id, qty_done, location, notes)
                   VALUES (%s,%s,%s,%s,%s)""",
                (rep["id"], item["id"], e.qty_done, e.location, e.notes))
        await mirror_object(conn, "daily_report", project_id,
                            {"code": f"DR-{d.report_date}-{d.shift}",
                             "manpower": d.manpower_total},
                            "daily_reports", rep["id"])
    return {**rep, "entries_saved": len(d.entries) - len(unresolved),
            "unresolved_item_codes": unresolved}


@router.get("/{project_id}/progress/physical")
async def physical_progress(project_id: uuid.UUID, top: int = Query(20, le=200)):
    """Physical progress: qty done (daily reports) vs BOQ quantity, weighted by cost."""
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        items = await (await conn.execute(
            """SELECT i.code, i.name, i.unit, i.quantity, i.total_cost,
                      COALESCE(sum(e.qty_done), 0) AS done
               FROM boq_items i
               LEFT JOIN daily_work_entries e ON e.boq_item_id = i.id
               WHERE i.project_id = %s AND i.status <> 'cancelled'
               GROUP BY i.id
               HAVING COALESCE(sum(e.qty_done), 0) > 0
               ORDER BY i.total_cost DESC LIMIT %s""",
            (project_id, top),
        )).fetchall()
        tot = await (await conn.execute(
            """SELECT sum(i.total_cost) AS plan_cost,
                      sum(LEAST(COALESCE(e.done,0)/NULLIF(i.quantity,0),1)
                          * i.total_cost) AS earned_cost
               FROM boq_items i
               LEFT JOIN (SELECT boq_item_id, sum(qty_done) AS done
                          FROM daily_work_entries GROUP BY 1) e
                 ON e.boq_item_id = i.id
               WHERE i.project_id = %s AND i.status <> 'cancelled'""",
            (project_id,),
        )).fetchone()
    for it in items:
        it["quantity"] = float(it["quantity"]); it["done"] = float(it["done"])
        it["pct"] = round(100 * min(it["done"] / it["quantity"], 1), 1) \
            if it["quantity"] else None
    plan = float(tot["plan_cost"] or 0); earned = float(tot["earned_cost"] or 0)
    return {"items_in_progress": items,
            "earned_value": round(earned, 2),
            "plan_value": round(plan, 2),
            "physical_percent": round(100 * earned / plan, 2) if plan else 0}
