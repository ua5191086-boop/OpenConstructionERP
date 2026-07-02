"""Quality & HSE minimum: M-03 NCR workflow + N-01 Permit to Work (V028 schema)."""
import uuid
from datetime import datetime
from typing import Optional

from fastapi import APIRouter, HTTPException, Query
from pydantic import BaseModel, Field
from psycopg.rows import dict_row

from app.routers.projects import mirror_object, link_objects

router = APIRouter()

NCR_FLOW = {"open": {"disposition", "void"},
            "disposition": {"corrective_action", "void"},
            "corrective_action": {"verification"},
            "verification": {"closed", "corrective_action"},
            "closed": set(), "void": set()}

PTW_FLOW = {"issued": {"active", "cancelled"},
            "active": {"suspended", "closed"},
            "suspended": {"active", "closed", "cancelled"},
            "closed": set(), "cancelled": set()}


def _pool():
    from app.main import get_pool
    return get_pool()


async def _proj_obj(conn, project_id):
    row = await (await conn.execute(
        "SELECT id FROM objects WHERE source_table='projects' AND source_id=%s",
        (project_id,))).fetchone()
    return row["id"] if row else None


# ---------------------------------------------------------------- NCR ----

class NcrIn(BaseModel):
    title: str = Field(min_length=3, max_length=255)
    description: str
    severity: str = Field(default="minor", pattern="^(minor|major|critical)$")
    location: Optional[str] = None
    boq_item_code: Optional[str] = None
    ring_no_ref: Optional[str] = None      # "DR-L/R42"
    raised_by: Optional[str] = None
    assigned_to: Optional[str] = None
    due_date: Optional[str] = None


class NcrTransition(BaseModel):
    status: str
    root_cause: Optional[str] = None
    corrective_action: Optional[str] = None


@router.get("/{project_id}/ncrs")
async def list_ncrs(project_id: uuid.UUID,
                    status: Optional[str] = None,
                    severity: Optional[str] = None):
    sql = """SELECT n.*, b.code AS boq_code,
                    (n.status NOT IN ('closed','void') AND n.due_date IS NOT NULL
                     AND n.due_date < CURRENT_DATE) AS overdue
             FROM ncrs n LEFT JOIN boq_items b ON b.id = n.boq_item_id
             WHERE n.project_id = %s"""
    args: list = [project_id]
    if status:
        sql += " AND n.status = %s"; args.append(status)
    if severity:
        sql += " AND n.severity = %s"; args.append(severity)
    sql += " ORDER BY n.number DESC"
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        return await (await conn.execute(sql, args)).fetchall()


@router.post("/{project_id}/ncrs", status_code=201)
async def create_ncr(project_id: uuid.UUID, n: NcrIn):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        boq_id = None
        if n.boq_item_code:
            r = await (await conn.execute(
                "SELECT id FROM boq_items WHERE project_id=%s AND code=%s",
                (project_id, n.boq_item_code))).fetchone()
            if not r:
                raise HTTPException(400, f"BOQ item {n.boq_item_code} not found")
            boq_id = r["id"]
        ring_id = None
        if n.ring_no_ref and "/" in n.ring_no_ref:
            dcode, rno = n.ring_no_ref.split("/R", 1)
            r = await (await conn.execute(
                """SELECT r.id FROM tunnel_rings r
                   JOIN tunnel_drives d ON d.id = r.drive_id
                   WHERE d.project_id=%s AND d.code=%s AND r.ring_no=%s""",
                (project_id, dcode, int(rno)))).fetchone()
            ring_id = r["id"] if r else None
        row = await (await conn.execute(
            """INSERT INTO ncrs (project_id, number, code, title, description,
                                 severity, location, boq_item_id, ring_id,
                                 raised_by, assigned_to, due_date)
               VALUES (%(p)s,
                       COALESCE((SELECT max(number) FROM ncrs WHERE project_id=%(p)s),0)+1,
                       'NCR-'||lpad((COALESCE((SELECT max(number) FROM ncrs
                                     WHERE project_id=%(p)s),0)+1)::text,4,'0'),
                       %(t)s,%(d)s,%(sv)s,%(l)s,%(bi)s,%(ri)s,%(rb)s,%(at)s,%(dd)s)
               RETURNING *""",
            {"p": project_id, "t": n.title, "d": n.description, "sv": n.severity,
             "l": n.location, "bi": boq_id, "ri": ring_id, "rb": n.raised_by,
             "at": n.assigned_to, "dd": n.due_date})).fetchone()
        obj = await mirror_object(conn, "ncr", project_id,
                                  {"code": row["code"], "title": n.title,
                                   "severity": n.severity, "status": "open"},
                                  "ncrs", row["id"])
        po = await _proj_obj(conn, project_id)
        if po:
            await link_objects(conn, "belongs_to", obj, po)
        if boq_id:
            bo = await (await conn.execute(
                "SELECT id FROM objects WHERE source_table='boq_items' AND source_id=%s",
                (boq_id,))).fetchone()
            if bo:
                await link_objects(conn, "concerns", obj, bo["id"])
    return row


@router.post("/{project_id}/ncrs/{ncr_id}/transition")
async def ncr_transition(project_id: uuid.UUID, ncr_id: uuid.UUID, t: NcrTransition):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        ncr = await (await conn.execute(
            "SELECT * FROM ncrs WHERE id=%s AND project_id=%s",
            (ncr_id, project_id))).fetchone()
        if not ncr:
            raise HTTPException(404, "NCR not found")
        if t.status not in NCR_FLOW.get(ncr["status"], set()):
            raise HTTPException(409,
                f"Illegal NCR transition {ncr['status']} -> {t.status}; "
                f"allowed: {sorted(NCR_FLOW[ncr['status']])}")
        if t.status == "corrective_action" and not (t.corrective_action or ncr["corrective_action"]):
            raise HTTPException(422, "corrective_action text required")
        row = await (await conn.execute(
            """UPDATE ncrs SET status=%s,
                   root_cause=COALESCE(%s, root_cause),
                   corrective_action=COALESCE(%s, corrective_action),
                   closed_at=CASE WHEN %s='closed' THEN NOW() ELSE closed_at END,
                   updated_at=NOW()
               WHERE id=%s RETURNING *""",
            (t.status, t.root_cause, t.corrective_action, t.status, ncr_id))).fetchone()
        await mirror_object(conn, "ncr", project_id,
                            {"code": ncr["code"], "title": ncr["title"],
                             "severity": ncr["severity"], "status": t.status},
                            "ncrs", ncr_id)
    return row


# ------------------------------------------------------ Permit to Work ----

class PermitIn(BaseModel):
    permit_type: str
    description: str
    location: str
    issued_to: str
    issued_by: Optional[str] = None
    contractor: Optional[str] = None
    valid_from: datetime
    valid_to: datetime
    precautions: Optional[str] = None
    gas_test_required: bool = False


class PermitAction(BaseModel):
    notes: Optional[str] = None


@router.get("/{project_id}/permits")
async def list_permits(project_id: uuid.UUID,
                       active_only: bool = Query(False)):
    sql = """SELECT *,
                    (status IN ('issued','active') AND valid_to < NOW()) AS expired
             FROM work_permits WHERE project_id = %s"""
    args: list = [project_id]
    if active_only:
        sql += " AND status IN ('issued','active')"
    sql += " ORDER BY number DESC"
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        return await (await conn.execute(sql, args)).fetchall()


@router.post("/{project_id}/permits", status_code=201)
async def issue_permit(project_id: uuid.UUID, p: PermitIn):
    if p.valid_to <= p.valid_from:
        raise HTTPException(422, "valid_to must be after valid_from")
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        row = await (await conn.execute(
            """INSERT INTO work_permits
                 (project_id, number, code, permit_type, description, location,
                  contractor, issued_to, issued_by, valid_from, valid_to,
                  precautions, gas_test_required)
               VALUES (%(p)s,
                       COALESCE((SELECT max(number) FROM work_permits
                                 WHERE project_id=%(p)s),0)+1,
                       'PTW-'||lpad((COALESCE((SELECT max(number) FROM work_permits
                                     WHERE project_id=%(p)s),0)+1)::text,4,'0'),
                       %(pt)s,%(d)s,%(l)s,%(c)s,%(it)s,%(ib)s,%(vf)s,%(vt)s,%(pr)s,%(g)s)
               RETURNING *""",
            {"p": project_id, "pt": p.permit_type, "d": p.description,
             "l": p.location, "c": p.contractor, "it": p.issued_to,
             "ib": p.issued_by, "vf": p.valid_from, "vt": p.valid_to,
             "pr": p.precautions, "g": p.gas_test_required})).fetchone()
        obj = await mirror_object(conn, "permit", project_id,
                                  {"code": row["code"], "type": p.permit_type,
                                   "location": p.location, "status": "issued"},
                                  "work_permits", row["id"])
        po = await _proj_obj(conn, project_id)
        if po:
            await link_objects(conn, "belongs_to", obj, po)
    return row


@router.post("/{project_id}/permits/{permit_id}/{action}")
async def permit_action(project_id: uuid.UUID, permit_id: uuid.UUID,
                        action: str, body: PermitAction = PermitAction()):
    target = {"activate": "active", "suspend": "suspended",
              "close": "closed", "cancel": "cancelled"}.get(action)
    if not target:
        raise HTTPException(404, "action must be activate/suspend/close/cancel")
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        pt = await (await conn.execute(
            "SELECT * FROM work_permits WHERE id=%s AND project_id=%s",
            (permit_id, project_id))).fetchone()
        if not pt:
            raise HTTPException(404, "Permit not found")
        if target not in PTW_FLOW.get(pt["status"], set()):
            raise HTTPException(409,
                f"Illegal permit transition {pt['status']} -> {target}; "
                f"allowed: {sorted(PTW_FLOW[pt['status']])}")
        if target == "active" and pt["valid_to"] < datetime.now(pt["valid_to"].tzinfo):
            raise HTTPException(409, "Permit validity window has expired — reissue")
        row = await (await conn.execute(
            """UPDATE work_permits SET status=%s,
                   activated_at=CASE WHEN %s='active' AND activated_at IS NULL
                                     THEN NOW() ELSE activated_at END,
                   closed_at=CASE WHEN %s IN ('closed','cancelled')
                                  THEN NOW() ELSE closed_at END,
                   closure_notes=COALESCE(%s, closure_notes),
                   updated_at=NOW()
               WHERE id=%s RETURNING *""",
            (target, target, target, body.notes, permit_id))).fetchone()
        await mirror_object(conn, "permit", project_id,
                            {"code": pt["code"], "type": pt["permit_type"],
                             "location": pt["location"], "status": target},
                            "work_permits", permit_id)
    return row
