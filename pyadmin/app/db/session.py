"""异步数据库引擎与会话工厂。"""

from __future__ import annotations

from typing import Any, Dict

from sqlalchemy import text
from sqlalchemy.ext.asyncio import (
    AsyncEngine,
    AsyncSession,
    async_sessionmaker,
    create_async_engine,
)

from app.config import DatabaseSettings

# 业务时区，Postgres 会话时区固定为 Asia/Shanghai，保证接口时间带 +08:00。
BUSINESS_TIMEZONE = "Asia/Shanghai"


def create_engine(settings: DatabaseSettings) -> AsyncEngine:
    """按配置创建异步引擎。"""

    options: Dict[str, Any] = {
        "echo": settings.echo,
        "future": True,
        "pool_pre_ping": True,
    }
    if settings.dsn.startswith("postgresql"):
        options.update(
            pool_size=settings.pool_size,
            max_overflow=settings.max_overflow,
            pool_recycle=settings.pool_recycle_seconds,
            connect_args={"server_settings": {"timezone": BUSINESS_TIMEZONE}},
        )
    return create_async_engine(settings.dsn, **options)


def create_session_factory(engine: AsyncEngine) -> async_sessionmaker[AsyncSession]:
    """创建会话工厂。"""

    return async_sessionmaker(bind=engine, expire_on_commit=False, class_=AsyncSession)


async def ping(engine: AsyncEngine) -> None:
    """执行一次探活查询，失败时抛出异常。"""

    async with engine.connect() as connection:
        await connection.execute(text("SELECT 1"))
