"""Cost control vertical (F-01 Budget, F-02 Cost Control lite):
plan (BOQ) vs Commitment/Actual/Forecast from cost_transactions; budget snapshots."""
import uuid
from datetime import date
from typing import Optional

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field
from psycopg.rows import dict_row

router = APIRouter()


def _pool():
    from app.main import get_pool
    return get_pool()


class TxIn(BaseModel):
    boq_item_code: str
    transaction_type: str = Field(pattern="^(Plan|Actual|Forecast|Variance|Commitment)$")
    amount: float
    currency: str = "USD"
    period: date
    description: Optional[str] = None


class TxBulk(BaseModel):
    transactions: list[TxIn]


@router.post("/{project_id}/cost/transactions", status_code=201)
async def add_transactions(project_id: uuid.UUID, payload: TxBulk):
    added, errors = 0, []
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        for t in payload.transactions:
            item = await (await conn.execute(
                """SELECT id, cbs_chapter_id FROM boq_items
                   WHERE project_id = %s AND code = %s""",
                (project_id, t.boq_item_code),
            )).fetchone()
            if not item:
                errors.append(f"{t.boq_item_code}: BOQ item not found")
                continue
            await conn.execute(
                """INSERT INTO cost_transactions
                     (project_id, boq_item_id, cbs_chapter_id, transaction_type,
                      amount, currency, period, description)
                   VALUES (%s,%s,%s,%s,%s,%s,%s,%s)""",
                (project_id, item["id"], item["cbs_chapter_id"], t.transaction_type,
                 t.amount, t.currency, t.period, t.description),
            )
            added += 1
    return {"added": added, "errors": errors[:20]}


@router.get("/{project_id}/cost/summary")
async def cost_summary(project_id: uuid.UUID):
    """Per CBS chapter: Plan (BOQ) vs Commitment / Actual / Forecast, variance."""
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        rows = await (await conn.execute(
            """WITH plan AS (
                 SELECT split_part(c.code,'.',1) AS chapter, sum(i.total_cost) AS plan
                 FROM boq_items i JOIN cbs_chapters c ON c.id = i.cbs_chapter_id
                 WHERE i.project_id = %(p)s AND i.status <> 'cancelled'
                 GROUP BY 1
               ), fact AS (
                 SELECT split_part(c.code,'.',1) AS chapter,
                        sum(t.amount_base_currency) FILTER (WHERE t.transaction_type='Actual')     AS actual,
                        sum(t.amount_base_currency) FILTER (WHERE t.transaction_type='Commitment') AS committed,
                        sum(t.amount_base_currency) FILTER (WHERE t.transaction_type='Forecast')   AS forecast
                 FROM cost_transactions t JOIN cbs_chapters c ON c.id = t.cbs_chapter_id
                 WHERE t.project_id = %(p)s
                 GROUP BY 1
               )
               SELECT COALESCE(p.chapter, f.chapter) AS chapter,
                      (SELECT name FROM cbs_chapters
                        WHERE code = COALESCE(p.chapter, f.chapter)
                          AND project_id IS NULL) AS chapter_name,
                      COALESCE(p.plan,0)      AS plan,
                      COALESCE(f.committed,0) AS committed,
                      COALESCE(f.actual,0)    AS actual,
                      COALESCE(f.forecast,0)  AS forecast
               FROM plan p FULL OUTER JOIN fact f ON f.chapter = p.chapter
               ORDER BY 1""",
            {"p": project_id},
        )).fetchall()
    total = {"plan": 0.0, "committed": 0.0, "actual": 0.0, "forecast": 0.0}
    for r in rows:
        for k in total:
            r[k] = float(r[k] or 0); total[k] += r[k]
        r["variance"] = round(r["plan"] - r["actual"] - r["forecast"], 2)
        r["spent_pct"] = round(100 * r["actual"] / r["plan"], 1) if r["plan"] else None
    return {"by_chapter": rows,
            "total": {**{k: round(v, 2) for k, v in total.items()},
                      "variance": round(total["plan"] - total["actual"] - total["forecast"], 2)}}


class BudgetVersionIn(BaseModel):
    version_name: Optional[str] = None
    notes: Optional[str] = None


@router.post("/{project_id}/budget/versions", status_code=201)
async def snapshot_budget(project_id: uuid.UUID, b: BudgetVersionIn):
    """Freeze the current BOQ total as a budget version (baseline discipline)."""
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        total = await (await conn.execute(
            """SELECT COALESCE(sum(total_cost),0) AS t FROM boq_items
               WHERE project_id = %s AND status <> 'cancelled'""",
            (project_id,),
        )).fetchone()
        ver = await (await conn.execute(
            """INSERT INTO budget_versions (project_id, version_number, version_name,
                                            status, total_amount, notes)
               VALUES (%s,
                       COALESCE((SELECT max(version_number)+1 FROM budget_versions
                                 WHERE project_id = %s), 1),
                       %s, 'draft', %s, %s)
               RETURNING *""",
            (project_id, project_id, b.version_name, total["t"], b.notes),
        )).fetchone()
    return ver


@router.get("/{project_id}/budget/versions")
async def list_budget_versions(project_id: uuid.UUID):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        return await (await conn.execute(
            """SELECT * FROM budget_versions WHERE project_id = %s
               ORDER BY version_number DESC""",
            (project_id,),
        )).fetchall()
