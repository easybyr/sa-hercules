"""通用正则常量。"""

from __future__ import annotations

import re

# 权限编码：report.read 形式的小写点分格式。
PERMISSION_CODE_PATTERN = re.compile(r"^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$")

# 邮箱：与 GoFrame 的 email 校验同等宽松，仅做基本形状校验。
EMAIL_PATTERN = re.compile(r"^[^@\s]+@[^@\s]+\.[^@\s]+$")
