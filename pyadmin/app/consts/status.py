"""资源状态枚举。

数据库存 int，接口统一返回字符串（normal / disabled），与 Go 版本行为一致。
"""

from __future__ import annotations

from enum import IntEnum
from typing import Dict

_STATUS_LABELS: Dict[int, str] = {1: "normal", 2: "disabled"}


class Status(IntEnum):
    """资源状态。"""

    NORMAL = 1  # 正常
    DISABLED = 2  # 禁用

    @classmethod
    def is_valid(cls, value: int) -> bool:
        """校验状态值是否合法。"""

        return int(value) in _STATUS_LABELS

    def label(self) -> str:
        """返回接口输出使用的状态名称。"""

        return _STATUS_LABELS.get(int(self), "unknown")

    @classmethod
    def label_of(cls, value: int) -> str:
        """按原始 int 值返回状态名称，非法值返回 unknown。"""

        return _STATUS_LABELS.get(int(value), "unknown")
