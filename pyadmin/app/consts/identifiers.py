"""业务标识常量。

用户 uid 为全局唯一 10 位数字，组织 org_id 为 8 位数字，角色 role_id 为全局唯一 10 位数字。
"""

from __future__ import annotations

# 超级管理员固定标识，与需求指定值一致。
SUPER_ADMIN_UID: int = 1_000_000_000
SUPER_ADMIN_ORG_ID: int = 10_000_000
SUPER_ADMIN_ROLE_ID: int = 10_000_000

# 超级管理员账号信息，首次启动时幂等创建。
SUPER_ADMIN_USERNAME: str = "admin"
SUPER_ADMIN_PASSWORD: str = "admin"
SUPER_ADMIN_NICKNAME: str = "超级管理员"

# 根组织与系统角色信息。
SUPER_ADMIN_ORGANIZATION_NAME: str = "系统管理组织"
SUPER_ADMIN_ORGANIZATION_DESCRIPTION: str = "超级管理员根组织"
SUPER_ADMIN_ROLE_NAME: str = "超级管理员"
SUPER_ADMIN_ROLE_DESCRIPTION: str = "系统内置超级管理员角色"

# 业务 ID 位数与取值范围。
UID_DIGITS: int = 10
ORG_ID_DIGITS: int = 8
ROLE_ID_DIGITS: int = 10
UID_MAX_VALUE: int = 9_999_999_999
ORG_ID_MAX_VALUE: int = 99_999_999

# 生成唯一业务 ID 时的最大重试次数。
IDENTIFIER_ATTEMPTS: int = 32

# 管理员创建用户时密码留空，生成的随机初始密码长度。
INITIAL_PASSWORD_LENGTH: int = 16
