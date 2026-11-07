"""测试夹具：为每个用例准备独立 SQLite 库并启动完整应用。"""

from __future__ import annotations

import asyncio
from pathlib import Path
from typing import Iterator

import pytest
from fastapi.testclient import TestClient

from app.config import (
    DatabaseSettings,
    SecuritySettings,
    ServerSettings,
    Settings,
)
from app.db.session import create_engine
from app.main import create_app
from app.model import Base

SUPER_ADMIN_USERNAME = "admin"
SUPER_ADMIN_PASSWORD = "admin"


def build_settings(dsn: str) -> Settings:
    """构造测试用配置。"""

    return Settings(
        environment="test",
        server=ServerSettings(host="127.0.0.1", port=8000, cors_origins=("*",)),
        database=DatabaseSettings(
            dsn=dsn,
            echo=False,
            pool_size=2,
            max_overflow=2,
            pool_recycle_seconds=1800,
        ),
        security=SecuritySettings(jwt_secret="test-secret-key", jwt_expire_hours=12),
        config_file=Path("config/config.test.yaml"),
    )


@pytest.fixture()
def client(tmp_path) -> Iterator[TestClient]:
    """提供已完成建表与初始化超级管理员的应用客户端。"""

    settings = build_settings("sqlite+aiosqlite:///{}".format(tmp_path / "pyadmin.db"))
    engine = create_engine(settings.database)

    async def prepare() -> None:
        async with engine.begin() as connection:
            await connection.run_sync(Base.metadata.create_all)
        await engine.dispose()

    asyncio.run(prepare())

    with TestClient(create_app(settings)) as test_client:
        yield test_client


def login(client: TestClient, username: str, password: str) -> str:
    """登录并返回 token。"""

    result = client.post(
        "/api/auth/login", json={"username": username, "password": password}
    )
    assert result.status_code == 200, result.text
    return result.json()["data"]["token"]


def auth_headers(token: str) -> dict:
    """构造 Bearer 认证头。"""

    return {"Authorization": "Bearer {}".format(token)}


@pytest.fixture()
def admin_headers(client: TestClient) -> dict:
    """超级管理员认证头。"""

    return auth_headers(login(client, SUPER_ADMIN_USERNAME, SUPER_ADMIN_PASSWORD))
