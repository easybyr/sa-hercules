"""角色表实体。"""

from __future__ import annotations

from sqlalchemy import Column, DateTime, Integer, String, func

from app.model.base import Base, BigIntColumn, BigIntPrimaryKey, JsonColumn


class Role(Base):
    """角色表。role_id 全局唯一；用户个人角色使用负数 role_id（系统内部）。"""

    __tablename__ = "role"

    id = Column("id", BigIntPrimaryKey, primary_key=True, autoincrement=True, comment="主键ID")
    role_id = Column("role_id", BigIntColumn, nullable=False, comment="角色业务ID")
    name = Column("name", String(100), nullable=False, comment="角色名称")
    description = Column(
        "description", String(500), nullable=False, server_default="", comment="角色描述"
    )
    org_id = Column("org_id", BigIntColumn, nullable=False, comment="所属组织业务ID")
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
