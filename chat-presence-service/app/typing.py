import json
import redis.asyncio as aioredis

_TYPING_CHANNEL = "conversation:{}:typing"


async def publish_typing(
    user_uuid: str, username: str, conversation_id: str, redis: aioredis.Redis
) -> None:
    channel = _TYPING_CHANNEL.format(conversation_id)
    payload = json.dumps({"user_uuid": user_uuid, "username": username})
    await redis.publish(channel, payload)
