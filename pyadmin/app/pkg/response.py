"""统一接口响应。

响应体固定为 `{code, message, data}`：
- code：int，成功为 0，失败取对应 HTTP 状态整数；
- message：string，结果说明；
- data：any，业务数据，无数据时为 null。
"""

from __future__ import annotations

from typing import Any

from fastapi.encoders import jsonable_encoder
from fastapi.responses import JSONResponse
from pydantic import BaseModel


class ApiEnvelope(BaseModel):
    """统一响应体模型，用于 OpenAPI 文档。"""

    code: int
    message: str
    data: Any = None


def envelope(code: int, message: str, data: Any = None) -> dict:
    """构造响应字典。"""

    return {"code": int(code), "message": message, "data": data}


def success(data: Any = None, status_code: int = 200, message: str = "success") -> JSONResponse:
    """成功响应。"""

    return JSONResponse(status_code=status_code, content=jsonable_encoder(envelope(0, message, data)))


def created(data: Any = None) -> JSONResponse:
    """创建成功响应，状态码 201、message 为 created。"""

    return success(data, status_code=201, message="created")


def failure(status: int, message: str) -> JSONResponse:
    """失败响应。"""

    return JSONResponse(status_code=status, content=envelope(status, message, None))
