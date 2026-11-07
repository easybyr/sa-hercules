"""业务 ID 生成：保证 uid / org_id / role_id 唯一。"""

from __future__ import annotations

from app.consts import identifiers as const
from app.pkg import apperror, idgen
from app.repository.store import Store


class IdentifierGenerator:
    """生成并校验唯一业务 ID。"""

    def __init__(self, store: Store) -> None:
        self._store = store

    async def generate_uid(self) -> int:
        """生成 10 位全局唯一 UID。"""

        for _ in range(const.IDENTIFIER_ATTEMPTS):
            candidate = idgen.numeric(const.UID_DIGITS)
            if await self._store.find_user_by_uid(candidate) is None:
                return candidate
        raise apperror.internal_error()

    async def generate_org_id(self) -> int:
        """生成 8 位唯一组织 ID。"""

        for _ in range(const.IDENTIFIER_ATTEMPTS):
            candidate = idgen.numeric(const.ORG_ID_DIGITS)
            if await self._store.find_organization_by_org_id(candidate) is None:
                return candidate
        raise apperror.internal_error()

    async def generate_role_id(self) -> int:
        """生成 10 位全局唯一角色 ID。"""

        for _ in range(const.IDENTIFIER_ATTEMPTS):
            candidate = idgen.numeric(const.ROLE_ID_DIGITS)
            if await self._store.find_role_by_role_id(candidate) is None:
                return candidate
        raise apperror.internal_error()
