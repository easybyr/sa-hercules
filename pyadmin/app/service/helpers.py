"""服务层通用能力：用户直接权限角色的维护。"""

from __future__ import annotations

from typing import Any, Dict

from app.model import Role, User
from app.pkg import idgen
from app.repository.store import Store


def permission_operation_log(
    operator_uid: int,
    org_id: int,
    target_type: str,
    target_id: int,
    operation: str,
    before_data: Any,
    after_data: Any,
) -> Dict[str, Any]:
    """Build one audit record; repositories persist it with the RBAC change."""

    return {
        "org_id": org_id,
        "operator_uid": operator_uid,
        "target_type": target_type,
        "target_id": target_id,
        "operation": operation,
        "before_data": before_data if before_data is not None else {"exists": False},
        "after_data": after_data if after_data is not None else {"exists": False},
    }


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


def permission_snapshot(item: object) -> Dict[str, Any]:
    """Return stable permission fields for an audit record."""

    return {
        "id": getattr(item, "id"),
        "code": getattr(item, "code"),
        "name": getattr(item, "name"),
        "description": getattr(item, "description"),
        "org_id": getattr(item, "org_id"),
        "version": getattr(item, "version"),
    }


def role_snapshot(item: object) -> Dict[str, Any]:
    """Return stable role fields for an audit record."""

    return {
        "id": getattr(item, "id"),
        "role_id": getattr(item, "role_id"),
        "name": getattr(item, "name"),
        "description": getattr(item, "description"),
        "org_id": getattr(item, "org_id"),
        "version": getattr(item, "version"),
    }
