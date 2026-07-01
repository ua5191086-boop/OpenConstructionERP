"""Projects API. Every project is mirrored into the ontology graph (ADR-001)."""
import uuid
from typing import Optional

from fastapi import APIRouter, HTTPException
from pydantic import BaseModel, Field
from psycopg.rows import dict_row
from psycopg.types.json import Json

router = APIRouter()


class ProjectIn(BaseModel):
    code: str = Field(min_length=1, max_length=50)
    name: str = Field(min_length=1, max_length=255)
    name_ru: Optional[str] = None
    project_type: str = "metro"
    status: str = "tender"
    country: Optional[str] = None
    currency: str = "USD"
    contract_value: Optional[float] = None


def _pool():
    from app.main import get_pool
    return get_pool()


async def mirror_object(conn, type_code: str, project_id, props: dict,
                        source_table: str, source_id) -> uuid.UUID:
    """Create/update the ontology mirror of a domain record. Returns object id."""
    row = await (await conn.execute(
        """
        INSERT INTO objects (type_id, project_id, props, source_table, source_id)
        SELECT id, %s, %s, %s, %s FROM object_types WHERE code = %s
        ON CONFLICT (source_table, source_id) WHERE source_table IS NOT NULL
        DO UPDATE SET props = EXCLUDED.props,
                      updated_at = NOW(),
                      version = objects.version + 1
        RETURNING id
        """,
        (project_id, Json(props), source_table, source_id, type_code),
    )).fetchone()
    return row["id"] if isinstance(row, dict) else row[0]


async def link_objects(conn, link_type: str, from_obj, to_obj):
    await conn.execute(
        """INSERT INTO links (link_type, from_object, to_object)
           VALUES (%s, %s, %s) ON CONFLICT DO NOTHING""",
        (link_type, from_obj, to_obj),
    )


@router.get("")
async def list_projects():
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        rows = await (await conn.execute(
            """SELECT p.*,
                      (SELECT count(*) FROM boq_items i WHERE i.project_id = p.id) AS boq_items_count,
                      (SELECT COALESCE(sum(total_cost),0) FROM boq_items i
                        WHERE i.project_id = p.id AND i.status <> 'cancelled') AS boq_total
               FROM projects p ORDER BY p.created_at DESC"""
        )).fetchall()
    return rows


@router.post("", status_code=201)
async def create_project(p: ProjectIn):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        try:
            row = await (await conn.execute(
                """INSERT INTO projects (code, name, name_ru, project_type, status,
                                         country, currency, contract_value)
                   VALUES (%s,%s,%s,%s,%s,%s,%s,%s) RETURNING *""",
                (p.code, p.name, p.name_ru, p.project_type, p.status,
                 p.country, p.currency, p.contract_value),
            )).fetchone()
        except Exception as e:
            raise HTTPException(409, f"Cannot create project: {e}")
        await mirror_object(conn, "project", row["id"],
                            {"code": p.code, "name": p.name},
                            "projects", row["id"])
    return row


@router.get("/{project_id}")
async def get_project(project_id: uuid.UUID):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        row = await (await conn.execute(
            "SELECT * FROM projects WHERE id = %s", (project_id,)
        )).fetchone()
    if not row:
        raise HTTPException(404, "Project not found")
    return row
