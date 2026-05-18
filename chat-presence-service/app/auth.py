import os
from fastapi import HTTPException, Security
from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
from jose import jwt, JWTError

_bearer = HTTPBearer()

_EXPECTED_AUD = "chat-presence-service"
_EXPECTED_ISS = "chat-web"


def verify_token(
    creds: HTTPAuthorizationCredentials = Security(_bearer),
) -> dict:
    secret = os.environ["JWT_SECRET"]
    try:
        payload = jwt.decode(
            creds.credentials,
            secret,
            algorithms=["HS256"],
            audience=_EXPECTED_AUD,
            issuer=_EXPECTED_ISS,
        )
    except JWTError as exc:
        raise HTTPException(status_code=401, detail=str(exc))
    return payload
