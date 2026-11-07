"""业务 ID 与随机密码生成。"""

from __future__ import annotations

import secrets
from typing import Final

_PASSWORD_ALPHABET: Final[str] = (
    "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789!@#$%"
)


def numeric(digits: int) -> int:
    """生成指定长度的正整数，首位不会为 0。"""

    if digits < 1 or digits > 18:
        raise ValueError("数字ID长度必须在1到18之间")
    minimum = 10 ** (digits - 1)
    return minimum + secrets.randbelow(minimum * 9)


def random_password(length: int) -> str:
    """生成一次性初始密码。"""

    if length < 8:
        raise ValueError("随机密码长度不能少于8位")
    return "".join(secrets.choice(_PASSWORD_ALPHABET) for _ in range(length))


def personal_role_id(uid: int) -> int:
    """用户直接权限使用的隐藏个人角色 ID（负数，供系统内部使用）。"""

    return -int(uid)


