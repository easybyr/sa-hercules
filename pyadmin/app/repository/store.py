"""数据仓库聚合：把各资源的读写能力组合到同一个 Store 上。"""

from __future__ import annotations

from app.repository.access import AccessRepository
from app.repository.organization import OrganizationRepository
from app.repository.permission import PermissionRepository
from app.repository.role import RoleRepository
from app.repository.user import UserRepository


class Store(
    AccessRepository,
    OrganizationRepository,
    PermissionRepository,
    RoleRepository,
    UserRepository,
):
    """后台管理系统的全部数据库访问入口。"""


__all__ = ["Store"]
