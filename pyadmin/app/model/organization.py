"""组织表实体。

organization.parent_id 保存直接上级组织的数据库主键 id；根组织为 NULL。
"""

from __future__ import annotations

from sqlalchemy import Column, DateTime, Integer, String, func

from app.model.base import Base, BigIntColumn, BigIntPrimaryKey, JsonColumn


class Organization(Base):
    """组织表。"""

    __tablename__ = "organization"

    id = Column("id", BigIntPrimaryKey, primary_key=True, autoincrement=True, comment="主键ID")
    org_id = Column("org_id", BigIntColumn, nullable=False, comment="组织业务ID，最多8位数字")
    name = Column("name", String(100), nullable=False, comment="组织名称")
    parent_id = Column("parent_id", BigIntColumn, nullable=True, comment="直接上级组织主键ID")
    description = Column(
        "description", String(500), nullable=False, server_default="", comment="组织描述"
    )
    status = Column("status", Integer, nullable=False, server_default="1", comment="状态：1正常，2禁用")
    extra = Column("extra", JsonColumn, nullable=True, comment="扩展信息")
    version = Column("version", Integer, nullable=False, server_default="1", comment="乐观锁版本号")
    created_at = Column(
        "created_at",
        DateTime(timezone=True),
        nullable=False,
        server_default=func.now(),
        comment="创建时间",
    )
    updated_at = Column(
        "updated_at",
        DateTime(timezone=True),
        nullable=False,
        server_default=func.now(),
        onupdate=func.now(),
        comment="更新时间",
    )
