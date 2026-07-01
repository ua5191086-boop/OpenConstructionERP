"""CDE Core lite (D-01): register with mask numbering, revisions,
ISO 19650 state transitions (WIP -> Shared -> Published -> Archived), transmittals."""
import re
import uuid
from typing import Optional

from fastapi import APIRouter, HTTPException, Query
from pydantic import BaseModel, Field
from psycopg.rows import dict_row

from app.routers.projects import mirror_object, link_objects

router = APIRouter()

ALLOWED = {"WIP": {"Shared", "Archived"},
           "Shared": {"Published", "WIP", "Archived"},
           "Published": {"Archived"},
           "Archived": set()}


def _pool():
    from app.main import get_pool
    return get_pool()


class DocIn(BaseModel):
    title: str = Field(min_length=3, max_length=255)
    doc_type: str = Field(pattern="^[A-Z]{2,5}$")   # DWG, SPC, MS, ITP, RPT, COR...
    discipline: Optional[str] = None
    originator: Optional[str] = "CAI"
    revision: str = "P01"
    notes: Optional[str] = None


class RevisionIn(BaseModel):
    revision: str = Field(pattern=r"^[PC]\d{2}$")
    suitability: Optional[str] = None
    issued_by: Optional[str] = None
    notes: Optional[str] = None


class StateIn(BaseModel):
    state: str = Field(pattern="^(WIP|Shared|Published|Archived)$")
    suitability: Optional[str] = None


class TransmittalIn(BaseModel):
    to_party: str
    purpose: str = "for_information"
    cover_note: Optional[str] = None
    issued_by: Optional[str] = None
    document_numbers: list[str] = Field(min_length=1)


async def _next_number(conn, project_id, doc_type: str, proj_code: str) -> str:
    rule = await (await conn.execute(
        """INSERT INTO document_numbering_rules (project_id, doc_type, prefix, next_seq)
           VALUES (%s, %s, %s, 1)
           ON CONFLICT (project_id, doc_type)
           DO UPDATE SET next_seq = document_numbering_rules.next_seq + 1
           RETURNING prefix, pad,
                     CASE WHEN xmax = 0 THEN 1
                          ELSE next_seq END AS seq""",
        (project_id, doc_type, f"{proj_code}-CAI-{doc_type}-"),
    )).fetchone()
    return f"{rule['prefix']}{str(rule['seq']).zfill(rule['pad'])}"


@router.get("/{project_id}/documents")
async def list_documents(project_id: uuid.UUID,
                         doc_type: Optional[str] = None,
                         state: Optional[str] = None,
                         q: Optional[str] = None,
                         limit: int = Query(100, le=500), offset: int = 0):
    sql = """SELECT d.*,
                    (SELECT count(*) FROM document_revisions r
                      WHERE r.document_id = d.id) AS revisions_count
             FROM documents d WHERE d.project_id = %s"""
    args: list = [project_id]
    if doc_type:
        sql += " AND d.doc_type = %s"; args.append(doc_type)
    if state:
        sql += " AND d.state = %s"; args.append(state)
    if q:
        sql += " AND (d.doc_number ILIKE %s OR d.title ILIKE %s)"
        args += [f"%{q}%", f"%{q}%"]
    sql += " ORDER BY d.doc_number LIMIT %s OFFSET %s"; args += [limit, offset]
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        return await (await conn.execute(sql, args)).fetchall()


@router.post("/{project_id}/documents", status_code=201)
async def create_document(project_id: uuid.UUID, d: DocIn):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        proj = await (await conn.execute(
            "SELECT code FROM projects WHERE id=%s", (project_id,))).fetchone()
        if not proj:
            raise HTTPException(404, "Project not found")
        number = await _next_number(conn, project_id, d.doc_type, proj["code"])
        doc = await (await conn.execute(
            """INSERT INTO documents (project_id, doc_number, title, doc_type,
                                      discipline, originator, revision, notes)
               VALUES (%s,%s,%s,%s,%s,%s,%s,%s) RETURNING *""",
            (project_id, number, d.title, d.doc_type, d.discipline,
             d.originator, d.revision, d.notes),
        )).fetchone()
        await conn.execute(
            """INSERT INTO document_revisions (document_id, revision, state, notes)
               VALUES (%s,%s,'WIP','initial')""", (doc["id"], d.revision))
        obj = await mirror_object(conn, "cde_document", project_id,
                                  {"code": number, "title": d.title,
                                   "state": "WIP"}, "documents", doc["id"])
        proj_obj = await (await conn.execute(
            "SELECT id FROM objects WHERE source_table='projects' AND source_id=%s",
            (project_id,))).fetchone()
        if proj_obj:
            await link_objects(conn, "belongs_to", obj, proj_obj["id"])
    return doc


@router.post("/{project_id}/documents/{doc_id}/revisions")
async def new_revision(project_id: uuid.UUID, doc_id: uuid.UUID, r: RevisionIn):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        doc = await (await conn.execute(
            "SELECT * FROM documents WHERE id=%s AND project_id=%s",
            (doc_id, project_id))).fetchone()
        if not doc:
            raise HTTPException(404, "Document not found")
        if doc["state"] == "Archived":
            raise HTTPException(409, "Archived documents cannot be revised")

        def rev_key(v):  # C > P, then numeric
            return (v[0] == "C", int(v[1:]))
        if rev_key(r.revision) <= rev_key(doc["revision"]):
            raise HTTPException(409,
                f"Revision {r.revision} must be higher than current {doc['revision']}")
        try:
            await conn.execute(
                """INSERT INTO document_revisions
                     (document_id, revision, state, suitability, issued_by, notes)
                   VALUES (%s,%s,'WIP',%s,%s,%s)""",
                (doc_id, r.revision, r.suitability, r.issued_by, r.notes))
        except Exception:
            raise HTTPException(409, f"Revision {r.revision} already exists")
        upd = await (await conn.execute(
            """UPDATE documents SET revision=%s, state='WIP',
                   suitability=%s, updated_at=NOW()
               WHERE id=%s RETURNING *""",
            (r.revision, r.suitability, doc_id))).fetchone()
    return upd


@router.post("/{project_id}/documents/{doc_id}/state")
async def change_state(project_id: uuid.UUID, doc_id: uuid.UUID, s: StateIn):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        doc = await (await conn.execute(
            "SELECT * FROM documents WHERE id=%s AND project_id=%s",
            (doc_id, project_id))).fetchone()
        if not doc:
            raise HTTPException(404, "Document not found")
        if s.state not in ALLOWED[doc["state"]]:
            raise HTTPException(409,
                f"Illegal transition {doc['state']} -> {s.state}. "
                f"Allowed: {sorted(ALLOWED[doc['state']])}")
        if s.state in ("Shared", "Published") and not (s.suitability or doc["suitability"]):
            raise HTTPException(422,
                "Suitability code (S1..S7/A/B) required to Share or Publish")
        upd = await (await conn.execute(
            """UPDATE documents SET state=%s,
                   suitability=COALESCE(%s, suitability), updated_at=NOW()
               WHERE id=%s RETURNING *""",
            (s.state, s.suitability, doc_id))).fetchone()
        await conn.execute(
            """UPDATE document_revisions SET state=%s,
                   suitability=COALESCE(%s, suitability)
               WHERE document_id=%s AND revision=%s""",
            (s.state, s.suitability, doc_id, doc["revision"]))
    return upd


@router.post("/{project_id}/transmittals", status_code=201)
async def create_transmittal(project_id: uuid.UUID, t: TransmittalIn):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        docs = []
        for num in t.document_numbers:
            d = await (await conn.execute(
                """SELECT id, doc_number, revision, state FROM documents
                   WHERE project_id=%s AND doc_number=%s""",
                (project_id, num))).fetchone()
            if not d:
                raise HTTPException(404, f"Document {num} not found")
            if d["state"] == "WIP":
                raise HTTPException(409,
                    f"{num} is WIP — Share or Publish before transmitting (ISO 19650)")
            docs.append(d)
        trn = await (await conn.execute(
            """INSERT INTO transmittals (project_id, number, code, to_party,
                                         purpose, cover_note, issued_by)
               VALUES (%(p)s,
                       COALESCE((SELECT max(number) FROM transmittals
                                 WHERE project_id=%(p)s),0)+1,
                       'TRN-' || lpad((COALESCE((SELECT max(number) FROM transmittals
                                      WHERE project_id=%(p)s),0)+1)::text,4,'0'),
                       %(to)s, %(pu)s, %(cn)s, %(ib)s)
               RETURNING *""",
            {"p": project_id, "to": t.to_party, "pu": t.purpose,
             "cn": t.cover_note, "ib": t.issued_by})).fetchone()
        for d in docs:
            await conn.execute(
                """INSERT INTO transmittal_items (transmittal_id, document_id, revision)
                   VALUES (%s,%s,%s)""", (trn["id"], d["id"], d["revision"]))
        await mirror_object(conn, "transmittal", project_id,
                            {"code": trn["code"], "to": t.to_party,
                             "docs": len(docs)}, "transmittals", trn["id"])
    return {**trn, "documents": [
        {"doc_number": d["doc_number"], "revision": d["revision"]} for d in docs]}


@router.get("/{project_id}/transmittals")
async def list_transmittals(project_id: uuid.UUID):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        return await (await conn.execute(
            """SELECT t.*,
                      (SELECT json_agg(json_build_object(
                           'doc_number', d.doc_number, 'revision', i.revision))
                       FROM transmittal_items i JOIN documents d ON d.id=i.document_id
                       WHERE i.transmittal_id = t.id) AS items
               FROM transmittals t WHERE t.project_id=%s
               ORDER BY t.number DESC""",
            (project_id,))).fetchall()
