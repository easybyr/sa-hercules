CREATE TABLE IF NOT EXISTS permission_operation_log (
    id           BIGSERIAL      NOT NULL,
    org_id       BIGINT         NOT NULL,
    operator_uid BIGINT         NOT NULL,
    target_type  VARCHAR(32)    NOT NULL,
    target_id    BIGINT         NOT NULL,
    operation    VARCHAR(64)    NOT NULL,
    before_data  JSONB          NOT NULL,
    after_data   JSONB          NOT NULL,
    created_at   TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id)
);
COMMENT ON TABLE permission_operation_log IS '权限操作日志表';
COMMENT ON COLUMN permission_operation_log.id IS '主键ID';
COMMENT ON COLUMN permission_operation_log.org_id IS '所属组织业务ID';
COMMENT ON COLUMN permission_operation_log.operator_uid IS '操作人用户业务ID';
COMMENT ON COLUMN permission_operation_log.target_type IS '操作对象类型：permission、role、user';
COMMENT ON COLUMN permission_operation_log.target_id IS '操作对象业务ID';
COMMENT ON COLUMN permission_operation_log.operation IS '操作类型';
COMMENT ON COLUMN permission_operation_log.before_data IS '变更前数据';
COMMENT ON COLUMN permission_operation_log.after_data IS '变更后数据';
COMMENT ON COLUMN permission_operation_log.created_at IS '操作时间';

CREATE INDEX IF NOT EXISTS idx_permission_operation_org_created_at
    ON permission_operation_log (org_id, created_at);
CREATE INDEX IF NOT EXISTS idx_permission_operation_target
    ON permission_operation_log (target_type, target_id, created_at);
CREATE INDEX IF NOT EXISTS idx_permission_operation_operator
    ON permission_operation_log (operator_uid, created_at);
