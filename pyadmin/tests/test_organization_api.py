"""组织管理接口测试：树、分页、搜索、层级校验与删除保护。"""

from __future__ import annotations

from tests.conftest import auth_headers, login
from tests.helpers import (
    SUPER_ORG_ID,
    create_organization,
    create_permission,
    create_user,
    grant_permissions,
    payload_of,
)


def test_list_organizations_contains_super_admin_org(client, admin_headers):
    response = client.get("/api/organizations", headers=admin_headers)
    assert response.status_code == 200
    data = payload_of(response)
    assert data["page"] == 1
    assert data["page_size"] == 10
    assert data["total"] >= 1
    names = [item["name"] for item in data["list"]]
    assert "系统管理组织" in names


def test_page_size_is_clamped_to_maximum(client, admin_headers):
    response = client.get(
        "/api/organizations", headers=admin_headers, params={"page_size": 5000}
    )
    assert payload_of(response)["page_size"] == 1000


def test_create_organization_generates_business_id(client, admin_headers):
    response = create_organization(client, admin_headers, "研发中心", description="技术团队")
    assert response.status_code == 201
    assert response.json()["message"] == "created"
    data = payload_of(response)
    assert data["name"] == "研发中心"
    assert data["description"] == "技术团队"
    assert data["status"] == "normal"
    assert data["parent_id"] is None
    assert len(str(data["org_id"])) == 8
    assert data["version"] == 1


def test_create_organization_requires_super_admin(client, admin_headers):
    organization = payload_of(create_organization(client, admin_headers, "普通组织"))
    payload_of(create_user(client, admin_headers, "org_admin", org_id=organization["org_id"], password="initial123"))
    headers = auth_headers(login(client, "org_admin", "initial123"))

    response = create_organization(client, headers, "越权组织")
    assert response.status_code == 403
    assert response.json()["message"] == "没有访问该资源的权限"


def test_organization_management_rejects_non_super_user_with_permission(client, admin_headers):
    organization = payload_of(create_organization(client, admin_headers, "授权组织"))
    payload_of(
        create_permission(
            client,
            admin_headers,
            "organization.create",
            "创建组织",
            org_id=organization["org_id"],
        )
    )
    user = payload_of(
        create_user(
            client,
            admin_headers,
            "org_manager",
            org_id=organization["org_id"],
            password="initial123",
        )
    )["user"]
    assigned = client.put(
        "/api/users/{}/permissions".format(user["uid"]),
        headers=admin_headers,
        json={"permission_codes": ["organization.create"]},
    )
    assert assigned.status_code == 200

    headers = auth_headers(login(client, "org_manager", "initial123"))
    response = create_organization(client, headers, "越权组织")
    assert response.status_code == 403
    assert response.json()["message"] == "仅超级管理员可管理组织"


def test_update_organization_uses_optimistic_lock(client, admin_headers):
    created = payload_of(create_organization(client, admin_headers, "待更新组织"))

    updated = client.put(
        "/api/organizations/{}".format(created["id"]),
        headers=admin_headers,
        json={
            "name": "已更新组织",
            "parent_id": None,
            "description": "更新描述",
            "status": 1,
            "version": created["version"],
        },
    )
    assert updated.status_code == 200
    data = payload_of(updated)
    assert data["name"] == "已更新组织"
    assert data["version"] == created["version"] + 1

    stale = client.put(
        "/api/organizations/{}".format(created["id"]),
        headers=admin_headers,
        json={
            "name": "再次更新",
            "parent_id": None,
            "description": "",
            "status": 1,
            "version": created["version"],
        },
    )
    assert stale.status_code == 409
    assert stale.json()["message"] == "数据已被其他请求修改，请刷新后重试"


def test_update_organization_rejects_invalid_status(client, admin_headers):
    created = payload_of(create_organization(client, admin_headers, "状态组织"))
    response = client.put(
        "/api/organizations/{}".format(created["id"]),
        headers=admin_headers,
        json={
            "name": "状态组织",
            "parent_id": None,
            "description": "",
            "status": 9,
            "version": created["version"],
        },
    )
    assert response.status_code == 400
    assert response.json()["message"] == "状态值无效"


def test_update_and_delete_super_admin_organization_is_protected(client, admin_headers):
    tree = payload_of(client.get("/api/organizations", headers=admin_headers))
    super_org = next(item for item in tree["list"] if item["org_id"] == SUPER_ORG_ID)

    update = client.put(
        "/api/organizations/{}".format(super_org["id"]),
        headers=admin_headers,
        json={
            "name": "改名",
            "parent_id": None,
            "description": "",
            "status": 1,
            "version": super_org["version"],
        },
    )
    assert update.status_code == 400
    assert update.json()["message"] == "系统管理组织不能修改"

    delete = client.delete(
        "/api/organizations/{}".format(super_org["id"]), headers=admin_headers
    )
    assert delete.status_code == 400
    assert delete.json()["message"] == "不能删除超级管理员组织"


def test_organization_tree_nests_children(client, admin_headers):
    parent = payload_of(create_organization(client, admin_headers, "集团总部"))
    child = payload_of(
        create_organization(client, admin_headers, "华东分部", parent_id=parent["id"])
    )
    grandchild = payload_of(
        create_organization(client, admin_headers, "上海办事处", parent_id=child["id"])
    )

    response = client.get("/api/organizations/tree", headers=admin_headers)
    assert response.status_code == 200
    nodes = payload_of(response)
    root = next(item for item in nodes if item["id"] == parent["id"])
    assert root["children"][0]["id"] == child["id"]
    assert root["children"][0]["children"][0]["id"] == grandchild["id"]
    assert "children" not in root["children"][0]["children"][0]


def test_organization_tree_filters_by_keyword_keeping_ancestors(client, admin_headers):
    parent = payload_of(create_organization(client, admin_headers, "西南集团"))
    child = payload_of(
        create_organization(client, admin_headers, "成都分公司", parent_id=parent["id"])
    )

    response = client.get(
        "/api/organizations/tree", headers=admin_headers, params={"q": "成都"}
    )
    nodes = payload_of(response)
    matched_parent = next(item for item in nodes if item["id"] == parent["id"])
    assert [item["id"] for item in matched_parent["children"]] == [child["id"]]
    assert all(item["id"] in (parent["id"], child["id"]) for item in nodes)


def test_organization_cannot_be_parent_of_itself_or_create_cycle(client, admin_headers):
    first = payload_of(create_organization(client, admin_headers, "环检测A"))
    second = payload_of(
        create_organization(client, admin_headers, "环检测B", parent_id=first["id"])
    )

    self_parent = client.put(
        "/api/organizations/{}".format(first["id"]),
        headers=admin_headers,
        json={
            "name": "环检测A",
            "parent_id": first["id"],
            "description": "",
            "status": 1,
            "version": first["version"],
        },
    )
    assert self_parent.status_code == 400
    assert self_parent.json()["message"] == "组织不能作为自己的上级"

    cycle = client.put(
        "/api/organizations/{}".format(first["id"]),
        headers=admin_headers,
        json={
            "name": "环检测A",
            "parent_id": second["id"],
            "description": "",
            "status": 1,
            "version": first["version"],
        },
    )
    assert cycle.status_code == 400
    assert cycle.json()["message"] == "组织层级不能形成循环"


def test_delete_organization_rejects_non_empty_organization(client, admin_headers):
    parent = payload_of(create_organization(client, admin_headers, "非空组织"))
    payload_of(create_organization(client, admin_headers, "子组织", parent_id=parent["id"]))

    response = client.delete(
        "/api/organizations/{}".format(parent["id"]), headers=admin_headers
    )
    assert response.status_code == 409
    assert response.json()["message"] == "组织包含子组织、用户、角色或权限，不能删除"


def test_delete_empty_organization_succeeds(client, admin_headers):
    created = payload_of(create_organization(client, admin_headers, "可删组织"))
    response = client.delete(
        "/api/organizations/{}".format(created["id"]), headers=admin_headers
    )
    assert response.status_code == 200
    assert payload_of(response) is None

    listing = payload_of(client.get("/api/organizations", headers=admin_headers))
    assert all(item["id"] != created["id"] for item in listing["list"])


def test_organization_list_search_matches_name_id_and_description(client, admin_headers):
    created = payload_of(
        create_organization(client, admin_headers, "搜索目标组织", description="含关键词ABC")
    )

    by_name = payload_of(
        client.get("/api/organizations", headers=admin_headers, params={"q": "搜索目标"})
    )
    assert [item["id"] for item in by_name["list"]] == [created["id"]]

    by_org_id = payload_of(
        client.get(
            "/api/organizations",
            headers=admin_headers,
            params={"q": str(created["org_id"])},
        )
    )
    assert created["id"] in [item["id"] for item in by_org_id["list"]]

    by_description = payload_of(
        client.get("/api/organizations", headers=admin_headers, params={"q": "abc"})
    )
    assert [item["id"] for item in by_description["list"]] == [created["id"]]


def test_normal_user_only_sees_own_branch_of_organization_tree(client, admin_headers):
    root = payload_of(create_organization(client, admin_headers, "隔离根组织"))
    branch = payload_of(
        create_organization(client, admin_headers, "隔离分支", parent_id=root["id"])
    )
    sibling = payload_of(
        create_organization(client, admin_headers, "隔离旁支", parent_id=root["id"])
    )
    payload_of(create_user(client, admin_headers, "branch_user", org_id=branch["org_id"], password="initial123"))
    branch_user = payload_of(
        client.get("/api/users", headers=admin_headers, params={"q": "branch_user"})
    )["list"][0]
    grant_permissions(
        client, admin_headers, branch_user["uid"], ["organization.read"], org_id=branch["org_id"]
    )
    headers = auth_headers(login(client, "branch_user", "initial123"))

    tree = payload_of(client.get("/api/organizations/tree", headers=headers))

    def flatten(nodes):
        result = []
        for node in nodes:
            result.append(node["id"])
            result.extend(flatten(node.get("children") or []))
        return result

    visible = flatten(tree)
    assert branch["id"] in visible
    assert root["id"] in visible
    assert sibling["id"] not in visible
