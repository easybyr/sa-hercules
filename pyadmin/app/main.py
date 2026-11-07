"""FastAPI 应用装配：生命周期、CORS、统一异常处理与路由。"""

from __future__ import annotations

import logging
from contextlib import asynccontextmanager
from typing import Optional

from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import RedirectResponse
from sqlalchemy.ext.asyncio import AsyncEngine
from starlette.exceptions import HTTPException as StarletteHTTPException

from app.config import Settings, load_settings
from app.controller.router import api_router
from app.db.session import create_engine, create_session_factory
from app.model import Base  # noqa: F401  # 导入即注册 ORM 元数据
from app.pkg import response
from app.pkg.apperror import AppError
from app.pkg.token import TokenManager
from app.pkg.validation import format_validation_errors
from app.repository.store import Store
from app.service.service import Services

logger = logging.getLogger(__name__)

API_TITLE = "Hercules PyAdmin"
API_VERSION = "1.0.0"
API_DESCRIPTION = "Hercules Admin 的 Python3 + FastAPI + PostgreSQL 实现，接口与 admin-web 完全兼容。"


def _configure_logging(level: str) -> None:
    """按配置初始化日志。"""

    logging.basicConfig(
        level=getattr(logging, level.upper(), logging.INFO),
        format="%(asctime)s %(levelname)s %(name)s %(message)s",
    )


def _register_exception_handlers(app: FastAPI) -> None:
    """注册统一异常处理，保证错误也返回 {code, message, data}。"""

    @app.exception_handler(AppError)
    async def _handle_app_error(request: Request, exc: AppError):
        return response.failure(exc.status, exc.message)

    @app.exception_handler(RequestValidationError)
    async def _handle_validation_error(request: Request, exc: RequestValidationError):
        return response.failure(400, format_validation_errors(exc.errors()))

    @app.exception_handler(StarletteHTTPException)
    async def _handle_http_error(request: Request, exc: StarletteHTTPException):
        message = exc.detail if isinstance(exc.detail, str) else "请求处理失败"
        if exc.status_code == 404:
            message = "接口不存在"
        if exc.status_code == 405:
            message = "请求方法不被支持"
        return response.failure(exc.status_code, message)

    @app.exception_handler(Exception)
    async def _handle_unexpected_error(request: Request, exc: Exception):
        logger.exception("未处理异常: %s %s", request.method, request.url.path)
        return response.failure(500, "服务暂时不可用")


def create_app(
    settings: Optional[Settings] = None,
    engine: Optional[AsyncEngine] = None,
) -> FastAPI:
    """创建应用实例；测试可注入 settings 与 engine。"""

    resolved = settings or load_settings()
    _configure_logging("INFO")

    @asynccontextmanager
    async def lifespan(application: FastAPI):
        database_engine = engine or create_engine(resolved.database)
        application.state.engine = database_engine
        session_factory = create_session_factory(database_engine)
        store = Store(session_factory)
        tokens = TokenManager(
            resolved.security.jwt_secret,
            resolved.security.jwt_expire_hours,
        )
        services = Services(store, tokens)
        application.state.services = services
        application.state.settings = resolved

        await services.bootstrap.ensure_super_admin()
        logger.info(
            "PyAdmin 启动完成，环境=%s，配置文件=%s",
            resolved.environment,
            resolved.config_file,
        )
        try:
            yield
        finally:
            await database_engine.dispose()

    app = FastAPI(
        title=API_TITLE,
        version=API_VERSION,
        description=API_DESCRIPTION,
        lifespan=lifespan,
        openapi_url="/api.json",
        docs_url="/docs",
        redoc_url="/redoc",
    )
    app.state.settings = resolved
    app.state.engine_override = engine

    app.add_middleware(
        CORSMiddleware,
        allow_origins=list(resolved.server.cors_origins),
        allow_methods=["GET", "POST", "PUT", "DELETE", "OPTIONS"],
        allow_headers=["Authorization", "Content-Type"],
        allow_credentials=False,
    )

    _register_exception_handlers(app)
    app.include_router(api_router)

    @app.get("/swagger", include_in_schema=False)
    async def swagger_compat():
        """兼容 Go 版本的 /swagger 入口，跳转到 FastAPI 文档页。"""

        return RedirectResponse(url="/docs")

    return app


app = create_app()
