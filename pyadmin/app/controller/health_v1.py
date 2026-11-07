"""健康检查接口。"""

from __future__ import annotations

from fastapi import APIRouter, Depends

from app.middleware.auth import get_services
from app.pkg import response
from app.service.service import Services

router = APIRouter(tags=["health"])


@router.get("/health", summary="服务健康检查")
async def health(services: Services = Depends(get_services)):
    """检查服务与数据库连通性。"""

    await services.health.check()
    return response.success({"status": "healthy"})
