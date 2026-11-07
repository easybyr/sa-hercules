"""用户管理接口测试。"""

from __future__ import annotations

from tests.conftest import auth_headers, login
from tests.helpers import (
    SUPER_ORG_ID,
    create_organization,
    create_role,
    create_user,
    grant_permissions,
    payload_of,
)


def _user_by_username(client, headers, username):
    listing = payload_of(
        client.get("/api/users", headers=headers, params={"q": username})
    )
    return next(item for item in listing["list"] if item["username"] == username)


def test_create_user_generates_initial_password_once(client, admin_headers):
    response = create_user(client, admin_headers, "alice", nickname="Alice")
    assert response.status_code == 201
    assert response.json()["message"] == "created"
    data = payload_of(response)
    assert data["user"]["username"] == "alice"
    assert data["user"]["org_id"] == SUPER_ORG_ID
    assert data["user"]["status"] == "normal"
    assert len(str(data["user"]["uid"])) == 10
    assert len(data["initial_password"]) == 16

    token = login(client, "alice", data["initial_password"])
    me = payload_of(client.get("/api/me", headers=auth_headers(token)))
    assert me["user"]["uid"] == data["user"]["uid"]


def test_create_user_rejects_short_explicit_password(client, admin_headers):
    response = create_user(client, admin_headers, "shortpwd", password="12345")
    assert response.status_code == 400
    assert response.json()["message"] == "密码至少需要6位"


def test_create_user_rejects_duplicate_username(client, admin_headers):
    payload_of(create_user(client, admin_headers, "duplicate_user"))
    response = create_user(client, admin_headers, "duplicate_user")
    assert response.status_code == 409
    assert response.json()["message"] == "用户名已存在"


def test_create_user_requires_organization_scope(client, admin_headers):
    first = payload_of(create_organization(client, admin_headers, "用户组织A"))
    payload_of(create_organization(client, admin_headers, "用户组织B"))
    scoped = payload_of(
        create_user(client, admin_headers, "scoped_admin", org_id=first["org_id"], password="initial123")
    )["user"]
    grant_permissions(
        client, admin_headers, scoped["uid"], ["user.create"], org_id=first["org_id"]
    )
    headers = auth_headers(login(client, "scoped_admin", "initial123"))

    response = create_user(client, headers, "cross_org_user", org_id=SUPER_ORG_ID)
    assert response.status_code == 403
    assert response.json()["message"] == "不能访问其他组织的数据"


def test_list_users_supports_keyword_and_organization_filter(client, admin_headers):
    organization = payload_of(create_organization(client, admin_headers, "列表组织"))
    payload_of(
        create_user(client, admin_headers, "keyword_target", org_id=organization["org_id"], nickname="目标昵称")
    )
    payload_of(create_user(client, admin_headers, "other_user"))

    by_keyword = payload_of(
        client.get("/api/users", headers=admin_headers, params={"q": "keyword"})
    )
    assert [item["username"] for item in by_keyword["list"]] == ["keyword_target"]

    by_nickname = payload_of(
        client.get("/api/users", headers=admin_headers, params={"q": "目标昵称"})
    )
    assert [item["username"] for item in by_nickname["list"]] == ["keyword_target"]

    by_org = payload_of(
        client.get(
            "/api/users",
            headers=admin_headers,
            params={"org_id": organization["org_id"]},
        )
    )
    assert [item["username"] for item in by_org["list"]] == ["keyword_target"]
    assert by_org["total"] == 1


def test_user_detail_contains_roles_permissions_and_direct_codes(client, admin_headers):
    user = payload_of(create_user(client, admin_headers, "detail_user"))["user"]
    role = payload_of(create_role(client, admin_headers, "详情角色"))

    assigned = client.put(
        "/api/users/{}/roles".format(user["uid"]),
        headers=admin_headers,
        json={"role_ids": [role["role_id"]]},
    )
    assert assigned.status_code == 200

    detail = payload_of(
        client.get("/api/users/{}".format(user["uid"]), headers=admin_headers)
    )
    assert detail["user"]["username"] == "detail_user"
    assert [item["name"] for item in detail["roles"]] == ["详情角色"]
    assert detail["permissions"] == []
    assert detail["direct_permission_codes"] == []
    assert detail["organization"]["org_id"] == SUPER_ORG_ID


def test_update_user_uses_version_and_checks_username_conflict(client, admin_headers):
    user = payload_of(create_user(client, admin_headers, "update_target"))["user"]
    payload_of(create_user(client, admin_headers, "existing_name"))

    conflict = client.put(
        "/api/users/{}".format(user["uid"]),
        headers=admin_headers,
        json={
            "username": "existing_name",
            "nickname": "",
            "phone_num": "",
            "email": "",
            "password": "",
            "status": 1,
            "version": user["version"],
        },
    )
    assert conflict.status_code == 409
    assert conflict.json()["message"] == "用户名已存在"

    updated = client.put(
        "/api/users/{}".format(user["uid"]),
        headers=admin_headers,
        json={
            "username": "update_target",
            "nickname": "改名",
            "phone_num": "13700000000",
            "email": "target@example.com",
            "password": "",
            "status": 2,
            "version": user["version"],
        },
    )
    assert updated.status_code == 200
    data = payload_of(updated)
    assert data["nickname"] == "改名"
    assert data["status"] == "disabled"
    assert data["version"] == user["version"] + 1

    stale = client.put(
        "/api/users/{}".format(user["uid"]),
        headers=admin_headers,
        json={
            "username": "update_target",
            "nickname": "",
            "phone_num": "",
            "email": "",
            "password": "",
            "status": 1,
            "version": user["version"],
        },
    )
    assert stale.status_code == 409


def test_disabled_user_cannot_login_and_token_becomes_invalid(client, admin_headers):
    user = payload_of(create_user(client, admin_headers, "disabled_user", password="initial123"))["user"]
    token = login(client, "disabled_user", "initial123")
    assert client.get("/api/me", headers=auth_headers(token)).status_code == 200

    client.put(
        "/api/users/{}".format(user["uid"]),
        headers=admin_headers,
        json={
            "username": "disabled_user",
            "nickname": "",
            "phone_num": "",
            "email": "",
            "password": "",
            "status": 2,
            "version": user["version"],
        },
    )

    login_attempt = client.post(
        "/api/auth/login", json={"username": "disabled_user", "password": "initial123"}
    )
    assert login_attempt.status_code == 403
    assert login_attempt.json()["message"] == "用户已被禁用"

    session = client.get("/api/me", headers=auth_headers(token))
    assert session.status_code == 401
    assert session.json()["message"] == "登录状态已失效"


def test_assign_roles_rejects_invalid_and_foreign_roles(client, admin_headers):
    user = payload_of(create_user(client, admin_headers, "role_assignee"))["user"]
    other_org = payload_of(create_organization(client, admin_headers, "角色隔离组织"))
    foreign_role = payload_of(
        create_role(client, admin_headers, "外部角色", org_id=other_org["org_id"])
    )

    invalid = client.put(
        "/api/users/{}/roles".format(user["uid"]),
        headers=admin_headers,
        json={"role_ids": [999999]},
    )
    assert invalid.status_code == 400
    assert invalid.json()["message"] == "角色不存在或不属于用户组织"

    foreign = client.put(
        "/api/users/{}/roles".format(user["uid"]),
        headers=admin_headers,
        json={"role_ids": [foreign_role["role_id"]]},
    )
    assert foreign.status_code == 400

    duplicated = client.put(
        "/api/users/{}/roles".format(user["uid"]),
        headers=admin_headers,
        json={"role_ids": [foreign_role["role_id"], foreign_role["role_id"]]},
    )
    assert duplicated.status_code == 400


def test_assign_permissions_uses_personal_role_and_is_cleared_by_role_update(client, admin_headers):
    user = payload_of(create_user(client, admin_headers, "direct_user"))["user"]
    role = payload_of(
        create_role(client, admin_headers, "直接权限角色", permission_codes=["user.read"])
    )

    assigned = client.put(
        "/api/users/{}/permissions".format(user["uid"]),
        headers=admin_headers,
        json={"permission_codes": ["user.read"]},
    )
    assert assigned.status_code == 200

    detail = payload_of(
        client.get("/api/users/{}".format(user["uid"]), headers=admin_headers)
    )
    assert detail["direct_permission_codes"] == ["user.read"]
    assert [item["code"] for item in detail["permissions"]] == ["user.read"]

    client.put(
        "/api/users/{}/roles".format(user["uid"]),
        headers=admin_headers,
        json={"role_ids": [role["role_id"]]},
    )
    after = payload_of(
        client.get("/api/users/{}".format(user["uid"]), headers=admin_headers)
    )
    assert after["direct_permission_codes"] == []
    assert [item["code"] for item in after["permissions"]] == ["user.read"]


def test_assign_permissions_rejects_unknown_code(client, admin_headers):
    user = payload_of(create_user(client, admin_headers, "bad_permission_user"))["user"]
    response = client.put(
        "/api/users/{}/permissions".format(user["uid"]),
        headers=admin_headers,
        json={"permission_codes": ["unknown.code"]},
    )
    assert response.status_code == 400
    assert response.json()["message"] == "权限不存在或不属于当前组织"


def test_assign_organization_requires_super_admin_and_resets_access(client, admin_headers):
    source = payload_of(create_organization(client, admin_headers, "源组织"))
    target = payload_of(create_organization(client, admin_headers, "目标组织"))
    user = payload_of(
        create_user(client, admin_headers, "moving_user", org_id=source["org_id"], password="initial123")
    )["user"]
    role = payload_of(create_role(client, admin_headers, "源角色", org_id=source["org_id"]))
    client.put(
        "/api/users/{}/roles".format(user["uid"]),
        headers=admin_headers,
        json={"role_ids": [role["role_id"]]},
    )
    token = login(client, "moving_user", "initial123")

    moved = client.put(
        "/api/users/{}/organization".format(user["uid"]),
        headers=admin_headers,
        json={"org_id": target["org_id"]},
    )
    assert moved.status_code == 200

    detail = payload_of(
        client.get("/api/users/{}".format(user["uid"]), headers=admin_headers)
    )
    assert detail["user"]["org_id"] == target["org_id"]
    assert detail["roles"] == []

    stale = client.get("/api/me", headers=auth_headers(token))
    assert stale.status_code == 401
    assert stale.json()["message"] == "登录状态已失效"


def test_delete_user_cleans_relations_and_protects_super_admin(client, admin_headers):
    user = payload_of(create_user(client, admin_headers, "delete_me"))["user"]
    client.put(
        "/api/users/{}/permissions".format(user["uid"]),
        headers=admin_headers,
        json={"permission_codes": ["user.read"]},
    )

    deleted = client.delete("/api/users/{}".format(user["uid"]), headers=admin_headers)
    assert deleted.status_code == 200
    assert payload_of(deleted) is None
    assert client.get("/api/users/{}".format(user["uid"]), headers=admin_headers).status_code == 404

    protected = client.delete("/api/users/1000000000", headers=admin_headers)
    assert protected.status_code == 400
    assert protected.json()["message"] == "不能删除超级管理员"


def test_normal_user_cannot_read_other_organization_members(client, admin_headers):
    first = payload_of(create_organization(client, admin_headers, "受限组织A"))
    second = payload_of(create_organization(client, admin_headers, "受限组织B"))
    payload_of(create_user(client, admin_headers, "restricted_a", org_id=first["org_id"], password="initial123"))
    outsider = payload_of(
        create_user(client, admin_headers, "restricted_b", org_id=second["org_id"])
    )["user"]

    reader = _user_by_username(client, admin_headers, "restricted_a")
    grant_permissions(
        client, admin_headers, reader["uid"], ["user.read"], org_id=first["org_id"]
    )
    headers = auth_headers(login(client, "restricted_a", "initial123"))

    listing = payload_of(client.get("/api/users", headers=headers))
    assert all(item["org_id"] == first["org_id"] for item in listing["list"])

    forbidden = client.get("/api/users/{}".format(outsider["uid"]), headers=headers)
    assert forbidden.status_code == 403
    assert forbidden.json()["message"] == "不能访问其他组织的数据"

    cross_list = client.get(
        "/api/users", headers=headers, params={"org_id": second["org_id"]}
    )
    assert cross_list.status_code == 403


def test_disabled_organization_blocks_registration_and_login(client, admin_headers):
    organization = payload_of(create_organization(client, admin_headers, "待禁用组织"))
    member = payload_of(
        create_user(
            client,
            admin_headers,
            "org_member",
            org_id=organization["org_id"],
            password="initial123",
        )
    )["user"]
    token = login(client, "org_member", "initial123")

    disabled = client.put(
        "/api/organizations/{}".format(organization["id"]),
        headers=admin_headers,
        json={
            "name": organization["name"],
            "parent_id": None,
            "description": "",
            "status": 2,
            "version": organization["version"],
        },
    )
    assert disabled.status_code == 200
    assert payload_of(disabled)["status"] == "disabled"

    register = client.post(
        "/api/auth/register",
        json={
            "username": "new_member",
            "password": "register123",
            "org_id": organization["org_id"],
        },
    )
    assert register.status_code == 400
    assert register.json()["message"] == "组织不存在或不可用"

    created = create_user(client, admin_headers, "blocked_member", org_id=organization["org_id"])
    assert created.status_code == 400
    assert created.json()["message"] == "组织不存在或不可用"

    denied = client.post(
        "/api/auth/login", json={"username": "org_member", "password": "initial123"}
    )
    assert denied.status_code == 403
    assert denied.json()["message"] == "所属组织不可用"

    stale = client.get("/api/me", headers=auth_headers(token))
    assert stale.status_code == 401
    assert stale.json()["message"] == "所属组织不可用"
    assert member["org_id"] == organization["org_id"]
