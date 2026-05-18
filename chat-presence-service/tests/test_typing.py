import json
from unittest.mock import AsyncMock
import pytest

from app.typing import publish_typing


@pytest.mark.asyncio
async def test_publish_typing_sends_to_correct_channel():
    redis = AsyncMock()
    redis.publish = AsyncMock()

    await publish_typing("user-uuid-1", "alice", "42", redis)

    redis.publish.assert_awaited_once()
    channel, payload_str = redis.publish.call_args[0]

    assert channel == "conversation:42:typing"
    payload = json.loads(payload_str)
    assert payload["user_uuid"] == "user-uuid-1"
    assert payload["username"] == "alice"
