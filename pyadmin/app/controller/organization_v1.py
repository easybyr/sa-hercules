"""组织管理接口。"""

from __future__ import annotations

from fastapi import APIRouter, Depends, Query

from app.controller.pagination import page_request
from app.middleware.auth import authorize, get_services
from app.pkg import response
from app.pkg.token import Claims
from app.schema.organization import OrganizationCreateRequest, OrganizationUpdateRequest
from app.schema.pagination import PageRequest
from app.service.service import Services

router = APIRouter(tags=["organization"])


@router.get("/organizations", summary="分页查询组织")
async def list_organizations(
    q: str = Query(default="", description="按组织名称、组织 ID 或描述搜索"),
    page: PageRequest = Depends(page_request),
    claims: Claims = Depends(authorize("organization.read")),
    services: Services = Depends(get_services),
):
    """分页查询当前用户可见的组织。"""

    result = await services.organizations.list(claims, q, page)
    return response.success(result.model_dump())


@router.get("/organizations/tree", summary="查询组织树")
async def organization_tree(
    q: str = Query(default=""),
    claims: Claims = Depends(authorize("organization.read")),
    services: Services = Depends(get_services),
):
    """返回完整可见组织树，不分页。"""

    nodes = await services.organizations.tree(claims, q)
    return response.success([node.model_dump() for node in nodes])


@router.post("/organizations", summary="新建组织")
async def create_organization(
    payload: OrganizationCreateRequest,
    claims: Claims = Depends(authorize("organization.create")),
    services: Services = Depends(get_services),
):
    """新建组织，仅超级管理员可调用。"""

    view = await services.organizations.create(claims, payload)
    return response.created(view.model_dump())


@router.put("/organizations/{identifier}", summary="修改组织")
async def update_organization(
    identifier: int,
    payload: OrganizationUpdateRequest,
    claims: Claims = Depends(authorize("organization.update")),
    services: Services = Depends(get_services),
):
    """按主键更新组织。"""

    view = await services.organizations.update(claims, identifier, payload)
    return response.success(view.model_dump())


@router.delete("/organizations/{identifier}", summary="删除组织")
async def delete_organization(
    identifier: int,
    claims: Claims = Depends(authorize("organization.delete")),
    services: Services = Depends(get_services),
):
    """删除空组织，存在子组织、用户、角色或权限时拒绝。"""

    await services.organizations.delete(claims, identifier)
    return response.success(None)
