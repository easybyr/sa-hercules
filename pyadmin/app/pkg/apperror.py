"""可安全返回给客户端的业务错误。

与 Go 版本一致：错误最终以 HTTP 状态码 + 统一响应体 `{code, message, data}` 返回，
`code` 取 HTTP 状态整数，避免前后端各自维护一套业务码。
"""

from __future__ import annotations


class AppError(Exception):
    """携带 HTTP 语义的业务异常。"""

    def __init__(self, status: int, reason: str, message: str) -> None:
        super().__init__(message)
        self.status = int(status)
        self.reason = reason
        self.message = message


def bad_request(reason: str, message: str) -> AppError:
    """400 参数或业务规则错误。"""

    return AppError(400, reason, message)


def unauthorized(reason: str, message: str) -> AppError:
    """401 未登录或登录态失效。"""

    return AppError(401, reason, message)


def forbidden(reason: str, message: str) -> AppError:
    """403 无权限或跨组织访问。"""

    return AppError(403, reason, message)


def not_found(resource: str) -> AppError:
    """404 资源不存在。"""

    return AppError(404, "NOT_FOUND", "{}不存在".format(resource))


def conflict(reason: str, message: str) -> AppError:
    """409 唯一性冲突或乐观锁冲突。"""

    return AppError(409, reason, message)


def version_conflict() -> AppError:
    """乐观锁版本冲突。"""

    return conflict("VERSION_CONFLICT", "数据已被其他请求修改，请刷新后重试")


def invalid_status() -> AppError:
    """状态值非法。"""

    return bad_request("INVALID_STATUS", "状态值无效")


def internal_error() -> AppError:
    """内部错误，避免向客户端泄露数据库细节。"""

    return AppError(500, "INTERNAL_ERROR", "服务暂时不可用")
