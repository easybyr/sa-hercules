"""测试断言与造数辅助。"""

from __future__ import annotations

from typing import Any, Dict

from fastapi.testclient import TestClient

SUPER_ORG_ID = 10_000_000


def payload_of(result) -> Dict[str, Any]:
    """取出 data 字段并校验响应体结构。"""

    body = result.json()
    assert set(body.keys()) == {"code", "message", "data"}, body
    return body["data"]


def create_organization(
    client: TestClient,
    headers: Dict[str, str],
    name: str,
    parent_id: int | None = None,
    description: str = "",
    status: int = 1,
):
    """新建组织。"""

    return client.post(
        "/api/organizations",
        headers=headers,
        json={
            "name": name,
            "parent_id": parent_id,
            "description": description,
            "status": status,
        },
    )


def create_permission(
    client: TestClient,
    headers: Dict[str, str],
    code: str,
    name: str,
    org_id: int = SUPER_ORG_ID,
):
    """新建权限。"""

    return client.post(
        "/api/permissions",
        headers=headers,
        json={"code": code, "name": name, "description": "", "org_id": org_id},
    )


def create_role(
    client: TestClient,
    headers: Dict[str, str],
    name: str,
    org_id: int = SUPER_ORG_ID,
    permission_codes=None,
):
    """新建角色。"""

    body = {"name": name, "description": "", "org_id": org_id}
    if permission_codes is not None:
        body["permission_codes"] = permission_codes
    return client.post("/api/roles", headers=headers, json=body)


def create_user(
    client: TestClient,
    headers: Dict[str, str],
    username: str,
    org_id: int = SUPER_ORG_ID,
    password: str = "",
    nickname: str = "",
    status: int = 1,
):
    """新建用户。"""

    return client.post(
        "/api/users",
        headers=headers,
        json={
            "username": username,
            "nickname": nickname,
            "phone_num": "",
            "email": "",
            "password": password,
            "org_id": org_id,
            "status": status,
        },
    )


def grant_permissions(
    client: TestClient,
    headers: Dict[str, str],
    uid: int,
    codes,
    org_id: int = SUPER_ORG_ID,
) -> None:
    """确保权限存在，并作为用户直接权限授予该用户。"""

    for code in codes:
        create_permission(client, headers, code, code, org_id=org_id)
    response = client.put(
        "/api/users/{}/permissions".format(uid),
        headers=headers,
        json={"permission_codes": list(codes)},
    )
    assert response.status_code == 200, response.text
