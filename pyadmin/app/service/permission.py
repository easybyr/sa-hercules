"""权限业务逻辑。"""

from __future__ import annotations

from typing import Sequence

from app.consts.patterns import PERMISSION_CODE_PATTERN
from app.model import Permission
from app.pkg import apperror
from app.pkg.token import Claims
from app.repository.store import Store
from app.schema.pagination import PageRequest, PageResult
from app.schema.permission import (
    PermissionCreateRequest,
    PermissionUpdateRequest,
    PermissionView,
)
from app.service.access import AccessGuard
from app.service.helpers import permission_operation_log, permission_snapshot
from app.service.views import permission_view


async def validate_permission_codes(
    store: Store,
    org_id: int,
    codes: Sequence[str],
) -> None:
    """校验权限编码列表：不允许空值、重复值或不属于当前组织的编码。"""

    seen = set()
    for code in codes:
        if not code or code in seen:
            raise apperror.bad_request(
                "INVALID_PERMISSION", "权限列表包含空值或重复值"
            )
        seen.add(code)
        permission = await store.find_permission(org_id, code)
        if permission is None:
            raise apperror.bad_request(
                "INVALID_PERMISSION", "权限不存在或不属于当前组织"
            )


class PermissionService:
    """权限业务逻辑。"""

    def __init__(self, store: Store, guard: AccessGuard) -> None:
        self._store = store
        self._guard = guard

    async def list(
        self,
        claims: Claims,
        requested_org_id: int,
        keyword: str,
        page: PageRequest,
    ) -> PageResult[PermissionView]:
        """分页查询权限。"""

        org_id = self._guard.list_organization_id(claims, requested_org_id)
        items, total = await self._store.list_permissions(org_id, keyword.strip(), page)
        return PageResult[PermissionView](
            list=[permission_view(item) for item in items],
            total=total,
            page=page.page,
            page_size=page.page_size,
        )

    async def create(
        self,
        claims: Claims,
        payload: PermissionCreateRequest,
    ) -> PermissionView:
        """新建组织内权限。"""

        self._guard.require_org(claims, payload.org_id)
        code = payload.code.strip()
        name = payload.name.strip()
        if not PERMISSION_CODE_PATTERN.match(code):
            raise apperror.bad_request(
                "INVALID_PERMISSION_CODE",
                "权限编码格式无效，请使用类似 report.read 的小写点分格式",
            )
        if not name:
            raise apperror.bad_request(
                "PERMISSION_NAME_REQUIRED", "权限名称不能为空"
            )

        organization = await self._store.find_organization_by_org_id(payload.org_id)
        if organization is None:
            raise apperror.bad_request("INVALID_ORGANIZATION", "组织不存在")
        if await self._store.find_permission(payload.org_id, code) is not None:
            raise apperror.conflict("PERMISSION_EXISTS", "组织内权限编码已存在")
        if await self._store.find_permission_by_name(payload.org_id, name) is not None:
            raise apperror.conflict(
                "PERMISSION_NAME_EXISTS", "组织内权限名称已存在，请重新输入权限名称"
            )

        item = Permission(
            code=code,
            name=name,
            description=payload.description,
            org_id=payload.org_id,
            version=1,
        )
        operation = permission_operation_log(
            claims.uid,
            payload.org_id,
            "permission",
            0,
            "permission.create",
            None,
            permission_snapshot(item),
        )
        created = await self._store.create_permission(item, operation)
        return permission_view(created)

    async def update(
        self,
        claims: Claims,
        org_id: int,
        code: str,
        payload: PermissionUpdateRequest,
    ) -> PermissionView:
        """更新权限：编码不可修改。"""

        self._guard.require_org(claims, org_id)
        item = await self._store.find_permission(org_id, code)
        if item is None:
            raise apperror.not_found("权限")

        name = payload.name.strip()
        if not name:
            raise apperror.bad_request(
                "PERMISSION_NAME_REQUIRED", "权限名称不能为空"
            )
        existing = await self._store.find_permission_by_name(org_id, name)
        if existing is not None and existing.code != code:
            raise apperror.conflict(
                "PERMISSION_NAME_EXISTS", "组织内权限名称已存在，请重新输入权限名称"
            )

        before = permission_snapshot(item)
        after = {
            **before,
            "name": name,
            "description": payload.description,
            "version": payload.version + 1,
        }
        operation = permission_operation_log(
            claims.uid,
            org_id,
            "permission",
            item.id,
            "permission.update",
            before,
            after,
        )
        affected = await self._store.update_permission(
            org_id, code, name, payload.description, payload.version, operation
        )
        if affected == 0:
            raise apperror.version_conflict()

        item.name = name
        item.description = payload.description
        item.version = payload.version + 1
        return permission_view(item)

    async def delete(self, claims: Claims, org_id: int, code: str) -> None:
        """删除权限及其在组织内角色中的关联。"""

        self._guard.require_org(claims, org_id)
        item = await self._store.find_permission(org_id, code)
        if item is None:
            raise apperror.not_found("权限")
        operation = permission_operation_log(
            claims.uid,
            org_id,
            "permission",
            item.id,
            "permission.delete",
            permission_snapshot(item),
            None,
        )
        await self._store.delete_permission(org_id, code, operation)
