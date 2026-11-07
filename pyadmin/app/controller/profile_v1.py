"""个人中心接口。"""

from __future__ import annotations

from fastapi import APIRouter, Depends

from app.middleware.auth import get_services, require_login
from app.pkg import response
from app.pkg.token import Claims
from app.schema.user import UserUpdateRequest
from app.service.service import Services

router = APIRouter(tags=["profile"])


@router.get("/me", summary="获取当前登录用户信息")
async def me(
    claims: Claims = Depends(require_login()),
    services: Services = Depends(get_services),
):
    """返回用户、组织、角色、权限与按角色分组的权限。"""

    result = await services.profile.me(claims)
    return response.success(result.model_dump())


@router.put("/me", summary="修改个人信息")
async def update_me(
    payload: UserUpdateRequest,
    claims: Claims = Depends(require_login()),
    services: Services = Depends(get_services),
):
    """修改自己的用户名、昵称、手机号、邮箱与密码。"""

    view = await services.profile.update(claims, payload)
    return response.success(view.model_dump())
