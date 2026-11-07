"""请求级认证：解析 Bearer Token、校验登录态与入口权限。"""

from __future__ import annotations

from typing import Callable

from fastapi import Depends, Request

from app.pkg import apperror
from app.pkg.token import Claims
from app.service.service import Services


def get_services(request: Request) -> Services:
    """从应用状态中取出服务容器。"""

    return request.app.state.services


async def current_claims(
    request: Request,
    services: Services = Depends(get_services),
) -> Claims:
    """校验 Authorization 头并返回令牌声明。"""

    header = request.headers.get("authorization") or ""
    parts = header.split(" ", 1)
    if len(parts) != 2 or parts[0].lower() != "bearer" or not parts[1].strip():
        raise apperror.unauthorized("UNAUTHORIZED", "请先登录")

    claims = services.tokens.parse(parts[1].strip())
    await services.guard.validate_claims(claims)
    return claims


def authorize(permission: str) -> Callable:
    """构造入口权限校验依赖。"""

    async def dependency(
        claims: Claims = Depends(current_claims),
        services: Services = Depends(get_services),
    ) -> Claims:
        await services.guard.require_permission(claims, permission)
        return claims

    return dependency


def require_login() -> Callable:
    """仅要求登录，不校验具体权限。"""

    async def dependency(claims: Claims = Depends(current_claims)) -> Claims:
        return claims

    return dependency
