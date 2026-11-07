"""启动初始化：幂等创建根组织、超级管理员、系统角色与内置权限。"""

from __future__ import annotations

import logging

from app.consts import identifiers as const
from app.consts.permissions import BUILTIN_PERMISSIONS
from app.consts.status import Status
from app.model import Organization, Permission, Role, User
from app.pkg.password import hash_password
from app.repository.store import Store

logger = logging.getLogger(__name__)


class BootstrapService:
    """首次启动的数据初始化。"""

    def __init__(self, store: Store) -> None:
        self._store = store

    async def ensure_super_admin(self) -> None:
        """幂等初始化超级管理员及其组织、角色、权限。"""

        organization = await self._store.find_organization_by_org_id(
            const.SUPER_ADMIN_ORG_ID
        )
        if organization is None:
            await self._store.create_organization(
                Organization(
                    org_id=const.SUPER_ADMIN_ORG_ID,
                    name=const.SUPER_ADMIN_ORGANIZATION_NAME,
                    description=const.SUPER_ADMIN_ORGANIZATION_DESCRIPTION,
                    status=int(Status.NORMAL),
                    version=1,
                )
            )
            logger.info("已创建超级管理员根组织 org_id=%s", const.SUPER_ADMIN_ORG_ID)

        user = await self._store.find_user_by_uid(const.SUPER_ADMIN_UID)
        if user is None:
            await self._store.create_user(
                User(
                    uid=const.SUPER_ADMIN_UID,
                    username=const.SUPER_ADMIN_USERNAME,
                    nickname=const.SUPER_ADMIN_NICKNAME,
                    password=hash_password(const.SUPER_ADMIN_PASSWORD),
                    org_id=const.SUPER_ADMIN_ORG_ID,
                    status=int(Status.NORMAL),
                    version=1,
                )
            )
            logger.info("已创建超级管理员账号 %s", const.SUPER_ADMIN_USERNAME)

        role = await self._store.find_role(
            const.SUPER_ADMIN_ORG_ID, const.SUPER_ADMIN_ROLE_ID
        )
        if role is None:
            await self._store.create_role(
                Role(
                    role_id=const.SUPER_ADMIN_ROLE_ID,
                    name=const.SUPER_ADMIN_ROLE_NAME,
                    description=const.SUPER_ADMIN_ROLE_DESCRIPTION,
                    org_id=const.SUPER_ADMIN_ORG_ID,
                    version=1,
                )
            )
        await self._store.ensure_user_role(
            const.SUPER_ADMIN_UID, const.SUPER_ADMIN_ROLE_ID
        )

        codes = []
        for seed in BUILTIN_PERMISSIONS:
            permission = await self._store.find_permission(
                const.SUPER_ADMIN_ORG_ID, seed.code
            )
            if permission is None:
                await self._store.create_permission(
                    Permission(
                        code=seed.code,
                        name=seed.name,
                        description="系统内置权限",
                        org_id=const.SUPER_ADMIN_ORG_ID,
                        version=1,
                    )
                )
            codes.append(seed.code)
        await self._store.replace_role_permissions(const.SUPER_ADMIN_ROLE_ID, codes)
