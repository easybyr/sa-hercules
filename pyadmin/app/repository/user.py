"""用户数据访问。"""

from __future__ import annotations

from typing import List, Optional, Tuple

from sqlalchemy import delete, func, or_, select, update

from app.model import Role, RolePermission, User, UserRole
from app.pkg.idgen import personal_role_id
from app.repository.store_base import StoreBase
from app.schema.pagination import PageRequest


class UserRepository(StoreBase):
    """用户表读写。"""

    async def find_user_by_uid(self, uid: int) -> Optional[User]:
        """按 UID 查询用户。"""

        async with self.session() as session:
            result = await session.execute(select(User).where(User.uid == uid).limit(1))
            return result.scalars().first()

    async def find_user_by_username(self, username: str) -> Optional[User]:
        """按用户名查询用户。"""

        async with self.session() as session:
            result = await session.execute(
                select(User).where(User.username == username).limit(1)
            )
            return result.scalars().first()

    async def list_users(
        self,
        org_id: int,
        keyword: str,
        page: PageRequest,
    ) -> Tuple[List[User], int]:
        """分页查询用户；org_id 为 0 时查询全部组织。"""

        conditions = []
        if org_id:
            conditions.append(User.org_id == org_id)
        if keyword:
            pattern = "%{}%".format(keyword)
            conditions.append(
                or_(User.username.ilike(pattern), User.nickname.ilike(pattern))
            )

        statement = select(User)
        count_statement = select(func.count()).select_from(User)
        if conditions:
            statement = statement.where(*conditions)
            count_statement = count_statement.where(*conditions)
        statement = statement.order_by(User.updated_at.desc(), User.id.desc())

        async with self.session() as session:
            return await self.fetch_page(session, statement, count_statement, page)

    async def create_user(self, user: User) -> User:
        """新建用户。"""

        async with self.transaction() as session:
            session.add(user)
            await session.flush()
            await session.refresh(user)
        return user

    async def update_user(self, user: User, version: int) -> int:
        """按 version 乐观锁更新用户，返回受影响行数。"""

        values = {
            "username": user.username,
            "nickname": user.nickname,
            "phone_num": user.phone_num,
            "email": user.email,
            "password": user.password,
            "org_id": user.org_id,
            "status": user.status,
            "version": version + 1,
        }
        async with self.transaction() as session:
            result = await session.execute(
                update(User)
                .where(User.uid == user.uid, User.version == version)
                .values(**values)
            )
            return int(result.rowcount or 0)

    async def delete_user(self, uid: int) -> None:
        """删除用户，并清理角色关联与个人直接权限角色。"""

        personal_id = personal_role_id(uid)
        async with self.transaction() as session:
            await session.execute(delete(UserRole).where(UserRole.uid == uid))
            await session.execute(
                delete(RolePermission).where(RolePermission.role_id == personal_id)
            )
            await session.execute(delete(Role).where(Role.role_id == personal_id))
            await session.execute(delete(User).where(User.uid == uid))
