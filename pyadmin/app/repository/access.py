"""用户角色、角色权限关联的数据访问。"""

from __future__ import annotations

from typing import List, Sequence

from sqlalchemy import delete, select, update

from app.model import Permission, Role, RolePermission, User, UserRole
from app.pkg.idgen import personal_role_id
from app.repository.store_base import StoreBase


class AccessRepository(StoreBase):
    """RBAC 关联读写。"""

    async def user_roles(self, uid: int, org_id: int) -> List[Role]:
        """查询用户在其组织内的普通角色（隐藏个人直接权限角色）。"""

        statement = (
            select(Role)
            .join(UserRole, UserRole.role_id == Role.role_id)
            .where(
                UserRole.uid == uid,
                Role.org_id == org_id,
                Role.role_id > 0,
            )
            .order_by(Role.role_id.asc())
        )
        async with self.session() as session:
            result = await session.execute(statement)
            return list(result.scalars().all())

    async def user_permissions(self, uid: int, org_id: int) -> List[Permission]:
        """查询用户在当前组织内通过角色获得的权限。"""

        statement = (
            select(Permission)
            .join(RolePermission, RolePermission.permission_code == Permission.code)
            .join(
                Role,
                (Role.role_id == RolePermission.role_id)
                & (Role.org_id == Permission.org_id),
            )
            .join(UserRole, UserRole.role_id == Role.role_id)
            .where(UserRole.uid == uid, Permission.org_id == org_id)
            .distinct()
            .order_by(Permission.code.asc())
        )
        async with self.session() as session:
            result = await session.execute(statement)
            return list(result.scalars().all())

    async def replace_user_roles(self, uid: int, role_ids: Sequence[int]) -> None:
        """原子替换用户普通角色，并清空旧的用户直接权限。"""

        async with self.transaction() as session:
            await session.execute(
                delete(UserRole).where(UserRole.uid == uid, UserRole.role_id > 0)
            )
            for role_id in role_ids:
                session.add(UserRole(uid=uid, role_id=int(role_id), version=1))
            await session.execute(
                delete(RolePermission).where(
                    RolePermission.role_id == personal_role_id(uid)
                )
            )

    async def ensure_user_role(self, uid: int, role_id: int) -> None:
        """确保用户角色关联存在（先查询再插入）。"""

        async with self.transaction() as session:
            existing = await session.scalar(
                select(UserRole.id)
                .where(UserRole.uid == uid, UserRole.role_id == role_id)
                .limit(1)
            )
            if existing:
                return
            session.add(UserRole(uid=uid, role_id=role_id, version=1))

    async def replace_role_permissions(
        self, role_id: int, codes: Sequence[str]
    ) -> None:
        """原子替换角色权限。"""

        async with self.transaction() as session:
            await session.execute(
                delete(RolePermission).where(RolePermission.role_id == role_id)
            )
            for code in codes:
                session.add(
                    RolePermission(
                        role_id=role_id, permission_code=code, version=1
                    )
                )

    async def role_permissions(self, role_id: int) -> List[str]:
        """查询角色权限编码，按编码升序。"""

        statement = (
            select(RolePermission.permission_code)
            .where(RolePermission.role_id == role_id)
            .order_by(RolePermission.permission_code.asc())
        )
        async with self.session() as session:
            values = await session.scalars(statement)
            return [str(item) for item in values.all()]

    async def set_user_organization(self, uid: int, org_id: int) -> None:
        """修改用户组织，并清理旧组织下的角色与个人直接权限。"""

        personal_id = personal_role_id(uid)
        async with self.transaction() as session:
            await session.execute(delete(UserRole).where(UserRole.uid == uid))
            await session.execute(
                delete(RolePermission).where(RolePermission.role_id == personal_id)
            )
            await session.execute(delete(Role).where(Role.role_id == personal_id))
            await session.execute(
                update(User)
                .where(User.uid == uid)
                .values(org_id=org_id, version=User.version + 1)
            )
