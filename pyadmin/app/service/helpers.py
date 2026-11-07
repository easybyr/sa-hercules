"""服务层通用能力：用户直接权限角色的维护。"""

from __future__ import annotations

from app.model import Role, User
from app.pkg import idgen
from app.repository.store import Store


async def ensure_personal_role(store: Store, user: User) -> int:
    """确保用户的隐藏个人角色存在，并绑定到用户。"""

    role_id = idgen.personal_role_id(user.uid)
    role = await store.find_role(user.org_id, role_id)
    if role is None:
        role = Role(
            role_id=role_id,
            name="_user_{}_direct".format(user.uid),
            description="系统维护的用户直接权限角色",
            org_id=user.org_id,
            version=1,
        )
        await store.create_role(role)
    await store.ensure_user_role(user.uid, role_id)
    return role_id
