"""密码哈希与校验，使用 bcrypt（与 Go 版本的 golang.org/x/crypto/bcrypt 等价）。"""

from __future__ import annotations

import bcrypt

# bcrypt 算法最长只处理 72 字节，超长密码在 Go 版本中也会直接报错。
_MAX_PASSWORD_BYTES: int = 72


def hash_password(raw: str) -> str:
    """生成密码哈希。"""

    encoded = raw.encode("utf-8")
    if len(encoded) > _MAX_PASSWORD_BYTES:
        raise ValueError("密码长度不能超过72字节")
    return bcrypt.hashpw(encoded, bcrypt.gensalt()).decode("utf-8")


def verify_password(hashed: str, raw: str) -> bool:
    """校验密码。哈希非法或密码不匹配时返回 False。"""

    if not hashed:
        return False
    encoded = raw.encode("utf-8")
    if len(encoded) > _MAX_PASSWORD_BYTES:
        return False
    try:
        return bcrypt.checkpw(encoded, hashed.encode("utf-8"))
    except (ValueError, TypeError):
        return False
