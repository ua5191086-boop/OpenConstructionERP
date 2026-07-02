"""Money-change vertical: G-02 Variation Orders + F-05 IPC (payment certificates).
VO: draft -> submitted -> under_evaluation -> approved -> incorporated (budget).
IPC: generated from earned-value delta since the previous certificate."""
import uuid
from datetime import date
from typing import Optional

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field
from psycopg.rows import dict_row

from app.routers.projects import mirror_object, link_objects

router = APIRouter()

VO_FLOW = {"draft": {"submitted", "void"},
           "submitted": {"under_evaluation", "void"},
           "under_evaluation": {"approved", "rejected"},
           "approved": {"incorporated"},
           "rejected": set(), "incorporated": set(), "void": set()}


def _pool():
    from app.main import get_pool
    return get_pool()


# ------------------------------------------------------ Variation Orders ----

class VoItem(BaseModel):
    description: str
    unit: Optional[str] = None
    quantity: float = 0
    unit_price: float = 0
    boq_item_code: Optional[str] = None


class VoIn(BaseModel):
    title: str = Field(min_length=3, max_length=255)
    description: Optional[str] = None
    origin: str = "client_instruction"
    time_impact_days: int = 0
    notice_ref: Optional[str] = None
    created_by: Optional[str] = None
    items: list[VoItem] = []


class VoTransition(BaseModel):
    status: str
    approved_by: Optional[str] = None


@router.get("/{project_id}/variation-orders")
async def list_vos(project_id: uuid.UUID, status: Optional[str] = None):
    sql = """SELECT v.*,
                    (SELECT count(*) FROM variation_order_items i WHERE i.vo_id=v.id) AS items_count
             FROM variation_orders v WHERE v.project_id = %s"""
    args: list = [project_id]
    if status:
        sql += " AND v.status = %s"; args.append(status)
    sql += " ORDER BY v.number DESC"
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        rows = await (await conn.execute(sql, args)).fetchall()
        tot = await (await conn.execute(
            """SELECT COALESCE(sum(cost_impact) FILTER (WHERE status='incorporated'),0) AS incorporated,
                      COALESCE(sum(cost_impact) FILTER (WHERE status IN ('approved')),0) AS approved_pending,
                      COALESCE(sum(cost_impact) FILTER (WHERE status IN ('submitted','under_evaluation')),0) AS in_review
               FROM variation_orders WHERE project_id=%s""",
            (project_id,))).fetchone()
    return {"variation_orders": rows,
            "totals": {k: float(v) for k, v in tot.items()}}


@router.post("/{project_id}/variation-orders", status_code=201)
async def create_vo(project_id: uuid.UUID, v: VoIn):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        vo = await (await conn.execute(
            """INSERT INTO variation_orders
                 (project_id, number, code, title, description, origin,
                  time_impact_days, notice_ref, created_by)
               VALUES (%(p)s,
                       COALESCE((SELECT max(number) FROM variation_orders
                                 WHERE project_id=%(p)s),0)+1,
                       'VO-'||lpad((COALESCE((SELECT max(number) FROM variation_orders
                                    WHERE project_id=%(p)s),0)+1)::text,4,'0'),
                       %(t)s,%(d)s,%(o)s,%(ti)s,%(nr)s,%(cb)s)
               RETURNING *""",
            {"p": project_id, "t": v.title, "d": v.description, "o": v.origin,
             "ti": v.time_impact_days, "nr": v.notice_ref, "cb": v.created_by},
        )).fetchone()
        for it in v.items:
            boq_id = None
            if it.boq_item_code:
                r = await (await conn.execute(
                    "SELECT id FROM boq_items WHERE project_id=%s AND code=%s",
                    (project_id, it.boq_item_code))).fetchone()
                boq_id = r["id"] if r else None
            await conn.execute(
                """INSERT INTO variation_order_items
                     (vo_id, boq_item_id, description, unit, quantity, unit_price)
                   VALUES (%s,%s,%s,%s,%s,%s)""",
                (vo["id"], boq_id, it.description, it.unit, it.quantity, it.unit_price))
        upd = await (await conn.execute(
            """UPDATE variation_orders SET cost_impact =
                 COALESCE((SELECT sum(amount) FROM variation_order_items WHERE vo_id=%s),0)
               WHERE id=%s RETURNING *""",
            (vo["id"], vo["id"]))).fetchone()
        obj = await mirror_object(conn, "variation_order", project_id,
                                  {"code": vo["code"], "title": v.title,
                                   "cost_impact": float(upd["cost_impact"]),
                                   "status": "draft"},
                                  "variation_orders", vo["id"])
        po = await (await conn.execute(
            "SELECT id FROM objects WHERE source_table='projects' AND source_id=%s",
            (project_id,))).fetchone()
        if po:
            await link_objects(conn, "belongs_to", obj, po["id"])
    return upd


@router.post("/{project_id}/variation-orders/{vo_id}/transition")
async def vo_transition(project_id: uuid.UUID, vo_id: uuid.UUID, t: VoTransition):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        vo = await (await conn.execute(
            "SELECT * FROM variation_orders WHERE id=%s AND project_id=%s",
            (vo_id, project_id))).fetchone()
        if not vo:
            raise HTTPException(404, "VO not found")
        if t.status not in VO_FLOW.get(vo["status"], set()):
            raise HTTPException(409,
                f"Illegal VO transition {vo['status']} -> {t.status}; "
                f"allowed: {sorted(VO_FLOW[vo['status']])}")

        budget_version_id = None
        if t.status == "incorporated":
            # Discipline point: incorporation freezes a new budget baseline
            base = await (await conn.execute(
                """SELECT COALESCE(sum(total_cost),0) AS boq FROM boq_items
                   WHERE project_id=%s AND status<>'cancelled'""",
                (project_id,))).fetchone()
            vo_total = await (await conn.execute(
                """SELECT COALESCE(sum(cost_impact),0) AS t FROM variation_orders
                   WHERE project_id=%s AND status='incorporated'""",
                (project_id,))).fetchone()
            new_total = float(base["boq"]) + float(vo_total["t"]) + float(vo["cost_impact"])
            bv = await (await conn.execute(
                """INSERT INTO budget_versions (project_id, version_number, version_name,
                                                status, total_amount, notes)
                   VALUES (%(p)s,
                           COALESCE((SELECT max(version_number)+1 FROM budget_versions
                                     WHERE project_id=%(p)s),1),
                           %(n)s, 'approved', %(t)s, %(nt)s)
                   RETURNING id""",
                {"p": project_id, "n": f"Baseline + {vo['code']}",
                 "t": new_total,
                 "nt": f"Incorporation of {vo['code']}: "
                       f"{float(vo['cost_impact']):+,.2f} {vo['currency']}, "
                       f"EOT {vo['time_impact_days']:+d}d"})).fetchone()
            budget_version_id = bv["id"]
            # Post the approved change into cost control as a Commitment
            ch = await (await conn.execute(
                """SELECT b.cbs_chapter_id FROM variation_order_items i
                   JOIN boq_items b ON b.id = i.boq_item_id
                   WHERE i.vo_id=%s LIMIT 1""", (vo_id,))).fetchone()
            chapter_id = ch["cbs_chapter_id"] if ch else (await (await conn.execute(
                "SELECT id FROM cbs_chapters WHERE code='12.07' AND project_id IS NULL"
            )).fetchone())["id"]
            await conn.execute(
                """INSERT INTO cost_transactions
                     (project_id, cbs_chapter_id, transaction_type, amount,
                      currency, period, description)
                   VALUES (%s,%s,'Commitment',%s,%s,CURRENT_DATE,%s)""",
                (project_id, chapter_id, vo["cost_impact"], vo["currency"],
                 f"{vo['code']} incorporated: {vo['title']}"))

        row = await (await conn.execute(
            """UPDATE variation_orders SET status=%s,
                   submitted_at=CASE WHEN %s='submitted' THEN NOW() ELSE submitted_at END,
                   approved_at=CASE WHEN %s='approved' THEN NOW() ELSE approved_at END,
                   approved_by=COALESCE(%s, approved_by),
                   incorporated_at=CASE WHEN %s='incorporated' THEN NOW()
                                        ELSE incorporated_at END,
                   budget_version_id=COALESCE(%s, budget_version_id),
                   updated_at=NOW()
               WHERE id=%s RETURNING *""",
            (t.status, t.status, t.status, t.approved_by, t.status,
             budget_version_id, vo_id))).fetchone()
        await mirror_object(conn, "variation_order", project_id,
                            {"code": vo["code"], "title": vo["title"],
                             "cost_impact": float(vo["cost_impact"]),
                             "status": t.status},
                            "variation_orders", vo_id)
    return row


# ---------------------------------------------------------------- IPC ----

class IpcIn(BaseModel):
    retention_pct: float = Field(default=5.0, ge=0, le=20)
    advance_recovery_pct: float = Field(default=0.0, ge=0, le=30)
    tax_rate: float = Field(default=0.0, ge=0, le=30)
    notes: Optional[str] = None


@router.post("/{project_id}/ipcs", status_code=201)
async def create_ipc(project_id: uuid.UUID, p: IpcIn):
    """IPC = earned value now − everything already certified. Retention and
    advance recovery deducted; lands in V009 invoices as type 'IPC'."""
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        proj = await (await conn.execute(
            "SELECT code, currency FROM projects WHERE id=%s", (project_id,))).fetchone()
        if not proj:
            raise HTTPException(404, "Project not found")
        ev = await (await conn.execute(
            """SELECT COALESCE(sum(LEAST(COALESCE(e.done,0)/NULLIF(i.quantity,0),1)
                              * i.total_cost),0) AS ev
               FROM boq_items i LEFT JOIN (
                 SELECT boq_item_id, sum(qty_done) done
                 FROM daily_work_entries GROUP BY 1) e ON e.boq_item_id=i.id
               WHERE i.project_id=%s AND i.status<>'cancelled'""",
            (project_id,))).fetchone()
        prev = await (await conn.execute(
            """SELECT COALESCE(sum(amount),0) AS certified, count(*) AS n
               FROM invoices WHERE contract_id IS NULL AND invoice_type='IPC'
                 AND invoice_number LIKE %s AND status<>'cancelled'""",
            (f"IPC-{proj['code']}-%",))).fetchone()
        gross = round(float(ev["ev"]) - float(prev["certified"]), 2)
        if gross <= 0:
            raise HTTPException(409,
                f"Nothing to certify: EV {float(ev['ev']):,.2f} "
                f"already certified {float(prev['certified']):,.2f}")
        retention = round(gross * p.retention_pct / 100, 2)
        advance_rec = round(gross * p.advance_recovery_pct / 100, 2)
        tax = round((gross - retention - advance_rec) * p.tax_rate / 100, 2)
        net = round(gross - retention - advance_rec + tax, 2)
        n = int(prev["n"]) + 1
        inv = await (await conn.execute(
            """INSERT INTO invoices (invoice_number, invoice_type, invoice_date,
                                     due_date, amount, tax_amount, tax_rate,
                                     total_amount, currency, status, notes)
               VALUES (%s,'IPC',CURRENT_DATE,CURRENT_DATE + 28,%s,%s,%s,%s,%s,
                       'issued',%s)
               RETURNING *""",
            (f"IPC-{proj['code']}-{n:03d}", gross, tax, p.tax_rate, net,
             proj["currency"],
             (p.notes or "") + f" | gross {gross:,.2f}; retention "
             f"{p.retention_pct}% = {retention:,.2f}; advance recovery "
             f"{advance_rec:,.2f}"))).fetchone()
    return {**inv, "computed": {"earned_value": float(ev["ev"]),
                                "previously_certified": float(prev["certified"]),
                                "gross": gross, "retention": retention,
                                "advance_recovery": advance_rec, "net": net}}


@router.post("/{project_id}/ipcs/{invoice_id}/pay")
async def pay_ipc(project_id: uuid.UUID, invoice_id: uuid.UUID,
                  payment_ref: Optional[str] = None):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        inv = await (await conn.execute(
            """UPDATE invoices SET status='paid', paid_at=NOW(),
                   payment_ref=COALESCE(%s, payment_ref)
               WHERE id=%s AND invoice_type='IPC' AND status='issued'
               RETURNING *""",
            (payment_ref, invoice_id))).fetchone()
        if not inv:
            raise HTTPException(409, "IPC not found or not in 'issued' status")
        chapter = await (await conn.execute(
            "SELECT id FROM cbs_chapters WHERE code='12.07' AND project_id IS NULL"
        )).fetchone()
        await conn.execute(
            """INSERT INTO cost_transactions
                 (project_id, cbs_chapter_id, transaction_type, amount, currency,
                  period, description)
               VALUES (%s,%s,'Actual',%s,%s,CURRENT_DATE,%s)""",
            (project_id, chapter["id"], inv["total_amount"], inv["currency"],
             f"Payment of {inv['invoice_number']}"))
    return inv


@router.get("/{project_id}/ipcs")
async def list_ipcs(project_id: uuid.UUID):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        proj = await (await conn.execute(
            "SELECT code FROM projects WHERE id=%s", (project_id,))).fetchone()
        return await (await conn.execute(
            """SELECT * FROM invoices WHERE invoice_type='IPC'
                 AND invoice_number LIKE %s ORDER BY invoice_number DESC""",
            (f"IPC-{proj['code']}-%",))).fetchall()
