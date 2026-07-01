"""Minimal ontology API (ADR-001): types, objects, links, graph neighbourhood."""
import uuid
from typing import Optional

from fastapi import APIRouter, HTTPException, Query
from pydantic import BaseModel
from psycopg.rows import dict_row
from psycopg.types.json import Json

router = APIRouter()


def _pool():
    from app.main import get_pool
    return get_pool()


class ObjectIn(BaseModel):
    type_code: str
    project_id: Optional[uuid.UUID] = None
    props: dict = {}


class LinkIn(BaseModel):
    link_type: str
    from_object: uuid.UUID
    to_object: uuid.UUID
    props: dict = {}


@router.get("/object-types")
async def list_object_types():
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        return await (await conn.execute(
            """SELECT t.*, (SELECT count(*) FROM objects o WHERE o.type_id = t.id) AS objects_count
               FROM object_types t ORDER BY t.code"""
        )).fetchall()


@router.get("/objects")
async def list_objects(
    type_code: Optional[str] = None,
    project_id: Optional[uuid.UUID] = None,
    limit: int = Query(50, le=500),
    offset: int = 0,
):
    q = """SELECT o.id, t.code AS type, o.project_id, o.props, o.version, o.updated_at
           FROM objects o JOIN object_types t ON t.id = o.type_id WHERE TRUE"""
    args: list = []
    if type_code:
        q += " AND t.code = %s"; args.append(type_code)
    if project_id:
        q += " AND o.project_id = %s"; args.append(project_id)
    q += " ORDER BY o.updated_at DESC LIMIT %s OFFSET %s"; args += [limit, offset]
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        return await (await conn.execute(q, args)).fetchall()


@router.post("/objects", status_code=201)
async def create_object(o: ObjectIn):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        row = await (await conn.execute(
            """INSERT INTO objects (type_id, project_id, props)
               SELECT id, %s, %s FROM object_types WHERE code = %s
               RETURNING id, project_id, props, version""",
            (o.project_id, Json(o.props), o.type_code),
        )).fetchone()
    if not row:
        raise HTTPException(400, f"Unknown object type: {o.type_code}")
    return row


@router.post("/links", status_code=201)
async def create_link(l: LinkIn):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        try:
            return await (await conn.execute(
                """INSERT INTO links (link_type, from_object, to_object, props)
                   VALUES (%s,%s,%s,%s)
                   ON CONFLICT (link_type, from_object, to_object)
                   DO UPDATE SET props = EXCLUDED.props
                   RETURNING *""",
                (l.link_type, l.from_object, l.to_object, Json(l.props)),
            )).fetchone()
        except Exception as e:
            raise HTTPException(400, f"Cannot create link: {e}")


@router.get("/objects/{object_id}/graph")
async def object_graph(object_id: uuid.UUID, depth: int = Query(1, ge=1, le=2)):
    """Neighbourhood of an object: the ontology's value in one call."""
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        node = await (await conn.execute(
            """SELECT o.id, t.code AS type, o.props FROM objects o
               JOIN object_types t ON t.id = o.type_id WHERE o.id = %s""",
            (object_id,),
        )).fetchone()
        if not node:
            raise HTTPException(404, "Object not found")
        edges = await (await conn.execute(
            """WITH RECURSIVE hop AS (
                 SELECT l.id, l.link_type, l.from_object, l.to_object, 1 AS d
                 FROM links l WHERE l.from_object = %(id)s OR l.to_object = %(id)s
                 UNION
                 SELECT l.id, l.link_type, l.from_object, l.to_object, h.d + 1
                 FROM links l JOIN hop h
                   ON (l.from_object IN (h.from_object, h.to_object)
                    OR l.to_object   IN (h.from_object, h.to_object))
                 WHERE h.d < %(depth)s
               )
               SELECT DISTINCT h.link_type, h.from_object, h.to_object,
                      tf.code AS from_type, tt.code AS to_type,
                      of.props->>'code' AS from_code, ot.props->>'code' AS to_code
               FROM hop h
               JOIN objects of ON of.id = h.from_object
               JOIN objects ot ON ot.id = h.to_object
               JOIN object_types tf ON tf.id = of.type_id
               JOIN object_types tt ON tt.id = ot.type_id
               LIMIT 500""",
            {"id": object_id, "depth": depth},
        )).fetchall()
    return {"node": node, "edges": edges, "edge_count": len(edges)}
