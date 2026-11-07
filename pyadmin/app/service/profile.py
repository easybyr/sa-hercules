"""个人中心业务逻辑。"""

from __future__ import annotations

from app.pkg import apperror
from app.pkg.token import Claims
from app.repository.store import Store
from app.schema.access import MeView, RoleAccessView
from app.schema.user import UserUpdateRequest, UserView
from app.service.access import AccessGuard
from app.service.user import UserService
from app.service.views import (
    organization_view,
    permission_view,
    role_view,
    user_view,
)


class ProfileService:
    """当前用户信息与资料维护。"""

    def __init__(self, store: Store, guard: AccessGuard, users: UserService) -> None:
        self._store = store
        self._guard = guard
        self._users = users

    async def me(self, claims: Claims) -> MeView:
        """获取当前用户完整信息，包含按角色分组的权限。"""

        user = await self._store.find_user_by_uid(claims.uid)
        if user is None:
            raise apperror.AppError(404, "USER_NOT_FOUND", "用户不存在")

        roles = await self._store.user_roles(claims.uid, user.org_id)
        permissions = await self._store.user_permissions(claims.uid, user.org_id)
        permissions_by_code = {item.code: item for item in permissions}

        role_accesses = []
        for role in roles:
            codes = await self._store.role_permissions(role.role_id)
            granted = [
                permission_view(permissions_by_code[code])
                for code in codes
                if code in permissions_by_code
            ]
            role_accesses.append(
                RoleAccessView(role=role_view(role), permissions=granted)
            )

        organization = await self._store.find_organization_by_org_id(user.org_id)
        return MeView(
            user=user_view(user),
            roles=[role_view(item) for item in roles],
            permissions=[permission_view(item) for item in permissions],
            organization=organization_view(organization),
            role_accesses=role_accesses,
            is_super=claims.is_super,
        )

    async def update(self, claims: Claims, payload: UserUpdateRequest) -> UserView:
        """更新当前用户资料，不允许自助修改状态。"""

        user = await self._store.find_user_by_uid(claims.uid)
        if user is None:
            raise apperror.AppError(404, "USER_NOT_FOUND", "用户不存在")
        return await self._users.update_record(user, payload, allow_status=False)
