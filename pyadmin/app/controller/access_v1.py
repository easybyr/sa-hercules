"""访问校验接口。"""

from __future__ import annotations

from fastapi import APIRouter, Depends, Query

from app.middleware.auth import get_services, require_login
from app.pkg import response
from app.pkg.token import Claims
from app.service.service import Services

router = APIRouter(tags=["access"])


@router.get("/access/check", summary="检查权限、角色与组织访问")
async def check_access(
    permission: str = Query(default="", description="权限编码"),
    role_id: int = Query(default=0, description="角色业务 ID"),
    org_id: int = Query(default=0, description="组织业务 ID"),
    claims: Claims = Depends(require_login()),
    services: Services = Depends(get_services),
):
    """同时校验权限、角色与组织访问结果。"""

    result = await services.guard.check_access(claims, permission, role_id, org_id)
    return response.success(result.model_dump())
