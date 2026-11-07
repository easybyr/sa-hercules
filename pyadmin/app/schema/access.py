"""个人中心、用户详情与访问校验的响应模型。"""

from __future__ import annotations

from typing import List

from app.schema.organization import OrganizationView
from app.schema.pagination import Schema
from app.schema.permission import PermissionView
from app.schema.role import RoleView
from app.schema.user import UserView


class RoleAccessView(Schema):
    """角色及其在当前组织内的权限列表。"""

    role: RoleView
    permissions: List[PermissionView] = []


class MeView(Schema):
    """GET /api/me 响应。"""

    user: UserView
    roles: List[RoleView] = []
    permissions: List[PermissionView] = []
    organization: OrganizationView
    role_accesses: List[RoleAccessView] = []
    is_super: bool = False


class UserDetailView(Schema):
    """GET /api/users/:uid 响应。"""

    user: UserView
    roles: List[RoleView] = []
    permissions: List[PermissionView] = []
    direct_permission_codes: List[str] = []
    organization: OrganizationView


class CheckAccessResult(Schema):
    """GET /api/access/check 响应。"""

    permission_allowed: bool
    role_allowed: bool
    organization_allowed: bool
    allowed: bool
