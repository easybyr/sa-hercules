"""用户相关请求与响应模型。"""

from __future__ import annotations

from datetime import datetime
from typing import List, Optional

from pydantic import field_validator

from app.consts.status import Status
from app.pkg.timeutil import to_iso
from app.schema.pagination import Schema
from app.schema.validators import (
    check_email,
    check_optional_text_length,
    check_text_length,
)


class UserCreateRequest(Schema):
    """管理员创建用户：无需提交 uid，密码留空时由后端生成随机初始密码。"""

    username: str
    nickname: str = ""
    phone_num: str = ""
    email: str = ""
    password: str = ""
    org_id: int
    status: Optional[int] = None

    @field_validator("username")
    @classmethod
    def _validate_username(cls, value: str) -> str:
        return check_text_length(value, 3, 64, "用户名")

    @field_validator("nickname")
    @classmethod
    def _validate_nickname(cls, value: str) -> str:
        return check_optional_text_length(value, 100, "昵称")

    @field_validator("phone_num")
    @classmethod
    def _validate_phone(cls, value: str) -> str:
        return check_optional_text_length(value, 32, "手机号")

    @field_validator("email")
    @classmethod
    def _validate_email(cls, value: str) -> str:
        return check_email(value)

    @field_validator("password")
    @classmethod
    def _validate_password(cls, value: str) -> str:
        return check_optional_text_length(value, 72, "密码")


class UserUpdateRequest(Schema):
    """修改用户/个人资料：password 留空表示不修改。"""

    username: str
    nickname: str = ""
    phone_num: str = ""
    email: str = ""
    password: str = ""
    status: Optional[int] = None
    version: int

    @field_validator("username")
    @classmethod
    def _validate_username(cls, value: str) -> str:
        return check_text_length(value, 3, 64, "用户名")

    @field_validator("nickname")
    @classmethod
    def _validate_nickname(cls, value: str) -> str:
        return check_optional_text_length(value, 100, "昵称")

    @field_validator("phone_num")
    @classmethod
    def _validate_phone(cls, value: str) -> str:
        return check_optional_text_length(value, 32, "手机号")

    @field_validator("email")
    @classmethod
    def _validate_email(cls, value: str) -> str:
        return check_email(value)

    @field_validator("password")
    @classmethod
    def _validate_password(cls, value: str) -> str:
        return check_optional_text_length(value, 72, "密码")


class RoleIDListRequest(Schema):
    """用户角色分配参数。"""

    role_ids: List[int] = []


class PermissionCodeListRequest(Schema):
    """权限编码列表参数。"""

    permission_codes: List[str] = []


class UserView(Schema):
    """用户接口视图，status 以字符串返回。"""

    id: int
    uid: int
    username: str
    nickname: str = ""
    phone_num: str = ""
    email: str = ""
    org_id: int
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


class UserCreateResult(Schema):
    """新建用户结果，初始密码只在创建成功时返回一次。"""

    user: UserView
    initial_password: str
