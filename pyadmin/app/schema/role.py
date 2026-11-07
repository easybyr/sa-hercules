"""角色相关请求与响应模型。"""

from __future__ import annotations

from datetime import datetime
from typing import Any, List, Optional

from pydantic import field_validator

from app.pkg.timeutil import to_iso
from app.schema.pagination import Schema
from app.schema.validators import check_optional_text_length, check_text_length


class RoleCreateRequest(Schema):
    """新建角色：无需提交 role_id，由后端生成；可同时提交权限编码。"""

    name: str
    description: str = ""
    org_id: int
    permission_codes: Optional[List[str]] = None

    @field_validator("name")
    @classmethod
    def _validate_name(cls, value: str) -> str:
        return check_text_length(value, 1, 100, "角色名称")

    @field_validator("description")
    @classmethod
    def _validate_description(cls, value: str) -> str:
        return check_optional_text_length(value, 500, "角色描述")


class RoleUpdateRequest(Schema):
    """修改角色：permission_codes 不传表示保留原权限。"""

    name: str
    description: str = ""
    version: int
    permission_codes: Optional[List[str]] = None

    @field_validator("name")
    @classmethod
    def _validate_name(cls, value: str) -> str:
        return check_text_length(value, 1, 100, "角色名称")

    @field_validator("description")
    @classmethod
    def _validate_description(cls, value: str) -> str:
        return check_optional_text_length(value, 500, "角色描述")


class RoleView(Schema):
    """角色视图。"""

    id: int
    role_id: int
    name: str
    description: str = ""
    org_id: int
    extra: Optional[Any] = None
    version: int
    created_at: Optional[str] = None
    updated_at: Optional[str] = None

    @field_validator("created_at", "updated_at", mode="before")
    @classmethod
    def _serialize_time(cls, value: object) -> object:
        if isinstance(value, datetime):
            return to_iso(value)
        return value


class RoleDetailView(Schema):
    """角色详情：角色信息 + 权限编码列表。"""

    role: RoleView
    permission_codes: List[str] = []
