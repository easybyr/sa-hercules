"""请求参数校验错误的中文提示。"""

from __future__ import annotations

from typing import Any, Dict, Iterable

FIELD_LABELS: Dict[str, str] = {
    "username": "用户名",
    "password": "密码",
    "nickname": "昵称",
    "phone_num": "手机号",
    "email": "邮箱",
    "org_id": "组织ID",
    "uid": "用户ID",
    "role_id": "角色ID",
    "role_ids": "角色列表",
    "permission_codes": "权限列表",
    "permission": "权限编码",
    "code": "权限编码",
    "name": "名称",
    "description": "描述",
    "parent_id": "上级组织",
    "status": "状态",
    "version": "数据版本",
    "page": "页码",
    "page_size": "每页记录数",
    "q": "搜索关键词",
}

_VALUE_ERROR_PREFIX = "Value error, "


def _field_label(location: Iterable[Any]) -> str:
    """从错误位置中取出字段名并翻译为中文标签。"""

    parts = [str(item) for item in location if item not in ("body", "query", "path")]
    if not parts:
        return "参数"
    field = parts[0]
    return FIELD_LABELS.get(field, field)


def format_validation_errors(errors: Iterable[Dict[str, Any]]) -> str:
    """把 pydantic 校验错误转换为可直接返回给用户的中文提示。"""

    first = next(iter(errors), None)
    if not first:
        return "参数校验失败"

    label = _field_label(first.get("loc", ()))
    error_type = str(first.get("type", ""))
    message = str(first.get("msg", ""))

    if error_type == "missing":
        return "{}不能为空".format(label)
    if message.startswith(_VALUE_ERROR_PREFIX):
        return message[len(_VALUE_ERROR_PREFIX) :]
    if error_type.endswith("_parsing") or error_type.endswith("_type"):
        return "{}格式不正确".format(label)
    if error_type in ("json_invalid", "model_attributes_type"):
        return "请求体格式不正确"
    return "{}取值不合法".format(label)
