"""组织数据访问。"""

from __future__ import annotations

from typing import List, Optional, Tuple

from sqlalchemy import String, delete, func, or_, select, update

from app.model import Organization, Permission, Role, User
from app.repository.store_base import StoreBase
from app.schema.pagination import PageRequest


class OrganizationRepository(StoreBase):
    """组织表读写。"""

    async def find_organization_by_id(self, identifier: int) -> Optional[Organization]:
        """按主键查询组织。"""

        async with self.session() as session:
            result = await session.execute(
                select(Organization).where(Organization.id == identifier).limit(1)
            )
            return result.scalars().first()

    async def find_organization_by_org_id(self, org_id: int) -> Optional[Organization]:
        """按业务 ID 查询组织。"""

        async with self.session() as session:
            result = await session.execute(
                select(Organization).where(Organization.org_id == org_id).limit(1)
            )
            return result.scalars().first()

    async def search_organizations_exact(
        self, keyword: str, limit: int
    ) -> List[Organization]:
        """按完整组织名称或完整组织 ID 精确搜索有效组织。"""

        normalized = keyword.strip().lower()
        conditions = [Organization.status == 1, func.lower(Organization.name) == normalized]
        if keyword.strip().isdigit():
            conditions = [
                Organization.status == 1,
                or_(
                    func.lower(Organization.name) == normalized,
                    Organization.org_id == int(keyword.strip()),
                ),
            ]

        async with self.session() as session:
            result = await session.execute(
                select(Organization)
                .where(*conditions)
                .order_by(Organization.name.asc())
                .limit(limit)
            )
            return list(result.scalars().all())

    async def list_organizations(self) -> List[Organization]:
        """查询全部组织，按主键升序。"""

        async with self.session() as session:
            result = await session.execute(select(Organization).order_by(Organization.id.asc()))
            return list(result.scalars().all())

    async def list_organizations_page(
        self,
        allowed_ids: List[int],
        restricted: bool,
        keyword: str,
        page: PageRequest,
    ) -> Tuple[List[Organization], int]:
        """分页查询组织；restricted 为真时只查询 allowed_ids。"""

        if restricted and not allowed_ids:
            return [], 0

        conditions = []
        if restricted:
            conditions.append(Organization.id.in_(allowed_ids))
        if keyword:
            pattern = "%{}%".format(keyword)
            conditions.append(
                or_(
                    Organization.name.ilike(pattern),
                    func.cast(Organization.org_id, String).ilike(pattern),
                    Organization.description.ilike(pattern),
                )
            )

        statement = select(Organization)
        count_statement = select(func.count()).select_from(Organization)
        if conditions:
            statement = statement.where(*conditions)
            count_statement = count_statement.where(*conditions)
        statement = statement.order_by(
            Organization.updated_at.desc(), Organization.id.desc()
        )

        async with self.session() as session:
            return await self.fetch_page(session, statement, count_statement, page)

    async def create_organization(self, organization: Organization) -> Organization:
        """新建组织。"""

        async with self.transaction() as session:
            session.add(organization)
            await session.flush()
            await session.refresh(organization)
        return organization

    async def update_organization(self, organization: Organization, version: int) -> int:
        """按 version 乐观锁更新组织，返回受影响行数。"""

        async with self.transaction() as session:
            result = await session.execute(
                update(Organization)
                .where(Organization.id == organization.id, Organization.version == version)
                .values(
                    name=organization.name,
                    parent_id=organization.parent_id,
                    description=organization.description,
                    status=organization.status,
                    version=version + 1,
                )
            )
            return int(result.rowcount or 0)

    async def delete_organization(self, identifier: int) -> int:
        """删除组织，返回受影响行数。"""

        async with self.transaction() as session:
            result = await session.execute(
                delete(Organization).where(Organization.id == identifier)
            )
            return int(result.rowcount or 0)

    async def count_children(self, identifier: int) -> int:
        """统计直接子组织数量。"""

        async with self.session() as session:
            total = await session.scalar(
                select(func.count())
                .select_from(Organization)
                .where(Organization.parent_id == identifier)
            )
            return int(total or 0)

    async def count_users(self, org_id: int) -> int:
        """统计组织内用户数量。"""

        async with self.session() as session:
            total = await session.scalar(
                select(func.count()).select_from(User).where(User.org_id == org_id)
            )
            return int(total or 0)

    async def count_permissions(self, org_id: int) -> int:
        """统计组织内权限数量。"""

        async with self.session() as session:
            total = await session.scalar(
                select(func.count())
                .select_from(Permission)
                .where(Permission.org_id == org_id)
            )
            return int(total or 0)

    async def count_roles(self, org_id: int) -> int:
        """统计组织内角色数量。"""

        async with self.session() as session:
            total = await session.scalar(
                select(func.count()).select_from(Role).where(Role.org_id == org_id)
            )
            return int(total or 0)
