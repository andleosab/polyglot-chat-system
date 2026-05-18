from unittest.mock import AsyncMock, MagicMock, patch
import pytest

from app.presence import set_online, set_offline, get_presence, refresh_ttl


def make_redis(get_return=None):
    r = AsyncMock()
    r.get = AsyncMock(return_value=get_return)
    r.set = AsyncMock()
    r.delete = AsyncMock()
    r.expire = AsyncMock()
    pipe = AsyncMock()
    pipe.delete = MagicMock(return_value=pipe)
    pipe.set = MagicMock(return_value=pipe)
    pipe.execute = AsyncMock()
    r.pipeline = MagicMock(return_value=pipe)
    return r


def make_pool(row=None):
    conn = AsyncMock()
    conn.execute = AsyncMock()
    conn.fetchrow = AsyncMock(return_value=row)
    txn = AsyncMock()
    txn.__aenter__ = AsyncMock(return_value=None)
    txn.__aexit__ = AsyncMock(return_value=False)
    conn.transaction = MagicMock(return_value=txn)
    pool = AsyncMock()
    pool.acquire = MagicMock()
    pool.acquire.return_value.__aenter__ = AsyncMock(return_value=conn)
    pool.acquire.return_value.__aexit__ = AsyncMock(return_value=False)
    return pool, conn


@pytest.mark.asyncio
async def test_set_online_writes_status_with_ttl():
    redis = make_redis()
    await set_online("user-1", redis)
    redis.set.assert_awaited_once_with("user:user-1:status", "online", ex=30)


@pytest.mark.asyncio
async def test_refresh_ttl_extends_expiry():
    redis = make_redis()
    await refresh_ttl("user-1", redis)
    redis.expire.assert_awaited_once_with("user:user-1:status", 30)


@pytest.mark.asyncio
async def test_set_offline_deletes_status_and_writes_last_seen():
    redis = make_redis()
    pool, conn = make_pool()

    await set_offline("user-1", redis, pool)

    conn.execute.assert_awaited_once()
    sql = conn.execute.call_args[0][0]
    assert "user_last_seen" in sql
    assert "user-1" in conn.execute.call_args[0]

    redis.pipeline.assert_called_once()
    pipe = redis.pipeline.return_value
    pipe.delete.assert_called_once()
    pipe.set.assert_called_once()
    pipe.execute.assert_awaited_once()


@pytest.mark.asyncio
async def test_get_presence_online_user():
    redis = make_redis(get_return="online")
    pool, _ = make_pool()

    result = await get_presence("user-1", redis, pool)

    assert result["online"] is True
    assert result["last_seen"] is None


@pytest.mark.asyncio
async def test_get_presence_offline_with_cached_last_seen():
    redis = make_redis()
    redis.get = AsyncMock(side_effect=[None, "2025-01-01T12:00:00+00:00"])
    pool, _ = make_pool()

    result = await get_presence("user-1", redis, pool)

    assert result["online"] is False
    assert result["last_seen"] == "2025-01-01T12:00:00+00:00"


@pytest.mark.asyncio
async def test_get_presence_offline_reads_postgres_on_cache_miss():
    from datetime import datetime, timezone

    ts = datetime(2025, 1, 1, 12, 0, 0, tzinfo=timezone.utc)
    row = {"last_seen": ts}

    redis = make_redis()
    redis.get = AsyncMock(return_value=None)
    pool, conn = make_pool(row=row)

    result = await get_presence("user-1", redis, pool)

    assert result["online"] is False
    assert "2025-01-01" in result["last_seen"]
    redis.set.assert_awaited_once()


@pytest.mark.asyncio
async def test_get_presence_unknown_user():
    redis = make_redis()
    redis.get = AsyncMock(return_value=None)
    pool, _ = make_pool(row=None)

    result = await get_presence("unknown", redis, pool)

    assert result["online"] is False
    assert result["last_seen"] is None
