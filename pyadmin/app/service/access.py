"""访问控制：权限、组织隔离与登录态校验。"""

from __future__ import annotations

from app.consts.status import Status
from app.pkg import apperror
from app.pkg.token import Claims
from app.repository.store import Store
from app.schema.access import CheckAccessResult


class AccessGuard:
    """集中实现 RBAC 判定，避免各业务服务重复写组织隔离逻辑。"""

    def __init__(self, store: Store) -> None:
        self._store = store

    async def require_permission(self, claims: Claims, code: str) -> None:
        """校验当前组织内的权限；超级管理员直接放行。"""

        if claims.is_super:
            return
        permissions = await self._store.user_permissions(claims.uid, claims.org_id)
        if any(item.code == code for item in permissions):
            return
        raise apperror.forbidden("PERMISSION_DENIED", "没有访问该资源的权限")

    async def validate_claims(self, claims: Claims) -> None:
        """每次请求重新校验用户与组织状态，禁用后立即失效。"""

        user = await self._store.find_user_by_uid(claims.uid)
        if user is None or user.org_id != claims.org_id:
            raise apperror.unauthorized("INVALID_SESSION", "登录状态已失效")
        if int(user.status) != int(Status.NORMAL):
            raise apperror.unauthorized("INVALID_SESSION", "登录状态已失效")
        organization = await self._store.find_organization_by_org_id(user.org_id)
        if organization is None or int(organization.status) != int(Status.NORMAL):
            raise apperror.unauthorized("INVALID_SESSION", "所属组织不可用")

    def require_org(self, claims: Claims, org_id: int) -> None:
        """组织隔离校验：普通用户只能访问本组织数据。"""

        if claims.is_super or claims.org_id == org_id:
            return
        raise apperror.forbidden("ORGANIZATION_DENIED", "不能访问其他组织的数据")

    def list_organization_id(self, claims: Claims, requested_org_id: int) -> int:
        """解析列表组织筛选；普通用户始终只能查询当前组织。"""

        if claims.is_super:
            return requested_org_id
        if requested_org_id and requested_org_id != claims.org_id:
            raise apperror.forbidden("ORGANIZATION_DENIED", "不能访问其他组织的数据")
        return claims.org_id

    async def check_access(
        self,
        claims: Claims,
        permission_code: str,
        role_id: int,
        org_id: int,
    ) -> CheckAccessResult:
        """同时校验权限、角色与组织访问结果。"""

        permission_allowed = not permission_code
        role_allowed = not role_id
        organization_allowed = (
            not org_id or claims.is_super or org_id == claims.org_id
        )

        if claims.is_super:
            permission_allowed = True
            role_allowed = True
        else:
            if permission_code:
                try:
                    await self.require_permission(claims, permission_code)
                    permission_allowed = True
                except apperror.AppError:
                    permission_allowed = False
            if role_id:
                roles = await self._store.user_roles(claims.uid, claims.org_id)
                role_allowed = any(item.role_id == role_id for item in roles)

        return CheckAccessResult(
            permission_allowed=permission_allowed,
            role_allowed=role_allowed,
            organization_allowed=organization_allowed,
            allowed=permission_allowed and role_allowed and organization_allowed,
        )
