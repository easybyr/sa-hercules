"""JWT 令牌签发与校验（HS256），声明结构与 Go 版本保持一致。"""

from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from typing import Any, Dict

import jwt

from app.pkg import apperror


@dataclass(frozen=True)
class Claims:
    """访问令牌声明。"""

    uid: int
    org_id: int
    is_super: bool
    username: str


class TokenManager:
    """管理令牌签发与解析。"""

    _ALGORITHM = "HS256"

    def __init__(self, secret: str, expire_hours: int) -> None:
        if not secret:
            raise ValueError("JWT 密钥不能为空")
        self._secret = secret
        self._expire_seconds = max(int(expire_hours), 1) * 3600

    @property
    def expires_in(self) -> int:
        """令牌有效秒数。"""

        return self._expire_seconds

    def issue(self, uid: int, org_id: int, username: str, is_super: bool) -> str:
        """签发访问令牌。"""

        now = datetime.now(tz=timezone.utc)
        payload: Dict[str, Any] = {
            "uid": int(uid),
            "org_id": int(org_id),
            "is_super": bool(is_super),
            "username": username,
            "sub": username,
            "iat": int(now.timestamp()),
            "exp": int((now + timedelta(seconds=self._expire_seconds)).timestamp()),
        }
        return jwt.encode(payload, self._secret, algorithm=self._ALGORITHM)

    def parse(self, raw: str) -> Claims:
        """校验并解析访问令牌，失败时抛出 401 业务错误。"""

        try:
            payload = jwt.decode(
                raw,
                self._secret,
                algorithms=[self._ALGORITHM],
                options={"require": ["exp", "iat"]},
            )
        except jwt.PyJWTError as exc:
            raise apperror.unauthorized(
                "INVALID_TOKEN", "登录凭证无效或已过期"
            ) from exc

        try:
            uid = int(payload["uid"])
            org_id = int(payload["org_id"])
        except (KeyError, TypeError, ValueError) as exc:
            raise apperror.unauthorized(
                "INVALID_TOKEN", "登录凭证无效或已过期"
            ) from exc

        return Claims(
            uid=uid,
            org_id=org_id,
            is_super=bool(payload.get("is_super", False)),
            username=str(payload.get("username") or ""),
        )
