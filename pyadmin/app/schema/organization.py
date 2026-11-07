"""组织相关请求与响应模型。"""

from __future__ import annotations

from datetime import datetime
from typing import List, Optional

from pydantic import field_validator, model_serializer

from app.consts.status import Status
from app.pkg.timeutil import to_iso
from app.schema.pagination import Schema
from app.schema.validators import check_optional_text_length, check_text_length


class OrganizationCreateRequest(Schema):
    """新建组织：无需提交 org_id，由后端生成。"""

    name: str
    parent_id: Optional[int] = None
    description: str = ""
    status: Optional[int] = None

    @field_validator("name")
    @classmethod
    def _validate_name(cls, value: str) -> str:
        return check_text_length(value, 1, 100, "组织名称")

    @field_validator("description")
    @classmethod
    def _validate_description(cls, value: str) -> str:
        return check_optional_text_length(value, 500, "组织描述")


class OrganizationUpdateRequest(Schema):
    """修改组织：使用 version 做乐观锁。"""

    name: str
    parent_id: Optional[int] = None
    description: str = ""
    status: int
    version: int

    @field_validator("name")
    @classmethod
    def _validate_name(cls, value: str) -> str:
        return check_text_length(value, 1, 100, "组织名称")

    @field_validator("description")
    @classmethod
    def _validate_description(cls, value: str) -> str:
        return check_optional_text_length(value, 500, "组织描述")


class OrganizationAssignmentRequest(Schema):
    """用户组织分配参数。"""

    org_id: int


class OrganizationView(Schema):
    """组织接口视图，status 以字符串返回。"""

    id: int
    org_id: int
    name: str
    parent_id: Optional[int] = None
    description: str = ""
    status: str
    version: int
    created_at: Optional[str] = None
    updated_at: Optional[str] = None

    @field_validator("status", mode="before")
    @classmethod
    def _status_label(cls, value: object) -> object:
        if isinstance(value, str):
            return value
        return Status.label_of(int(value))

    @field_validator("created_at", "updated_at", mode="before")
    @classmethod
    def _serialize_time(cls, value: object) -> object:
        if isinstance(value, datetime):
            return to_iso(value)
        return value


class OrganizationTreeNode(OrganizationView):
    """组织树节点，children 为空时不输出（递归生效）。"""

    children: Optional[List["OrganizationTreeNode"]] = None

    @model_serializer(mode="wrap")
    def _serialize_children(self, handler):
        """children 为 None 时省略该字段，与 Go 的 omitempty 行为一致。"""

        data = handler(self)
        if data.get("children") is None:
            data.pop("children", None)
        return data
