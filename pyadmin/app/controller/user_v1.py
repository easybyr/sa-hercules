"""用户管理接口。"""

from __future__ import annotations

from fastapi import APIRouter, Depends, Query

from app.controller.pagination import page_request
from app.middleware.auth import authorize, get_services
from app.pkg import response
from app.pkg.token import Claims
from app.schema.organization import OrganizationAssignmentRequest
from app.schema.pagination import PageRequest
from app.schema.user import (
    PermissionCodeListRequest,
    RoleIDListRequest,
    UserCreateRequest,
    UserUpdateRequest,
)
from app.service.service import Services

router = APIRouter(tags=["user"])


@router.get("/users", summary="分页查询用户")
async def list_users(
    q: str = Query(default="", description="按用户名或昵称搜索"),
    org_id: int = Query(default=0, description="组织业务 ID，超级管理员可省略"),
    page: PageRequest = Depends(page_request),
    claims: Claims = Depends(authorize("user.read")),
    services: Services = Depends(get_services),
):
    """查询当前可管理范围内的用户。"""

    result = await services.users.list(claims, org_id, q, page)
    return response.success(result.model_dump())


@router.post("/users", summary="新建用户")
async def create_user(
    payload: UserCreateRequest,
    claims: Claims = Depends(authorize("user.create")),
    services: Services = Depends(get_services),
):
    """管理员创建用户；密码留空时返回一次性初始密码。"""

    result = await services.users.create(claims, payload)
    return response.created(result.model_dump())


@router.get("/users/{uid}", summary="查询用户详情")
async def get_user(
    uid: int,
    claims: Claims = Depends(authorize("user.read")),
    services: Services = Depends(get_services),
):
    """查询用户及其组织、角色、权限。"""

    detail = await services.users.detail(claims, uid)
    return response.success(detail.model_dump())


@router.put("/users/{uid}", summary="修改用户")
async def update_user(
    uid: int,
    payload: UserUpdateRequest,
    claims: Claims = Depends(authorize("user.update")),
    services: Services = Depends(get_services),
):
    """管理员修改用户信息。"""

    view = await services.users.update(claims, uid, payload)
    return response.success(view.model_dump())


@router.delete("/users/{uid}", summary="删除用户")
async def delete_user(
    uid: int,
    claims: Claims = Depends(authorize("user.delete")),
    services: Services = Depends(get_services),
):
    """删除用户及其角色关联。"""

    await services.users.delete(claims, uid)
    return response.success(None)


@router.put("/users/{uid}/roles", summary="分配用户角色")
async def assign_user_roles(
    uid: int,
    payload: RoleIDListRequest,
    claims: Claims = Depends(authorize("user.assign_role")),
    services: Services = Depends(get_services),
):
    """替换用户的普通角色集合。"""

    await services.users.assign_roles(claims, uid, payload.role_ids)
    return response.success(None)


@router.put("/users/{uid}/permissions", summary="分配用户直接权限")
async def assign_user_permissions(
    uid: int,
    payload: PermissionCodeListRequest,
    claims: Claims = Depends(authorize("user.assign_permission")),
    services: Services = Depends(get_services),
):
    """通过隐藏的个人角色分配用户直接权限。"""

    await services.users.assign_permissions(claims, uid, payload.permission_codes)
    return response.success(None)


@router.put("/users/{uid}/organization", summary="分配用户组织")
async def assign_user_organization(
    uid: int,
    payload: OrganizationAssignmentRequest,
    claims: Claims = Depends(authorize("user.assign_organization")),
    services: Services = Depends(get_services),
):
    """跨组织移动用户，仅超级管理员可调用。"""

    await services.users.assign_organization(claims, uid, payload.org_id)
    return response.success(None)
