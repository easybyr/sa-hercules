"""PostgreSQL 真实环境集成测试。

默认跳过；设置 PYADMIN_TEST_PG_DSN 指向一个**专用测试库**后执行，例如：

    PYADMIN_TEST_PG_DSN=postgresql+asyncpg://hercules:hercules@127.0.0.1:5432/hercules_admin_test \
        python -m pytest tests/test_postgres_integration.py

为安全起见，库名必须包含 "test"，脚本会先删除 pyadmin 的 6 张表再重新执行迁移脚本。
"""

from __future__ import annotations

import asyncio
import os
from pathlib import Path

import asyncpg
import pytest
from fastapi.testclient import TestClient

from app.config import (
    DatabaseSettings,
    SecuritySettings,
    ServerSettings,
    Settings,
)
from app.main import create_app

PG_DSN = os.getenv("PYADMIN_TEST_PG_DSN", "").strip()
MIGRATION_FILE = Path(__file__).resolve().parents[1] / "resources" / "migrations" / "001_init.sql"
TABLES = (
    "role_permission",
    "user_role",
    "role",
    "permission",
    "admin_user",
    "organization",
)

pytestmark = pytest.mark.skipif(
    not PG_DSN, reason="未设置 PYADMIN_TEST_PG_DSN，跳过真实 PostgreSQL 集成测试"
)


def _asyncpg_dsn() -> str:
    return PG_DSN.replace("postgresql+asyncpg://", "postgresql://").replace(
        "postgres://", "postgresql://"
    )


def _settings() -> Settings:
    return Settings(
        environment="test",
        server=ServerSettings(host="127.0.0.1", port=8000, cors_origins=("*",)),
        database=DatabaseSettings(
            dsn=PG_DSN, echo=False, pool_size=2, max_overflow=2, pool_recycle_seconds=1800
        ),
        security=SecuritySettings(jwt_secret="pg-test-secret", jwt_expire_hours=12),
        config_file=Path("config/config.test.yaml"),
    )


async def _reset_database() -> None:
    connection = await asyncpg.connect(_asyncpg_dsn())
    try:
        database = await connection.fetchval("SELECT current_database()")
        if "test" not in str(database).lower():
            pytest.skip("测试库名必须包含 test，当前为 {}".format(database))
        for table in TABLES:
            await connection.execute("DROP TABLE IF EXISTS {} CASCADE".format(table))
        await connection.execute(MIGRATION_FILE.read_text(encoding="utf-8"))
    finally:
        await connection.close()


@pytest.fixture()
def pg_client():
    """重置测试库并启动应用。"""

    asyncio.run(_reset_database())
    with TestClient(create_app(_settings())) as client:
        yield client


def test_schema_objects_are_created(pg_client):
    async def inspect():
        connection = await asyncpg.connect(_asyncpg_dsn())
        try:
            rows = await connection.fetch(
                "SELECT table_name FROM information_schema.tables "
                "WHERE table_schema = 'public' ORDER BY table_name"
            )
            indexes = await connection.fetch(
                "SELECT indexname FROM pg_indexes WHERE schemaname = 'public'"
            )
            triggers = await connection.fetch(
                "SELECT tgname FROM pg_trigger WHERE NOT tgisinternal"
            )
            comments = await connection.fetchval(
                "SELECT obj_description('admin_user'::regclass, 'pg_class')"
            )
            return (
                {row["table_name"] for row in rows},
                {row["indexname"] for row in indexes},
                {row["tgname"] for row in triggers},
                comments,
            )
        finally:
            await connection.close()

    tables, indexes, triggers, comments = asyncio.run(inspect())
    assert set(TABLES) <= tables
    assert "uk_admin_user_uid" in indexes
    assert "uk_role_permission_role_code" in indexes
    assert len(triggers) >= len(TABLES)
    assert comments == "用户表"


def test_updated_at_trigger_refreshes_timestamp(pg_client):
    async def touch():
        connection = await asyncpg.connect(_asyncpg_dsn())
        try:
            original = await connection.fetchrow(
                "SELECT id, created_at, updated_at FROM organization LIMIT 1"
            )
            await asyncio.sleep(0.05)
            await connection.execute(
                "UPDATE organization SET description = description || ' ' WHERE id = $1",
                original["id"],
            )
            updated = await connection.fetchrow(
                "SELECT created_at, updated_at FROM organization WHERE id = $1",
                original["id"],
            )
            return original, updated
        finally:
            await connection.close()

    original, updated = asyncio.run(touch())
    assert updated["updated_at"] > original["updated_at"]
    assert updated["created_at"] == original["created_at"]


def test_full_admin_flow_on_postgres(pg_client):
    token = pg_client.post(
        "/api/auth/login", json={"username": "admin", "password": "admin"}
    ).json()["data"]["token"]
    headers = {"Authorization": "Bearer {}".format(token)}

    me = pg_client.get("/api/me", headers=headers).json()["data"]
    assert me["is_super"] is True
    assert len(me["permissions"]) == 21

    organization = pg_client.post(
        "/api/organizations",
        headers=headers,
        json={"name": "Postgres 组织", "parent_id": None, "description": "", "status": 1},
    ).json()["data"]
    assert len(str(organization["org_id"])) == 8

    permission = pg_client.post(
        "/api/permissions",
        headers=headers,
        json={
            "code": "pg.read",
            "name": "PG 读取",
            "description": "",
            "org_id": organization["org_id"],
        },
    )
    assert permission.status_code == 201

    role = pg_client.post(
        "/api/roles",
        headers=headers,
        json={
            "name": "PG 角色",
            "description": "",
            "org_id": organization["org_id"],
            "permission_codes": ["pg.read"],
        },
    )
    assert role.status_code == 201

    user = pg_client.post(
        "/api/users",
        headers=headers,
        json={
            "username": "pg_user",
            "nickname": "",
            "phone_num": "",
            "email": "",
            "password": "pg_password",
            "org_id": organization["org_id"],
            "status": 1,
        },
    )
    assert user.status_code == 201

    uid = user.json()["data"]["user"]["uid"]
    assigned = pg_client.put(
        "/api/users/{}/roles".format(uid),
        headers=headers,
        json={"role_ids": [role.json()["data"]["role_id"]]},
    )
    assert assigned.status_code == 200

    user_token = pg_client.post(
        "/api/auth/login", json={"username": "pg_user", "password": "pg_password"}
    ).json()["data"]["token"]
    user_me = pg_client.get(
        "/api/me", headers={"Authorization": "Bearer {}".format(user_token)}
    ).json()["data"]
    assert [item["code"] for item in user_me["permissions"]] == ["pg.read"]

    tree = pg_client.get("/api/organizations/tree", headers=headers).json()["data"]
    assert any(item["name"] == "Postgres 组织" for item in tree)
