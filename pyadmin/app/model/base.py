"""ORM 基类与通用列类型。

主键统一 BIGSERIAL（Postgres），测试用 SQLite 时退化为 INTEGER 以便自增；
extra 在生产库使用 JSONB，其它方言退化为通用 JSON。
"""

from __future__ import annotations

from sqlalchemy import JSON, BigInteger, Integer
from sqlalchemy.dialects.postgresql import JSONB
from sqlalchemy.orm import DeclarativeBase

# 自增主键列类型。
BigIntPrimaryKey = BigInteger().with_variant(Integer, "sqlite")

# 普通 BIGINT 列类型。
BigIntColumn = BigInteger

# 扩展信息列类型。
JsonColumn = JSON().with_variant(JSONB, "postgresql")


class Base(DeclarativeBase):
    """所有 ORM 实体的基类。"""

    pass
