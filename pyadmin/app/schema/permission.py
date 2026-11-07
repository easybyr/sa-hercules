"""权限相关请求与响应模型。"""

from __future__ import annotations

from datetime import datetime
from typing import Any, Optional

from pydantic import field_validator

from app.pkg.timeutil import to_iso
from app.schema.pagination import Schema
from app.schema.validators import check_optional_text_length, check_text_length


class PermissionCreateRequest(Schema):
    """新建权限：编码需为 report.read 形式的小写点分格式。"""

    code: str
    name: str
    description: str = ""
    org_id: int

    @field_validator("code")
    @classmethod
    def _validate_code(cls, value: str) -> str:
        return check_text_length(value, 3, 100, "权限编码")

    @field_validator("name")
    @classmethod
    def _validate_name(cls, value: str) -> str:
        return check_text_length(value, 1, 100, "权限名称")

    @field_validator("description")
    @classmethod
    def _validate_description(cls, value: str) -> str:
        return check_optional_text_length(value, 500, "权限描述")


class PermissionUpdateRequest(Schema):
    """修改权限：编码不可修改。"""

    name: str
    description: str = ""
    version: int

    @field_validator("name")
    @classmethod
    def _validate_name(cls, value: str) -> str:
        return check_text_length(value, 1, 100, "权限名称")

    @field_validator("description")
    @classmethod
    def _validate_description(cls, value: str) -> str:
        return check_optional_text_length(value, 500, "权限描述")


class PermissionView(Schema):
    """权限视图。"""

    id: int
    code: str
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
