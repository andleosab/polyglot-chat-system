from datetime import datetime, timezone
from typing import Optional

import redis.asyncio as aioredis
import asyncpg

_STATUS_TTL = 30  # seconds — must be longer than heartbeat interval (~20s)
_STATUS_KEY = "user:{}:status"
_LAST_SEEN_KEY = "user:{}:last_seen"


async def set_online(user_uuid: str, redis: aioredis.Redis) -> None:
    await redis.set(_STATUS_KEY.format(user_uuid), "online", ex=_STATUS_TTL)


async def set_offline(
    user_uuid: str, redis: aioredis.Redis, pool: asyncpg.Pool
) -> None:
    now = datetime.now(timezone.utc)
    iso = now.isoformat()

    async with pool.acquire() as conn:
        async with conn.transaction():
            await conn.execute(
                """
                INSERT INTO user_last_seen (useruuid, last_seen)
                VALUES ($1, $2)
                ON CONFLICT (useruuid) DO UPDATE SET last_seen = EXCLUDED.last_seen
                """,
                user_uuid,
                now,
            )
            # update Redis cache after the DB write succeeds
            pipe = redis.pipeline()
            pipe.delete(_STATUS_KEY.format(user_uuid))
            pipe.set(_LAST_SEEN_KEY.format(user_uuid), iso)
            await pipe.execute()


async def refresh_ttl(user_uuid: str, redis: aioredis.Redis) -> None:
    await redis.expire(_STATUS_KEY.format(user_uuid), _STATUS_TTL)


async def get_presence(
    user_uuid: str, redis: aioredis.Redis, pool: asyncpg.Pool
) -> dict:
    status = await redis.get(_STATUS_KEY.format(user_uuid))
    online = status == "online"

    last_seen: Optional[str] = None
    if not online:
        cached = await redis.get(_LAST_SEEN_KEY.format(user_uuid))
        if cached:
            last_seen = cached
        else:
            async with pool.acquire() as conn:
                row = await conn.fetchrow(
                    "SELECT last_seen FROM user_last_seen WHERE useruuid = $1",
                    user_uuid,
                )
            if row:
                last_seen = row["last_seen"].isoformat()
                # warm the cache
                await redis.set(_LAST_SEEN_KEY.format(user_uuid), last_seen)

    return {"online": online, "last_seen": last_seen}
