"""登录与注册接口。"""

from __future__ import annotations

from fastapi import APIRouter, Depends, Query

from app.middleware.auth import get_services
from app.pkg import response
from app.schema.auth import LoginRequest, RegisterRequest
from app.service.service import Services

router = APIRouter(tags=["auth"])


@router.post("/auth/login", summary="用户登录")
async def login(
    payload: LoginRequest,
    services: Services = Depends(get_services),
):
    """校验用户名密码并返回 JWT。"""

    result = await services.auth.login(payload)
    return response.success(result.model_dump())


@router.post("/auth/register", summary="用户注册")
async def register(
    payload: RegisterRequest,
    services: Services = Depends(get_services),
):
    """在有效组织内注册新用户。"""

    user = await services.auth.register(payload)
    return response.created(user.model_dump())


@router.get("/public/organizations/search", summary="注册页精确搜索组织")
async def search_organizations(
    q: str = Query(default=""),
    services: Services = Depends(get_services),
):
    """按完整组织名称或完整组织 ID 精确搜索有效组织。"""

    items = await services.organizations.search(q)
    return response.success([item.model_dump() for item in items])
