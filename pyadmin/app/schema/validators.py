"""Schema 通用字段校验。"""

from __future__ import annotations

from typing import Optional

from app.consts.patterns import EMAIL_PATTERN


def check_text_length(value: str, minimum: int, maximum: int, label: str) -> str:
    """校验字符串长度（按字符计）。"""

    if len(value) < minimum or len(value) > maximum:
        raise ValueError(
            "{}长度需在{}到{}个字符之间".format(label, minimum, maximum)
        )
    return value


def check_optional_text_length(
    value: Optional[str], maximum: int, label: str
) -> str:
    """校验可选字符串的最大长度。"""

    text = value or ""
    if len(text) > maximum:
        raise ValueError("{}长度不能超过{}个字符".format(label, maximum))
    return text


def check_email(value: Optional[str]) -> str:
    """校验邮箱；留空视为合法。"""

    text = (value or "").strip()
    if not text:
        return ""
    if len(text) > 254:
        raise ValueError("邮箱长度不能超过254个字符")
    if not EMAIL_PATTERN.match(text):
        raise ValueError("邮箱格式不正确")
    return text
