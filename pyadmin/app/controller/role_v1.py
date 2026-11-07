"""角色管理接口。"""

from __future__ import annotations

from fastapi import APIRouter, Depends, Query

from app.controller.pagination import page_request
from app.middleware.auth import authorize, get_services
from app.pkg import response
from app.pkg.token import Claims
from app.schema.pagination import PageRequest
from app.schema.role import RoleCreateRequest, RoleUpdateRequest
from app.schema.user import PermissionCodeListRequest
from app.service.service import Services

router = APIRouter(tags=["role"])


@router.get("/roles", summary="分页查询角色")
async def list_roles(
    q: str = Query(default="", description="按角色名称搜索"),
    org_id: int = Query(default=0, description="组织业务 ID，超级管理员可省略"),
    page: PageRequest = Depends(page_request),
    claims: Claims = Depends(authorize("role.read")),
    services: Services = Depends(get_services),
):
    """查询可见角色，隐藏个人直接权限角色。"""

    result = await services.roles.list(claims, org_id, q, page)
    return response.success(result.model_dump())


@router.post("/roles", summary="新建角色")
async def create_role(
    payload: RoleCreateRequest,
    claims: Claims = Depends(authorize("role.create")),
    services: Services = Depends(get_services),
):
    """新建角色，可同时提交角色权限。"""

    view = await services.roles.create(claims, payload)
    return response.created(view.model_dump())


@router.get("/roles/{org_id}/{role_id}", summary="查询角色详情")
async def get_role(
    org_id: int,
    role_id: int,
    claims: Claims = Depends(authorize("role.read")),
    services: Services = Depends(get_services),
):
    """查询角色及其权限编码。"""

    detail = await services.roles.detail(claims, org_id, role_id)
    return response.success(detail.model_dump())


@router.put("/roles/{org_id}/{role_id}", summary="修改角色")
async def update_role(
    org_id: int,
    role_id: int,
    payload: RoleUpdateRequest,
    claims: Claims = Depends(authorize("role.update")),
    services: Services = Depends(get_services),
):
    """更新角色；未提交权限列表时保留原权限。"""

    view = await services.roles.update(claims, org_id, role_id, payload)
    return response.success(view.model_dump())


@router.delete("/roles/{org_id}/{role_id}", summary="删除角色")
async def delete_role(
    org_id: int,
    role_id: int,
    claims: Claims = Depends(authorize("role.delete")),
    services: Services = Depends(get_services),
):
    """删除角色及其关联。"""

    await services.roles.delete(claims, org_id, role_id)
    return response.success(None)


@router.put("/roles/{org_id}/{role_id}/permissions", summary="分配角色权限")
async def assign_role_permissions(
    org_id: int,
    role_id: int,
    payload: PermissionCodeListRequest,
    claims: Claims = Depends(authorize("role.assign_permission")),
    services: Services = Depends(get_services),
):
    """替换角色权限。"""

    await services.roles.assign_permissions(
        claims, org_id, role_id, payload.permission_codes
    )
    return response.success(None)
