"""分页参数解析：与 Go 版本一致，默认 page=1、page_size=10，上限 1000。"""

from __future__ import annotations

from fastapi import Query

from app.consts.pagination import DEFAULT_PAGE, DEFAULT_PAGE_SIZE
from app.schema.pagination import PageRequest


def page_request(
    page: int = Query(default=DEFAULT_PAGE, description="页码，从 1 开始"),
    page_size: int = Query(default=DEFAULT_PAGE_SIZE, description="每页记录数，最大 1000"),
) -> PageRequest:
    """构造经过边界修正的分页参数。"""

    return PageRequest(page=page, page_size=page_size)
