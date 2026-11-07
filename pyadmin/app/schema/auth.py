"""认证相关请求与响应模型。"""

from __future__ import annotations

from typing import Optional

from pydantic import field_validator

from app.schema.pagination import Schema
from app.schema.user import UserView
from app.schema.validators import (
    check_email,
    check_optional_text_length,
    check_text_length,
)


class LoginRequest(Schema):
    """登录参数。"""

    username: str
    password: str

    @field_validator("username")
    @classmethod
    def _validate_username(cls, value: str) -> str:
        return check_text_length(value, 3, 64, "用户名")

    @field_validator("password")
    @classmethod
    def _validate_password(cls, value: str) -> str:
        return check_text_length(value, 5, 72, "密码")


class RegisterRequest(Schema):
    """注册参数：组织以精确搜索方式匹配后提交 org_id。"""

    username: str
    password: str
    nickname: str = ""
    phone_num: str = ""
    email: str = ""
    org_id: int

    @field_validator("username")
    @classmethod
    def _validate_username(cls, value: str) -> str:
        return check_text_length(value, 3, 64, "用户名")

    @field_validator("password")
    @classmethod
    def _validate_password(cls, value: str) -> str:
        return check_text_length(value, 6, 72, "密码")

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


class LoginResult(Schema):
    """登录结果。"""

    token: str
    token_type: str = "Bearer"
    expires_in: int
    user: Optional[UserView] = None
