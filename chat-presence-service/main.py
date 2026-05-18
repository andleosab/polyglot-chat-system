import asyncio
import logging
import os
from contextlib import asynccontextmanager

import asyncpg
from fastapi import FastAPI

from app import deps
from app.routes.presence_routes import router

logging.basicConfig(level=logging.INFO)
log = logging.getLogger(__name__)


async def _run_migrations(pool: asyncpg.Pool) -> None:
    async with pool.acquire() as conn:
        await conn.execute("""
            CREATE TABLE IF NOT EXISTS user_last_seen (
                useruuid UUID PRIMARY KEY,
                last_seen TIMESTAMPTZ NOT NULL
            )
        """)
    log.info("migrations done")


@asynccontextmanager
async def lifespan(app: FastAPI):
    async for _ in deps.init(app):
        await _run_migrations(deps.get_pool())

        # Start gRPC server in background task
        grpc_port = int(os.environ.get("GRPC_PORT", "50051"))
        from app.grpc_server import serve
        grpc_task = asyncio.create_task(serve(grpc_port))

        yield

        grpc_task.cancel()
        try:
            await grpc_task
        except asyncio.CancelledError:
            pass


app = FastAPI(title="chat-presence-service", lifespan=lifespan)
app.include_router(router)


@app.get("/health")
def health():
    return {"ok": True}
