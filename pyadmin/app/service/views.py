"""ORM 实体到接口视图的转换。"""

from __future__ import annotations

from typing import List

from app.model import Organization, Permission, Role, User
from app.schema.organization import OrganizationTreeNode, OrganizationView
from app.schema.permission import PermissionView
from app.schema.role import RoleView
from app.schema.user import UserView


def organization_view(item: Organization) -> OrganizationView:
    """组织实体转视图。"""

    return OrganizationView.model_validate(item)


def organization_tree_node(
    item: Organization,
    children: List[OrganizationTreeNode],
) -> OrganizationTreeNode:
    """组织实体转树节点，children 为空时不输出该字段。"""

    node = OrganizationTreeNode.model_validate(item)
    return node.model_copy(update={"children": children or None})


def user_view(item: User) -> UserView:
    """用户实体转视图。"""

    return UserView.model_validate(item)


def role_view(item: Role) -> RoleView:
    """角色实体转视图。"""

    return RoleView.model_validate(item)


def permission_view(item: Permission) -> PermissionView:
    """权限实体转视图。"""

    return PermissionView.model_validate(item)
