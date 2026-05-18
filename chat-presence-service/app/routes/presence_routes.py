from fastapi import APIRouter, Depends, HTTPException
from pydantic import BaseModel

from app import presence as presence_svc
from app import typing as typing_svc
from app.auth import verify_token
from app.deps import get_redis, get_pool

router = APIRouter()


class TypingRequest(BaseModel):
    conversationId: str


@router.get("/presence/{user_uuid}")
async def get_presence(
    user_uuid: str,
    _: dict = Depends(verify_token),
):
    return await presence_svc.get_presence(user_uuid, get_redis(), get_pool())


@router.post("/presence/typing", status_code=200)
async def post_typing(
    body: TypingRequest,
    token: dict = Depends(verify_token),
):
    user_uuid: str = token.get("sub", "")
    username: str = token.get("username", user_uuid)
    if not user_uuid:
        raise HTTPException(status_code=400, detail="sub claim missing from token")
    await typing_svc.publish_typing(user_uuid, username, body.conversationId, get_redis())
    return {"ok": True}
