"""API 路由装配：所有接口统一挂在 /api 前缀下。"""

from __future__ import annotations

from fastapi import APIRouter

from app.controller import (
    access_v1,
    auth_v1,
    health_v1,
    organization_v1,
    permission_v1,
    profile_v1,
    role_v1,
    user_v1,
)

api_router = APIRouter(prefix="/api")
api_router.include_router(health_v1.router)
api_router.include_router(auth_v1.router)
api_router.include_router(profile_v1.router)
api_router.include_router(access_v1.router)
api_router.include_router(organization_v1.router)
api_router.include_router(user_v1.router)
api_router.include_router(permission_v1.router)
api_router.include_router(role_v1.router)
