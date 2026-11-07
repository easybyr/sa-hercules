"""用户表实体。"""

from __future__ import annotations

from sqlalchemy import Column, DateTime, Integer, String, func

from app.model.base import Base, BigIntColumn, BigIntPrimaryKey, JsonColumn


class User(Base):
    """用户表。uid 全局唯一，org_id 为所属组织业务 ID。"""

    __tablename__ = "admin_user"

    id = Column("id", BigIntPrimaryKey, primary_key=True, autoincrement=True, comment="主键ID")
    uid = Column("uid", BigIntColumn, nullable=False, comment="全局唯一用户ID，最多10位数字")
    username = Column("username", String(64), nullable=False, comment="用户名")
    nickname = Column("nickname", String(100), nullable=False, server_default="", comment="昵称")
    phone_num = Column("phone_num", String(32), nullable=False, server_default="", comment="手机号")
    email = Column("email", String(254), nullable=False, server_default="", comment="邮箱")
    password = Column("password", String(255), nullable=False, comment="密码哈希")
    org_id = Column("org_id", BigIntColumn, nullable=False, comment="所属组织业务ID")
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
