"""权限操作日志实体。"""

from __future__ import annotations

from sqlalchemy import Column, DateTime, Index, String, func

from app.model.base import Base, BigIntColumn, BigIntPrimaryKey, JsonColumn


class PermissionOperationLog(Base):
    """记录权限定义及授权关系的变更操作。"""

    __tablename__ = "permission_operation_log"
    __table_args__ = (
        Index("idx_permission_operation_org_created_at", "org_id", "created_at"),
        Index(
            "idx_permission_operation_target",
            "target_type",
            "target_id",
            "created_at",
        ),
        Index(
            "idx_permission_operation_operator", "operator_uid", "created_at"
        ),
    )

    id = Column("id", BigIntPrimaryKey, primary_key=True, autoincrement=True, comment="主键ID")
    org_id = Column("org_id", BigIntColumn, nullable=False, comment="所属组织业务ID")
    operator_uid = Column("operator_uid", BigIntColumn, nullable=False, comment="操作人用户业务ID")
    target_type = Column(
        "target_type", String(32), nullable=False, comment="操作对象类型：permission、role、user"
    )
    target_id = Column("target_id", BigIntColumn, nullable=False, comment="操作对象业务ID")
    operation = Column("operation", String(64), nullable=False, comment="操作类型")
    before_data = Column("before_data", JsonColumn, nullable=False, comment="变更前数据")
    after_data = Column("after_data", JsonColumn, nullable=False, comment="变更后数据")
    created_at = Column(
        "created_at",
        DateTime(timezone=True),
        nullable=False,
        server_default=func.now(),
        comment="操作时间",
    )
