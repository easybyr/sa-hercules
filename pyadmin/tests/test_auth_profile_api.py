"""健康检查、登录、注册与个人中心测试。"""

from __future__ import annotations

from tests.conftest import SUPER_ADMIN_PASSWORD, SUPER_ADMIN_USERNAME, auth_headers, login
from tests.helpers import create_organization, create_user, payload_of


def test_health_returns_healthy(client):
    response = client.get("/api/health")
    assert response.status_code == 200
    assert payload_of(response) == {"status": "healthy"}


def test_login_success_returns_token_and_user(client):
    response = client.post(
        "/api/auth/login",
        json={"username": SUPER_ADMIN_USERNAME, "password": SUPER_ADMIN_PASSWORD},
    )
    assert response.status_code == 200
    data = payload_of(response)
    assert data["token_type"] == "Bearer"
    assert data["expires_in"] == 12 * 3600
    assert data["user"]["username"] == "admin"
    assert data["user"]["uid"] == 1_000_000_000
    assert data["user"]["org_id"] == 10_000_000
    assert data["user"]["status"] == "normal"


def test_login_with_wrong_password_is_401(client):
    response = client.post(
        "/api/auth/login", json={"username": "admin", "password": "wrong-password"}
    )
    assert response.status_code == 401
    body = response.json()
    assert body["code"] == 401
    assert body["message"] == "用户名或密码错误"
    assert body["data"] is None


def test_login_with_unknown_user_is_401(client):
    response = client.post(
        "/api/auth/login", json={"username": "nobody", "password": "whatever"}
    )
    assert response.status_code == 401


def test_login_validation_error_is_400_with_chinese_message(client):
    response = client.post("/api/auth/login", json={"username": "ab", "password": "12345"})
    assert response.status_code == 400
    assert response.json()["code"] == 400
    assert "用户名" in response.json()["message"]


def test_protected_endpoint_requires_token(client):
    response = client.get("/api/me")
    assert response.status_code == 401
    assert response.json()["message"] == "请先登录"


def test_protected_endpoint_rejects_invalid_token(client):
    response = client.get("/api/me", headers={"Authorization": "Bearer broken"})
    assert response.status_code == 401
    assert response.json()["message"] == "登录凭证无效或已过期"


def test_me_returns_super_admin_profile(client, admin_headers):
    response = client.get("/api/me", headers=admin_headers)
    assert response.status_code == 200
    data = payload_of(response)
    assert data["is_super"] is True
    assert data["user"]["username"] == "admin"
    assert data["organization"]["org_id"] == 10_000_000
    assert data["organization"]["status"] == "normal"
    assert data["roles"][0]["role_id"] == 10_000_000
    codes = {item["code"] for item in data["permissions"]}
    assert {"user.create", "role.assign_permission", "organization.delete"} <= codes
    assert data["role_accesses"][0]["role"]["name"] == "超级管理员"
    assert len(data["role_accesses"][0]["permissions"]) == len(codes)


def test_register_creates_user_in_active_organization(client, admin_headers):
    organization = payload_of(create_organization(client, admin_headers, "注册测试组织"))

    response = client.post(
        "/api/auth/register",
        json={
            "username": "register_user",
            "password": "register123",
            "nickname": "注册用户",
            "phone_num": "13800000000",
            "email": "register@example.com",
            "org_id": organization["org_id"],
        },
    )
    assert response.status_code == 201
    assert response.json()["message"] == "created"
    data = payload_of(response)
    assert data["username"] == "register_user"
    assert data["org_id"] == organization["org_id"]
    assert data["status"] == "normal"
    assert len(str(data["uid"])) == 10

    token = login(client, "register_user", "register123")
    me = client.get("/api/me", headers=auth_headers(token))
    assert me.status_code == 200
    assert payload_of(me)["is_super"] is False


def test_register_rejects_unknown_organization(client):
    response = client.post(
        "/api/auth/register",
        json={
            "username": "no_org_user",
            "password": "register123",
            "org_id": 99_999_999,
        },
    )
    assert response.status_code == 400
    assert response.json()["message"] == "组织不存在或不可用"


def test_register_rejects_duplicate_username(client, admin_headers):
    organization = payload_of(create_organization(client, admin_headers, "重名测试组织"))
    body = {
        "username": "duplicated",
        "password": "register123",
        "org_id": organization["org_id"],
    }
    assert client.post("/api/auth/register", json=body).status_code == 201
    duplicated = client.post("/api/auth/register", json=body)
    assert duplicated.status_code == 409
    assert duplicated.json()["message"] == "用户名已存在"


def test_public_organization_search_requires_exact_match(client, admin_headers):
    payload_of(create_organization(client, admin_headers, "精确搜索公司"))

    matched = client.get("/api/public/organizations/search", params={"q": "精确搜索公司"})
    assert matched.status_code == 200
    assert len(payload_of(matched)) == 1

    partial = client.get("/api/public/organizations/search", params={"q": "精确搜索"})
    assert payload_of(partial) == []

    empty = client.get("/api/public/organizations/search", params={"q": ""})
    assert payload_of(empty) == []


def test_update_me_changes_profile_and_password(client, admin_headers):
    created = payload_of(
        create_user(client, admin_headers, "profile_user", password="initial123")
    )
    token = login(client, "profile_user", "initial123")
    me = payload_of(client.get("/api/me", headers=auth_headers(token)))

    updated = client.put(
        "/api/me",
        headers=auth_headers(token),
        json={
            "username": "profile_user",
            "nickname": "新昵称",
            "phone_num": "13900000000",
            "email": "profile@example.com",
            "password": "changed123",
            "version": me["user"]["version"],
        },
    )
    assert updated.status_code == 200
    data = payload_of(updated)
    assert data["nickname"] == "新昵称"
    assert data["email"] == "profile@example.com"
    assert data["version"] == me["user"]["version"] + 1

    assert login(client, "profile_user", "changed123")
    stale = client.post(
        "/api/auth/login", json={"username": "profile_user", "password": "initial123"}
    )
    assert stale.status_code == 401
    assert created["user"]["username"] == "profile_user"


def test_update_me_ignores_status_and_requires_version(client, admin_headers):
    payload_of(create_user(client, admin_headers, "status_user", password="initial123"))
    token = login(client, "status_user", "initial123")
    me = payload_of(client.get("/api/me", headers=auth_headers(token)))

    without_version = client.put(
        "/api/me",
        headers=auth_headers(token),
        json={"username": "status_user", "nickname": "", "phone_num": "", "email": ""},
    )
    assert without_version.status_code == 400
    assert "数据版本" in without_version.json()["message"]

    disabled = client.put(
        "/api/me",
        headers=auth_headers(token),
        json={
            "username": "status_user",
            "nickname": "",
            "phone_num": "",
            "email": "",
            "status": 2,
            "version": me["user"]["version"],
        },
    )
    assert disabled.status_code == 200
    assert payload_of(disabled)["status"] == "normal"
