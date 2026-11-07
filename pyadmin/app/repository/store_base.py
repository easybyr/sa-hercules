"""数据访问基类：会话管理与会话级事务。"""

from __future__ import annotations

from contextlib import asynccontextmanager
from typing import AsyncIterator, List, Tuple

from sqlalchemy import Select, text
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker

from app.schema.pagination import PageRequest


class StoreBase:
    """持有会话工厂，按需开启短会话或短事务。"""

    def __init__(self, session_factory: async_sessionmaker[AsyncSession]) -> None:
        self._session_factory = session_factory

    @asynccontextmanager
    async def session(self) -> AsyncIterator[AsyncSession]:
        """打开会话，只读查询使用。"""

        async with self._session_factory() as session:
            yield session

    @asynccontextmanager
    async def transaction(self) -> AsyncIterator[AsyncSession]:
        """打开事务：正常退出提交，异常回滚。"""

        async with self._session_factory() as session:
            async with session.begin():
                yield session

    async def fetch_page(
        self,
        session: AsyncSession,
        statement: Select,
        count_statement: Select,
        page: PageRequest,
    ) -> Tuple[List, int]:
        """执行分页查询并返回 (记录列表, 总数)。"""

        total = await session.scalar(count_statement)
        rows = await session.execute(
            statement.offset(page.offset).limit(page.page_size)
        )
        return list(rows.scalars().all()), int(total or 0)

    async def ping(self) -> None:
        """探活：执行 SELECT 1，失败时抛出异常。"""

        async with self.session() as session:
            await session.execute(text("SELECT 1"))
