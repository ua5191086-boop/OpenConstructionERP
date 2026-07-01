"""
OpenConstructionERP — core-py: Python reference implementation of the platform core.
Prototype 0.1 scope: minimal ontology API + Projects + BOQ (import/export/summary).

NOTE (ADR-003): the production core is Go. This service exists so the platform is
usable and testable by the current team today; it is the executable specification
for the Go implementation. Do not grow business logic here beyond the BOQ vertical.
"""
import os
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles
from psycopg_pool import AsyncConnectionPool

from app.routers import projects, ontology, boq

DATABASE_URL = os.getenv(
    "DATABASE_URL", "postgresql://oce:oce_dev_only@localhost:5432/oce"
)

pool: AsyncConnectionPool | None = None


def get_pool() -> AsyncConnectionPool:
    assert pool is not None, "DB pool not initialised"
    return pool


@asynccontextmanager
async def lifespan(app: FastAPI):
    global pool
    pool = AsyncConnectionPool(DATABASE_URL, min_size=1, max_size=10, open=False)
    await pool.open()
    yield
    await pool.close()


app = FastAPI(
    title="OpenConstructionERP core-py",
    version="0.1.0",
    lifespan=lifespan,
)

app.include_router(projects.router, prefix="/api/v1/projects", tags=["projects"])
app.include_router(ontology.router, prefix="/api/v1/core", tags=["ontology"])
app.include_router(boq.router, prefix="/api/v1/projects", tags=["boq"])


@app.get("/health")
async def health():
    async with get_pool().connection() as conn:
        row = await (await conn.execute("SELECT 1")).fetchone()
    return {"status": "ok", "db": row[0] == 1, "version": app.version}


# Dashboard: served at /
app.mount(
    "/",
    StaticFiles(directory=os.path.join(os.path.dirname(__file__), "static"), html=True),
    name="static",
)
