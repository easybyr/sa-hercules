"""健康检查业务逻辑。"""

from __future__ import annotations

from app.repository.store import Store


class HealthService:
    """服务健康检查。"""

    def __init__(self, store: Store) -> None:
        self._store = store

    async def check(self) -> None:
        """检查数据库连通性，失败时抛出异常。"""

        await self._store.ping()
