import os
from typing import Optional
import redis.asyncio as aioredis
import asyncpg

_redis: Optional[aioredis.Redis] = None
_pool: Optional[asyncpg.Pool] = None


async def init(app):
    global _redis, _pool
    _redis = aioredis.from_url(os.environ["REDIS_URL"], decode_responses=True)
    _pool = await asyncpg.create_pool(os.environ["DATABASE_URL"])

    yield

    await _redis.aclose()
    await _pool.close()


def get_redis() -> aioredis.Redis:
    assert _redis is not None, "Redis not initialised"
    return _redis


def get_pool() -> asyncpg.Pool:
    assert _pool is not None, "DB pool not initialised"
    return _pool
