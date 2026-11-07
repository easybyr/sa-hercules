"""角色数据访问。"""

from __future__ import annotations

from typing import List, Optional, Sequence, Tuple

from sqlalchemy import delete, func, select, update

from app.model import PermissionOperationLog, Role, RolePermission, UserRole
from app.repository.store_base import StoreBase
from app.schema.pagination import PageRequest


class RoleRepository(StoreBase):
    """角色表读写。"""

    async def list_roles(
        self,
        org_id: int,
        include_system: bool,
        keyword: str,
        page: PageRequest,
    ) -> Tuple[List[Role], int]:
        """分页查询角色；org_id 为 0 时查询全部组织。"""

        conditions = []
        if org_id:
            conditions.append(Role.org_id == org_id)
        if not include_system:
            conditions.append(Role.role_id > 0)
        if keyword:
            conditions.append(Role.name.ilike("%{}%".format(keyword)))

        statement = select(Role)
        count_statement = select(func.count()).select_from(Role)
        if conditions:
            statement = statement.where(*conditions)
            count_statement = count_statement.where(*conditions)
        statement = statement.order_by(Role.updated_at.desc(), Role.id.desc())

        async with self.session() as session:
            return await self.fetch_page(session, statement, count_statement, page)

    async def find_role(self, org_id: int, role_id: int) -> Optional[Role]:
        """按组织与角色业务 ID 查询角色。"""

        async with self.session() as session:
            result = await session.execute(
                select(Role)
                .where(Role.org_id == org_id, Role.role_id == role_id)
                .limit(1)
            )
            return result.scalars().first()

    async def find_role_by_role_id(self, role_id: int) -> Optional[Role]:
        """按全局角色业务 ID 查询角色。"""

        async with self.session() as session:
            result = await session.execute(
                select(Role).where(Role.role_id == role_id).limit(1)
            )
            return result.scalars().first()

    async def find_role_by_name(self, org_id: int, name: str) -> Optional[Role]:
        """查询组织内同名角色。"""

        async with self.session() as session:
            result = await session.execute(
                select(Role).where(Role.org_id == org_id, Role.name == name).limit(1)
            )
            return result.scalars().first()

    async def create_role(self, role: Role) -> Role:
        """新建角色。"""

        async with self.transaction() as session:
            session.add(role)
            await session.flush()
            await session.refresh(role)
        return role

    async def create_role_with_permissions(
        self,
        role: Role,
        permission_codes: Sequence[str],
        operation: Optional[dict] = None,
    ) -> Role:
        """原子创建角色及其权限关联。"""

        async with self.transaction() as session:
            session.add(role)
            await session.flush()
            for code in permission_codes:
                session.add(RolePermission(role_id=role.role_id, permission_code=code, version=1))
            await session.flush()
            await session.refresh(role)
            if operation is not None:
                session.add(
                    PermissionOperationLog(
                        **{**operation, "target_id": role.role_id}
                    )
                )
        return role

    async def update_role_with_permissions(
        self,
        org_id: int,
        role_id: int,
        name: str,
        description: str,
        version: int,
        permission_codes: Sequence[str],
        operation: Optional[dict] = None,
    ) -> int:
        """按 version 乐观锁原子更新角色与权限关联，返回受影响行数。"""

        async with self.transaction() as session:
            result = await session.execute(
                update(Role)
                .where(Role.org_id == org_id, Role.role_id == role_id, Role.version == version)
                .values(name=name, description=description, version=version + 1)
            )
            affected = int(result.rowcount or 0)
            if affected == 0:
                return 0
            await session.execute(
                delete(RolePermission).where(RolePermission.role_id == role_id)
            )
            for code in permission_codes:
                session.add(RolePermission(role_id=role_id, permission_code=code, version=1))
            if operation is not None:
                session.add(PermissionOperationLog(**operation))
        return affected

    async def delete_role(
        self, org_id: int, role_id: int, operation: Optional[dict] = None
    ) -> None:
        """删除角色及用户角色、角色权限关联。"""

        async with self.transaction() as session:
            await session.execute(delete(UserRole).where(UserRole.role_id == role_id))
            await session.execute(
                delete(RolePermission).where(RolePermission.role_id == role_id)
            )
            await session.execute(
                delete(Role).where(Role.org_id == org_id, Role.role_id == role_id)
            )
            if operation is not None:
                session.add(PermissionOperationLog(**operation))
