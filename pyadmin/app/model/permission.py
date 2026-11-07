"""权限表实体。"""

from __future__ import annotations

from sqlalchemy import Column, DateTime, Integer, String, func

from app.model.base import Base, BigIntColumn, BigIntPrimaryKey, JsonColumn


class Permission(Base):
    """权限表。同一组织内 code 与 name 均唯一。"""

    __tablename__ = "permission"

    id = Column("id", BigIntPrimaryKey, primary_key=True, autoincrement=True, comment="主键ID")
    code = Column("code", String(100), nullable=False, comment="权限业务编码")
    name = Column("name", String(100), nullable=False, comment="权限名称")
    description = Column(
        "description", String(500), nullable=False, server_default="", comment="权限描述"
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
