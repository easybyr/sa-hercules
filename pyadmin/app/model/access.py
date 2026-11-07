"""用户角色、角色权限关联表实体。"""

from __future__ import annotations

from sqlalchemy import Column, DateTime, Integer, String, func

from app.model.base import Base, BigIntColumn, BigIntPrimaryKey, JsonColumn


class UserRole(Base):
    """用户角色关联表。业务唯一键为 (uid, role_id)。"""

    __tablename__ = "user_role"

    id = Column("id", BigIntPrimaryKey, primary_key=True, autoincrement=True, comment="主键ID")
    uid = Column("uid", BigIntColumn, nullable=False, comment="用户业务ID")
    role_id = Column("role_id", BigIntColumn, nullable=False, comment="角色业务ID")
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


class RolePermission(Base):
    """角色权限关联表。业务唯一键为 (role_id, permission_code)。"""

    __tablename__ = "role_permission"

    id = Column("id", BigIntPrimaryKey, primary_key=True, autoincrement=True, comment="主键ID")
    role_id = Column("role_id", BigIntColumn, nullable=False, comment="角色业务ID")
    permission_code = Column(
        "permission_code", String(100), nullable=False, comment="权限编码，对应permission.code"
    )
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
