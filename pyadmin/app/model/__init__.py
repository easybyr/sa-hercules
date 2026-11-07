"""ORM 实体导出，保证导入即完成元数据注册。"""

from __future__ import annotations

from app.model.access import RolePermission, UserRole
from app.model.base import Base
from app.model.organization import Organization
from app.model.permission import Permission
from app.model.role import Role
from app.model.user import User

__all__ = [
    "Base",
    "Organization",
    "Permission",
    "Role",
    "RolePermission",
    "User",
    "UserRole",
]
