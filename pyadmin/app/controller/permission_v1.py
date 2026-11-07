"""权限管理接口。"""

from __future__ import annotations

from fastapi import APIRouter, Depends, Query

from app.controller.pagination import page_request
from app.middleware.auth import authorize, get_services
from app.pkg import response
from app.pkg.token import Claims
from app.schema.pagination import PageRequest
from app.schema.permission import PermissionCreateRequest, PermissionUpdateRequest
from app.service.service import Services

router = APIRouter(tags=["permission"])


@router.get("/permissions", summary="分页查询权限")
async def list_permissions(
    q: str = Query(default="", description="按权限名称或权限编码搜索"),
    org_id: int = Query(default=0, description="组织业务 ID，超级管理员可省略"),
    page: PageRequest = Depends(page_request),
    claims: Claims = Depends(authorize("permission.read")),
    services: Services = Depends(get_services),
):
    """查询当前可管理范围内的权限。"""

    result = await services.permissions.list(claims, org_id, q, page)
    return response.success(result.model_dump())


@router.post("/permissions", summary="新建权限")
async def create_permission(
    payload: PermissionCreateRequest,
    claims: Claims = Depends(authorize("permission.create")),
    services: Services = Depends(get_services),
):
    """新建组织内权限。"""

    view = await services.permissions.create(claims, payload)
    return response.created(view.model_dump())


@router.put("/permissions/{org_id}/{code}", summary="修改权限")
async def update_permission(
    org_id: int,
    code: str,
    payload: PermissionUpdateRequest,
    claims: Claims = Depends(authorize("permission.update")),
    services: Services = Depends(get_services),
):
    """更新权限名称与描述，编码不可修改。"""

    view = await services.permissions.update(claims, org_id, code, payload)
    return response.success(view.model_dump())


@router.delete("/permissions/{org_id}/{code}", summary="删除权限")
async def delete_permission(
    org_id: int,
    code: str,
    claims: Claims = Depends(authorize("permission.delete")),
    services: Services = Depends(get_services),
):
    """删除权限及其在组织内角色中的关联。"""

    await services.permissions.delete(claims, org_id, code)
    return response.success(None)
