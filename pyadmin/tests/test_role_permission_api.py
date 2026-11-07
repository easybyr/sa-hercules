"""角色与权限接口测试。"""

from __future__ import annotations

from tests.conftest import auth_headers, login
from tests.helpers import (
    SUPER_ORG_ID,
    create_organization,
    create_permission,
    create_role,
    create_user,
    payload_of,
)


def test_create_permission_validates_code_and_uniqueness(client, admin_headers):
    created = create_permission(client, admin_headers, "report.read", "查看报表")
    assert created.status_code == 201
    assert payload_of(created)["code"] == "report.read"

    invalid_code = create_permission(client, admin_headers, "InvalidCode", "非法编码")
    assert invalid_code.status_code == 400
    assert invalid_code.json()["message"].startswith("权限编码格式无效")

    duplicate_code = create_permission(client, admin_headers, "report.read", "另一个名称")
    assert duplicate_code.status_code == 409
    assert duplicate_code.json()["message"] == "组织内权限编码已存在"

    duplicate_name = create_permission(client, admin_headers, "report.write", "查看报表")
    assert duplicate_name.status_code == 409
    assert duplicate_name.json()["message"] == "组织内权限名称已存在，请重新输入权限名称"


def test_permission_list_filters_by_organization_and_keyword(client, admin_headers):
    organization = payload_of(create_organization(client, admin_headers, "权限组织"))
    create_permission(client, admin_headers, "order.read", "查看订单")
    create_permission(client, admin_headers, "order.update", "更新订单")
    create_permission(
        client, admin_headers, "stock.read", "查看库存", org_id=organization["org_id"]
    )

    in_super_org = payload_of(
        client.get(
            "/api/permissions", headers=admin_headers, params={"org_id": SUPER_ORG_ID}
        )
    )
    codes = [item["code"] for item in in_super_org["list"]]
    assert "order.read" in codes
    assert "stock.read" not in codes

    filtered = payload_of(
        client.get("/api/permissions", headers=admin_headers, params={"q": "order"})
    )
    assert {item["code"] for item in filtered["list"]} == {"order.read", "order.update"}


def test_update_permission_requires_version_and_unique_name(client, admin_headers):
    first = payload_of(create_permission(client, admin_headers, "audit.read", "审计查看"))
    second = payload_of(create_permission(client, admin_headers, "audit.write", "审计写入"))

    renamed = client.put(
        "/api/permissions/{}/{}".format(SUPER_ORG_ID, first["code"]),
        headers=admin_headers,
        json={"name": "审计读取", "description": "更新描述", "version": first["version"]},
    )
    assert renamed.status_code == 200
    assert payload_of(renamed)["name"] == "审计读取"

    conflict = client.put(
        "/api/permissions/{}/{}".format(SUPER_ORG_ID, second["code"]),
        headers=admin_headers,
        json={"name": "审计读取", "description": "", "version": second["version"]},
    )
    assert conflict.status_code == 409

    stale = client.put(
        "/api/permissions/{}/{}".format(SUPER_ORG_ID, first["code"]),
        headers=admin_headers,
        json={"name": "再次改名", "description": "", "version": first["version"]},
    )
    assert stale.status_code == 409
    assert stale.json()["message"] == "数据已被其他请求修改，请刷新后重试"

    missing = client.put(
        "/api/permissions/{}/{}".format(SUPER_ORG_ID, "not.exists"),
        headers=admin_headers,
        json={"name": "不存在", "description": "", "version": 1},
    )
    assert missing.status_code == 404
    assert missing.json()["message"] == "权限不存在"


def test_delete_permission_cleans_role_permission_relations(client, admin_headers):
    create_permission(client, admin_headers, "cleanup.read", "待清理权限")
    role = payload_of(
        create_role(
            client, admin_headers, "清理角色", permission_codes=["cleanup.read"]
        )
    )

    deleted = client.delete(
        "/api/permissions/{}/cleanup.read".format(SUPER_ORG_ID), headers=admin_headers
    )
    assert deleted.status_code == 200

    detail = payload_of(
        client.get(
            "/api/roles/{}/{}".format(SUPER_ORG_ID, role["role_id"]), headers=admin_headers
        )
    )
    assert detail["permission_codes"] == []


def test_create_role_with_permission_codes(client, admin_headers):
    create_permission(client, admin_headers, "role.read", "查看角色")
    create_permission(client, admin_headers, "role.update", "修改角色")

    response = create_role(
        client,
        admin_headers,
        "运营角色",
        permission_codes=["role.update", "role.read"],
    )
    assert response.status_code == 201
    data = payload_of(response)
    assert data["name"] == "运营角色"
    assert len(str(data["role_id"])) == 10

    detail = payload_of(
        client.get(
            "/api/roles/{}/{}".format(SUPER_ORG_ID, data["role_id"]), headers=admin_headers
        )
    )
    assert detail["permission_codes"] == ["role.read", "role.update"]


def test_create_role_rejects_duplicate_name_and_unknown_permission(client, admin_headers):
    payload_of(create_role(client, admin_headers, "重复角色"))
    duplicated = create_role(client, admin_headers, "重复角色")
    assert duplicated.status_code == 409
    assert duplicated.json()["message"] == "组织内角色名称已存在，请重新输入角色名称"

    unknown_permission = create_role(
        client, admin_headers, "非法权限角色", permission_codes=["missing.code"]
    )
    assert unknown_permission.status_code == 400
    assert unknown_permission.json()["message"] == "权限不存在或不属于当前组织"


def test_update_role_keeps_permissions_when_codes_omitted(client, admin_headers):
    create_permission(client, admin_headers, "keep.read", "保留权限")
    role = payload_of(
        create_role(client, admin_headers, "保留角色", permission_codes=["keep.read"])
    )

    updated = client.put(
        "/api/roles/{}/{}".format(SUPER_ORG_ID, role["role_id"]),
        headers=admin_headers,
        json={"name": "保留角色改名", "description": "新描述", "version": role["version"]},
    )
    assert updated.status_code == 200
    assert payload_of(updated)["name"] == "保留角色改名"

    detail = payload_of(
        client.get(
            "/api/roles/{}/{}".format(SUPER_ORG_ID, role["role_id"]), headers=admin_headers
        )
    )
    assert detail["permission_codes"] == ["keep.read"]


def test_update_role_replaces_permissions_when_codes_provided(client, admin_headers):
    create_permission(client, admin_headers, "replace.read", "替换前")
    create_permission(client, admin_headers, "replace.write", "替换后")
    role = payload_of(
        create_role(client, admin_headers, "替换角色", permission_codes=["replace.read"])
    )

    client.put(
        "/api/roles/{}/{}".format(SUPER_ORG_ID, role["role_id"]),
        headers=admin_headers,
        json={
            "name": role["name"],
            "description": "",
            "version": role["version"],
            "permission_codes": ["replace.write"],
        },
    )
    detail = payload_of(
        client.get(
            "/api/roles/{}/{}".format(SUPER_ORG_ID, role["role_id"]), headers=admin_headers
        )
    )
    assert detail["permission_codes"] == ["replace.write"]


def test_assign_role_permissions_replaces_all(client, admin_headers):
    create_permission(client, admin_headers, "assign.read", "分配读取")
    create_permission(client, admin_headers, "assign.write", "分配写入")
    role = payload_of(create_role(client, admin_headers, "分配角色"))

    response = client.put(
        "/api/roles/{}/{}/permissions".format(SUPER_ORG_ID, role["role_id"]),
        headers=admin_headers,
        json={"permission_codes": ["assign.read", "assign.write"]},
    )
    assert response.status_code == 200

    detail = payload_of(
        client.get(
            "/api/roles/{}/{}".format(SUPER_ORG_ID, role["role_id"]), headers=admin_headers
        )
    )
    assert detail["permission_codes"] == ["assign.read", "assign.write"]

    cleared = client.put(
        "/api/roles/{}/{}/permissions".format(SUPER_ORG_ID, role["role_id"]),
        headers=admin_headers,
        json={"permission_codes": []},
    )
    assert cleared.status_code == 200
    detail = payload_of(
        client.get(
            "/api/roles/{}/{}".format(SUPER_ORG_ID, role["role_id"]), headers=admin_headers
        )
    )
    assert detail["permission_codes"] == []


def test_delete_role_cleans_user_relations(client, admin_headers):
    role = payload_of(create_role(client, admin_headers, "待删角色"))
    user = payload_of(create_user(client, admin_headers, "role_user"))["user"]
    client.put(
        "/api/users/{}/roles".format(user["uid"]),
        headers=admin_headers,
        json={"role_ids": [role["role_id"]]},
    )

    deleted = client.delete(
        "/api/roles/{}/{}".format(SUPER_ORG_ID, role["role_id"]), headers=admin_headers
    )
    assert deleted.status_code == 200

    assert (
        client.get(
            "/api/roles/{}/{}".format(SUPER_ORG_ID, role["role_id"]), headers=admin_headers
        ).status_code
        == 404
    )
    detail = payload_of(
        client.get("/api/users/{}".format(user["uid"]), headers=admin_headers)
    )
    assert detail["roles"] == []


def test_super_admin_role_is_protected(client, admin_headers):
    update = client.put(
        "/api/roles/{}/{}".format(SUPER_ORG_ID, 10_000_000),
        headers=admin_headers,
        json={"name": "改名", "description": "", "version": 1},
    )
    assert update.status_code == 400
    assert update.json()["message"] == "不能修改超级管理员角色"

    delete = client.delete(
        "/api/roles/{}/{}".format(SUPER_ORG_ID, 10_000_000), headers=admin_headers
    )
    assert delete.status_code == 400
    assert delete.json()["message"] == "不能删除系统角色"


def test_role_list_hides_personal_direct_permission_roles(client, admin_headers):
    user = payload_of(create_user(client, admin_headers, "hidden_role_user"))["user"]
    client.put(
        "/api/users/{}/permissions".format(user["uid"]),
        headers=admin_headers,
        json={"permission_codes": ["user.read"]},
    )

    listing = payload_of(client.get("/api/roles", headers=admin_headers))
    assert all(item["role_id"] > 0 for item in listing["list"])
    assert all(not item["name"].startswith("_user_") for item in listing["list"])


def test_role_endpoints_respect_organization_isolation(client, admin_headers):
    first = payload_of(create_organization(client, admin_headers, "角色组织A"))
    second = payload_of(create_organization(client, admin_headers, "角色组织B"))
    admin_user = payload_of(
        create_user(client, admin_headers, "role_admin", org_id=first["org_id"], password="initial123")
    )["user"]
    create_permission(client, admin_headers, "role.read", "查看角色", org_id=first["org_id"])
    client.put(
        "/api/users/{}/permissions".format(admin_user["uid"]),
        headers=admin_headers,
        json={"permission_codes": ["role.read"]},
    )
    headers = auth_headers(login(client, "role_admin", "initial123"))

    listing = payload_of(client.get("/api/roles", headers=headers))
    assert all(item["org_id"] == first["org_id"] for item in listing["list"])

    forbidden = client.get(
        "/api/roles/{}/1".format(second["org_id"]), headers=headers
    )
    assert forbidden.status_code == 403
    assert forbidden.json()["message"] == "不能访问其他组织的数据"
