"""时间序列化工具。

数据库列使用毫秒精度，接口按 Go 版本输出 RFC3339 毫秒格式，
例如 2026-11-06T23:17:52.123+08:00。
"""

from __future__ import annotations

from datetime import datetime, timezone
from typing import Optional


def to_iso(value: Optional[datetime]) -> Optional[str]:
    """把 datetime 序列化为带时区的 RFC3339 毫秒字符串。"""

    if value is None:
        return None
    moment = value if value.tzinfo is not None else value.replace(tzinfo=timezone.utc)
    return moment.isoformat(timespec="milliseconds")
