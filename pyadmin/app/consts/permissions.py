"""系统内置权限。

首次启动时幂等写入超级管理员组织，并全部授予超级管理员角色。
权限编码使用 `resource.action` 形式的小写点分格式。
"""

from __future__ import annotations

from typing import NamedTuple, Tuple


class PermissionSeed(NamedTuple):
    """内置权限定义。"""

    code: str
    name: str


BUILTIN_PERMISSIONS: Tuple[PermissionSeed, ...] = (
    PermissionSeed("dashboard.read", "查看仪表盘"),
    PermissionSeed("organization.read", "查看组织"),
    PermissionSeed("organization.create", "创建组织"),
    PermissionSeed("organization.update", "修改组织"),
    PermissionSeed("organization.delete", "删除组织"),
    PermissionSeed("user.read", "查看用户"),
    PermissionSeed("user.create", "创建用户"),
    PermissionSeed("user.update", "修改用户"),
    PermissionSeed("user.delete", "删除用户"),
    PermissionSeed("user.assign_role", "分配用户角色"),
    PermissionSeed("user.assign_permission", "分配用户权限"),
    PermissionSeed("user.assign_organization", "分配用户组织"),
    PermissionSeed("permission.read", "查看权限"),
    PermissionSeed("permission.create", "创建权限"),
    PermissionSeed("permission.update", "修改权限"),
    PermissionSeed("permission.delete", "删除权限"),
    PermissionSeed("role.read", "查看角色"),
    PermissionSeed("role.create", "创建角色"),
    PermissionSeed("role.update", "修改角色"),
    PermissionSeed("role.delete", "删除角色"),
    PermissionSeed("role.assign_permission", "分配角色权限"),
)

BUILTIN_PERMISSION_CODES: Tuple[str, ...] = tuple(
    seed.code for seed in BUILTIN_PERMISSIONS
)
