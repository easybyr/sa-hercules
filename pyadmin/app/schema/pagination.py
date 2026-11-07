"""Schema 公共基类、分页参数与分页结果。"""

from __future__ import annotations

from typing import Generic, List, TypeVar

from pydantic import BaseModel, ConfigDict

from app.consts.pagination import DEFAULT_PAGE, DEFAULT_PAGE_SIZE, MAX_PAGE_SIZE

T = TypeVar("T")


class Schema(BaseModel):
    """所有请求/响应模型的基类。"""

    model_config = ConfigDict(from_attributes=True, populate_by_name=True)


class PageRequest:
    """分页参数：page 默认 1，page_size 默认 10、上限 1000。"""

    def __init__(
        self,
        page: int = DEFAULT_PAGE,
        page_size: int = DEFAULT_PAGE_SIZE,
    ) -> None:
        self.page = page if page >= DEFAULT_PAGE else DEFAULT_PAGE
        self.page_size = max(1, min(page_size, MAX_PAGE_SIZE))

    @property
    def offset(self) -> int:
        """SQL 偏移量。"""

        return (self.page - 1) * self.page_size


class PageResult(BaseModel, Generic[T]):
    """分页查询结果。"""

    list: List[T]
    total: int
    page: int
    page_size: int
