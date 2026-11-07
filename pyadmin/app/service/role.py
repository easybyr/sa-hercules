"""角色业务逻辑。"""

from __future__ import annotations

from typing import List

from app.consts import identifiers as const
from app.model import Role
from app.pkg import apperror
from app.pkg.token import Claims
from app.repository.store import Store
from app.schema.pagination import PageRequest, PageResult
from app.schema.role import RoleCreateRequest, RoleDetailView, RoleUpdateRequest, RoleView
from app.service.access import AccessGuard
from app.service.identifier import IdentifierGenerator
from app.service.permission import validate_permission_codes
from app.service.views import role_view


class RoleService:
    """角色业务逻辑。"""

    def __init__(
        self,
        store: Store,
        guard: AccessGuard,
        identifiers: IdentifierGenerator,
    ) -> None:
        self._store = store
        self._guard = guard
        self._identifiers = identifiers

    async def list(
        self,
        claims: Claims,
        requested_org_id: int,
        keyword: str,
        page: PageRequest,
    ) -> PageResult[RoleView]:
        """查询可见角色，隐藏个人直接权限角色。"""

        org_id = self._guard.list_organization_id(claims, requested_org_id)
        items, total = await self._store.list_roles(
            org_id, False, keyword.strip(), page
        )
        return PageResult[RoleView](
            list=[role_view(item) for item in items],
            total=total,
            page=page.page,
            page_size=page.page_size,
        )

    async def detail(self, claims: Claims, org_id: int, role_id: int) -> RoleDetailView:
        """查询角色及其权限编码。"""

        self._guard.require_org(claims, org_id)
        role = await self._store.find_role(org_id, role_id)
        if role is None or role.role_id <= 0:
            raise apperror.not_found("角色")
        codes = await self._store.role_permissions(role_id)
        return RoleDetailView(role=role_view(role), permission_codes=codes)

    async def create(self, claims: Claims, payload: RoleCreateRequest) -> RoleView:
        """新建角色，可同时保存角色权限。"""

        self._guard.require_org(claims, payload.org_id)
        name = payload.name.strip()
        if not name:
            raise apperror.bad_request("ROLE_NAME_REQUIRED", "角色名称不能为空")

        organization = await self._store.find_organization_by_org_id(payload.org_id)
        if organization is None:
            raise apperror.bad_request("INVALID_ORGANIZATION", "组织不存在")

        codes: List[str] = list(payload.permission_codes or [])
        await validate_permission_codes(self._store, payload.org_id, codes)

        if await self._store.find_role_by_name(payload.org_id, name) is not None:
            raise apperror.conflict(
                "ROLE_NAME_EXISTS", "组织内角色名称已存在，请重新输入角色名称"
            )

        role = Role(
            role_id=await self._identifiers.generate_role_id(),
            name=name,
            description=payload.description,
            org_id=payload.org_id,
            version=1,
        )
        created = await self._store.create_role_with_permissions(role, codes)
        return role_view(created)

    async def update(
        self,
        claims: Claims,
        org_id: int,
        role_id: int,
        payload: RoleUpdateRequest,
    ) -> RoleView:
        """更新角色；未提交权限列表时保留原有权限。"""

        self._guard.require_org(claims, org_id)
        item = await self._store.find_role(org_id, role_id)
        if item is None or role_id <= 0:
            raise apperror.not_found("角色")
        if role_id == const.SUPER_ADMIN_ROLE_ID:
            raise apperror.bad_request(
                "PROTECTED_RESOURCE", "不能修改超级管理员角色"
            )

        name = payload.name.strip()
        if not name:
            raise apperror.bad_request("ROLE_NAME_REQUIRED", "角色名称不能为空")
        existing = await self._store.find_role_by_name(org_id, name)
        if existing is not None and existing.role_id != role_id:
            raise apperror.conflict(
                "ROLE_NAME_EXISTS", "组织内角色名称已存在，请重新输入角色名称"
            )

        if payload.permission_codes is None:
            codes = await self._store.role_permissions(role_id)
        else:
            codes = list(payload.permission_codes)
            await validate_permission_codes(self._store, org_id, codes)

        affected = await self._store.update_role_with_permissions(
            org_id, role_id, name, payload.description, payload.version, codes
        )
        if affected == 0:
            raise apperror.version_conflict()

        item.name = name
        item.description = payload.description
        item.version = payload.version + 1
        return role_view(item)

    async def delete(self, claims: Claims, org_id: int, role_id: int) -> None:
        """删除角色及其关联。"""

        self._guard.require_org(claims, org_id)
        if role_id == const.SUPER_ADMIN_ROLE_ID or role_id <= 0:
            raise apperror.bad_request("PROTECTED_RESOURCE", "不能删除系统角色")
        role = await self._store.find_role(org_id, role_id)
        if role is None:
            raise apperror.not_found("角色")
        await self._store.delete_role(org_id, role_id)

    async def assign_permissions(
        self,
        claims: Claims,
        org_id: int,
        role_id: int,
        codes: List[str],
    ) -> None:
        """替换角色权限。"""

        self._guard.require_org(claims, org_id)
        role = await self._store.find_role(org_id, role_id)
        if role is None or role_id <= 0:
            raise apperror.not_found("角色")
        await validate_permission_codes(self._store, org_id, codes)
        await self._store.replace_role_permissions(role_id, codes)
