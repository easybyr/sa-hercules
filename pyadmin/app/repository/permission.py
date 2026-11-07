"""权限数据访问。"""

from __future__ import annotations

from typing import List, Optional, Tuple

from sqlalchemy import delete, func, or_, select, update

from app.model import Permission, Role, RolePermission
from app.repository.store_base import StoreBase
from app.schema.pagination import PageRequest


class PermissionRepository(StoreBase):
    """权限表读写。"""

    async def list_permissions(
        self,
        org_id: int,
        keyword: str,
        page: PageRequest,
    ) -> Tuple[List[Permission], int]:
        """分页查询权限；org_id 为 0 时查询全部组织。"""

        conditions = []
        if org_id:
            conditions.append(Permission.org_id == org_id)
        if keyword:
            pattern = "%{}%".format(keyword)
            conditions.append(
                or_(Permission.code.ilike(pattern), Permission.name.ilike(pattern))
            )

        statement = select(Permission)
        count_statement = select(func.count()).select_from(Permission)
        if conditions:
            statement = statement.where(*conditions)
            count_statement = count_statement.where(*conditions)
        statement = statement.order_by(
            Permission.updated_at.desc(), Permission.id.desc()
        )

        async with self.session() as session:
            return await self.fetch_page(session, statement, count_statement, page)

    async def find_permission(self, org_id: int, code: str) -> Optional[Permission]:
        """按组织与权限编码查询权限。"""

        async with self.session() as session:
            result = await session.execute(
                select(Permission)
                .where(Permission.org_id == org_id, Permission.code == code)
                .limit(1)
            )
            return result.scalars().first()

    async def find_permission_by_name(
        self, org_id: int, name: str
    ) -> Optional[Permission]:
        """查询组织内同名权限。"""

        async with self.session() as session:
            result = await session.execute(
                select(Permission)
                .where(Permission.org_id == org_id, Permission.name == name)
                .limit(1)
            )
            return result.scalars().first()

    async def create_permission(self, permission: Permission) -> Permission:
        """新建权限。"""

        async with self.transaction() as session:
            session.add(permission)
            await session.flush()
            await session.refresh(permission)
        return permission

    async def update_permission(
        self,
        org_id: int,
        code: str,
        name: str,
        description: str,
        version: int,
    ) -> int:
        """按 version 乐观锁更新权限，返回受影响行数。"""

        async with self.transaction() as session:
            result = await session.execute(
                update(Permission)
                .where(
                    Permission.org_id == org_id,
                    Permission.code == code,
                    Permission.version == version,
                )
                .values(name=name, description=description, version=version + 1)
            )
            return int(result.rowcount or 0)

    async def delete_permission(self, org_id: int, code: str) -> None:
        """删除权限，并清理组织内各角色对它的关联。"""

        async with self.transaction() as session:
            role_ids = await session.scalars(select(Role.role_id).where(Role.org_id == org_id))
            ids = [int(item) for item in role_ids.all()]
            if ids:
                await session.execute(
                    delete(RolePermission)
                    .where(RolePermission.role_id.in_(ids))
                    .where(RolePermission.permission_code == code)
                )
            await session.execute(
                delete(Permission)
                .where(Permission.org_id == org_id)
                .where(Permission.code == code)
            )
