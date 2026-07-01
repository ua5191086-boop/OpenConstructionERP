"""Tunnel vertical (L-01 TBM Operations, L-03 Ring Register, L-04 Segments).
Rings are mirrored into the ontology and linked to their drive."""
import uuid
from datetime import datetime
from typing import Optional

from fastapi import APIRouter, HTTPException, Query
from pydantic import BaseModel, Field
from psycopg.rows import dict_row
from psycopg.types.json import Json

from app.routers.projects import mirror_object, link_objects

router = APIRouter()


def _pool():
    from app.main import get_pool
    return get_pool()


class DriveIn(BaseModel):
    code: str
    name: str
    method: str = "TBM"
    chainage_from: float
    chainage_to: float
    ring_width_mm: int = 1500
    design_rings: Optional[int] = None
    tbm_code: Optional[str] = None


class RingIn(BaseModel):
    ring_no: int = Field(ge=1)
    built_at: Optional[datetime] = None
    shift: Optional[str] = None
    advance_mm: Optional[int] = None
    grout_volume_m3: Optional[float] = None
    grout_pressure_bar: Optional[float] = None
    key_position: Optional[int] = Field(default=None, ge=1, le=12)
    attitude: dict = {}
    notes: Optional[str] = None


class RingsBulk(BaseModel):
    """Shift report: several rings in one call."""
    rings: list[RingIn]


@router.get("/{project_id}/tunnel/drives")
async def list_drives(project_id: uuid.UUID):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        return await (await conn.execute(
            """SELECT d.*, t.code AS tbm_code,
                      (SELECT count(*) FROM tunnel_rings r WHERE r.drive_id = d.id) AS rings_built,
                      (SELECT max(r.ring_no) FROM tunnel_rings r WHERE r.drive_id = d.id) AS last_ring,
                      (SELECT max(r.chainage) FROM tunnel_rings r WHERE r.drive_id = d.id) AS current_chainage
               FROM tunnel_drives d
               LEFT JOIN tbm t ON t.id = d.tbm_id
               WHERE d.project_id = %s ORDER BY d.code""",
            (project_id,),
        )).fetchall()


@router.post("/{project_id}/tunnel/drives", status_code=201)
async def create_drive(project_id: uuid.UUID, d: DriveIn):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        tbm_id = None
        if d.tbm_code:
            row = await (await conn.execute(
                """INSERT INTO tbm (project_id, code, diameter_mm)
                   VALUES (%s, %s, 6000)
                   ON CONFLICT (project_id, code) DO UPDATE SET updated_at = NOW()
                   RETURNING id""",
                (project_id, d.tbm_code),
            )).fetchone()
            tbm_id = row["id"]
        length = abs(d.chainage_to - d.chainage_from)
        design_rings = d.design_rings or int(length * 1000 / d.ring_width_mm)
        try:
            drive = await (await conn.execute(
                """INSERT INTO tunnel_drives
                     (project_id, tbm_id, code, name, method, chainage_from,
                      chainage_to, ring_width_mm, design_rings, status)
                   VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,'planned') RETURNING *""",
                (project_id, tbm_id, d.code, d.name, d.method, d.chainage_from,
                 d.chainage_to, d.ring_width_mm, design_rings),
            )).fetchone()
        except Exception as e:
            raise HTTPException(409, f"Cannot create drive: {e}")
        await mirror_object(conn, "tunnel_drive", project_id,
                            {"code": d.code, "name": d.name, "design_rings": design_rings},
                            "tunnel_drives", drive["id"])
    return drive


@router.get("/{project_id}/tunnel/drives/{drive_id}/rings")
async def list_rings(project_id: uuid.UUID, drive_id: uuid.UUID,
                     limit: int = Query(100, le=1000), offset: int = 0):
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        return await (await conn.execute(
            """SELECT r.* FROM tunnel_rings r
               JOIN tunnel_drives d ON d.id = r.drive_id
               WHERE d.project_id = %s AND r.drive_id = %s
               ORDER BY r.ring_no DESC LIMIT %s OFFSET %s""",
            (project_id, drive_id, limit, offset),
        )).fetchall()


@router.post("/{project_id}/tunnel/drives/{drive_id}/rings", status_code=201)
async def add_rings(project_id: uuid.UUID, drive_id: uuid.UUID, payload: RingsBulk):
    """Shift entry: bulk ring registration. Chainage is derived from ring_no."""
    added, errors = 0, []
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        drive = await (await conn.execute(
            """SELECT * FROM tunnel_drives WHERE id = %s AND project_id = %s""",
            (drive_id, project_id),
        )).fetchone()
        if not drive:
            raise HTTPException(404, "Drive not found")
        drive_obj = await mirror_object(conn, "tunnel_drive", project_id,
                                        {"code": drive["code"], "name": drive["name"]},
                                        "tunnel_drives", drive_id)
        direction = 1 if drive["chainage_to"] >= drive["chainage_from"] else -1
        for r in payload.rings:
            chainage = float(drive["chainage_from"]) + direction * (
                r.ring_no * drive["ring_width_mm"] / 1000.0)
            try:
                ring = await (await conn.execute(
                    """INSERT INTO tunnel_rings
                         (drive_id, ring_no, chainage, built_at, shift, advance_mm,
                          grout_volume_m3, grout_pressure_bar, key_position, attitude, notes)
                       VALUES (%s,%s,%s,COALESCE(%s, NOW()),%s,%s,%s,%s,%s,%s,%s)
                       RETURNING id""",
                    (drive_id, r.ring_no, chainage, r.built_at, r.shift, r.advance_mm,
                     r.grout_volume_m3, r.grout_pressure_bar, r.key_position,
                     Json(r.attitude), r.notes),
                )).fetchone()
                ring_obj = await mirror_object(
                    conn, "ring", project_id,
                    {"code": f"{drive['code']}/R{r.ring_no}", "ring_no": r.ring_no},
                    "tunnel_rings", ring["id"])
                await link_objects(conn, "belongs_to", ring_obj, drive_obj)
                added += 1
            except Exception as e:
                errors.append(f"ring {r.ring_no}: {str(e)[:80]}")
        if added:
            await conn.execute(
                """UPDATE tunnel_drives SET status = 'boring',
                       started_at = COALESCE(started_at, CURRENT_DATE),
                       updated_at = NOW()
                   WHERE id = %s AND status = 'planned'""",
                (drive_id,))
    return {"added": added, "errors": errors}


@router.get("/{project_id}/tunnel/drives/{drive_id}/progress")
async def drive_progress(project_id: uuid.UUID, drive_id: uuid.UUID):
    """Rings per day + cumulative S-curve vs design, avg rate, forecast to breakthrough."""
    async with _pool().connection() as conn:
        conn.row_factory = dict_row
        drive = await (await conn.execute(
            "SELECT * FROM tunnel_drives WHERE id = %s AND project_id = %s",
            (drive_id, project_id),
        )).fetchone()
        if not drive:
            raise HTTPException(404, "Drive not found")
        daily = await (await conn.execute(
            """SELECT built_at::date AS day, count(*) AS rings,
                      sum(count(*)) OVER (ORDER BY built_at::date) AS cumulative
               FROM tunnel_rings WHERE drive_id = %s
               GROUP BY 1 ORDER BY 1""",
            (drive_id,),
        )).fetchall()
        built = int(daily[-1]["cumulative"]) if daily else 0
        days = len(daily)
        rate = round(built / days, 2) if days else 0.0
        remaining = max((drive["design_rings"] or 0) - built, 0)
        eta_days = round(remaining / rate, 1) if rate else None
    return {
        "drive": drive["code"], "design_rings": drive["design_rings"],
        "rings_built": built, "percent": round(100 * built / drive["design_rings"], 1)
                        if drive["design_rings"] else None,
        "working_days": days, "avg_rings_per_day": rate,
        "remaining_rings": remaining, "eta_working_days": eta_days,
        "daily": [{"day": str(d["day"]), "rings": d["rings"],
                   "cumulative": int(d["cumulative"])} for d in daily],
    }
