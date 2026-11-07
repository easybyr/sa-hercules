-- pyadmin 初始化脚本（PostgreSQL 16）
-- 与 admin/resource/migrations/001_init.sql 等价的表结构：
--   * 主键 BIGSERIAL，业务 ID 唯一；
--   * 统一 extra(JSONB)、version、created_at、updated_at(毫秒精度)；
--   * 不使用外键，也不使用 CHECK 约束，限定值由代码校验；
--   * 表名 user 在 PostgreSQL 中为保留字，此处使用 admin_user，其余表名保持一致。

CREATE OR REPLACE FUNCTION pyadmin_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := CURRENT_TIMESTAMP(3);
    NEW.created_at := OLD.created_at;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS organization (
    id          BIGSERIAL    NOT NULL,
    org_id      BIGINT       NOT NULL,
    name        VARCHAR(100) NOT NULL,
    parent_id   BIGINT       NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    status      INT          NOT NULL DEFAULT 1,
    extra       JSONB        NULL,
    version     INT          NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id)
);
COMMENT ON TABLE organization IS '组织表';
COMMENT ON COLUMN organization.id IS '主键ID';
COMMENT ON COLUMN organization.org_id IS '组织业务ID，最多8位数字';
COMMENT ON COLUMN organization.name IS '组织名称';
COMMENT ON COLUMN organization.parent_id IS '直接上级组织主键ID，无上级时为空';
COMMENT ON COLUMN organization.description IS '组织描述';
COMMENT ON COLUMN organization.status IS '状态：1正常，2禁用';
COMMENT ON COLUMN organization.extra IS '扩展信息';
COMMENT ON COLUMN organization.version IS '乐观锁版本号';
COMMENT ON COLUMN organization.created_at IS '创建时间';
COMMENT ON COLUMN organization.updated_at IS '更新时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_organization_org_id ON organization (org_id);
CREATE INDEX IF NOT EXISTS idx_organization_parent_id ON organization (parent_id);
CREATE INDEX IF NOT EXISTS idx_organization_name ON organization (name);
CREATE INDEX IF NOT EXISTS idx_organization_updated_at ON organization (updated_at);

CREATE TABLE IF NOT EXISTS admin_user (
    id         BIGSERIAL    NOT NULL,
    uid        BIGINT       NOT NULL,
    username   VARCHAR(64)  NOT NULL,
    nickname   VARCHAR(100) NOT NULL DEFAULT '',
    phone_num  VARCHAR(32)  NOT NULL DEFAULT '',
    email      VARCHAR(254) NOT NULL DEFAULT '',
    password   VARCHAR(255) NOT NULL,
    org_id     BIGINT       NOT NULL,
    status     INT          NOT NULL DEFAULT 1,
    extra      JSONB        NULL,
    version    INT          NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id)
);
COMMENT ON TABLE admin_user IS '用户表';
COMMENT ON COLUMN admin_user.id IS '主键ID';
COMMENT ON COLUMN admin_user.uid IS '全局唯一用户ID，最多10位数字';
COMMENT ON COLUMN admin_user.username IS '用户名';
COMMENT ON COLUMN admin_user.nickname IS '昵称';
COMMENT ON COLUMN admin_user.phone_num IS '手机号';
COMMENT ON COLUMN admin_user.email IS '邮箱';
COMMENT ON COLUMN admin_user.password IS '密码哈希（bcrypt）';
COMMENT ON COLUMN admin_user.org_id IS '所属组织业务ID';
COMMENT ON COLUMN admin_user.status IS '状态：1正常，2禁用';
COMMENT ON COLUMN admin_user.extra IS '扩展信息';
COMMENT ON COLUMN admin_user.version IS '乐观锁版本号';
COMMENT ON COLUMN admin_user.created_at IS '创建时间';
COMMENT ON COLUMN admin_user.updated_at IS '更新时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_admin_user_uid ON admin_user (uid);
CREATE UNIQUE INDEX IF NOT EXISTS uk_admin_user_username ON admin_user (username);
CREATE INDEX IF NOT EXISTS idx_admin_user_org_id ON admin_user (org_id);
CREATE INDEX IF NOT EXISTS idx_admin_user_org_updated_at ON admin_user (org_id, updated_at);

CREATE TABLE IF NOT EXISTS permission (
    id          BIGSERIAL    NOT NULL,
    code        VARCHAR(100) NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    org_id      BIGINT       NOT NULL,
    extra       JSONB        NULL,
    version     INT          NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id)
);
COMMENT ON TABLE permission IS '权限表';
COMMENT ON COLUMN permission.id IS '主键ID';
COMMENT ON COLUMN permission.code IS '权限业务编码，形如 report.read';
COMMENT ON COLUMN permission.name IS '权限名称';
COMMENT ON COLUMN permission.description IS '权限描述';
COMMENT ON COLUMN permission.org_id IS '所属组织业务ID';
COMMENT ON COLUMN permission.extra IS '扩展信息';
COMMENT ON COLUMN permission.version IS '乐观锁版本号';
COMMENT ON COLUMN permission.created_at IS '创建时间';
COMMENT ON COLUMN permission.updated_at IS '更新时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_permission_org_code ON permission (org_id, code);
CREATE UNIQUE INDEX IF NOT EXISTS uk_permission_org_name ON permission (org_id, name);
CREATE INDEX IF NOT EXISTS idx_permission_org_id ON permission (org_id);
CREATE INDEX IF NOT EXISTS idx_permission_org_updated_at ON permission (org_id, updated_at);

CREATE TABLE IF NOT EXISTS role (
    id          BIGSERIAL    NOT NULL,
    role_id     BIGINT       NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    org_id      BIGINT       NOT NULL,
    extra       JSONB        NULL,
    version     INT          NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id)
);
COMMENT ON TABLE role IS '角色表';
COMMENT ON COLUMN role.id IS '主键ID';
COMMENT ON COLUMN role.role_id IS '角色业务ID，全局唯一；负数表示用户个人直接权限角色';
COMMENT ON COLUMN role.name IS '角色名称';
COMMENT ON COLUMN role.description IS '角色描述';
COMMENT ON COLUMN role.org_id IS '所属组织业务ID';
COMMENT ON COLUMN role.extra IS '扩展信息';
COMMENT ON COLUMN role.version IS '乐观锁版本号';
COMMENT ON COLUMN role.created_at IS '创建时间';
COMMENT ON COLUMN role.updated_at IS '更新时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_role_role_id ON role (role_id);
CREATE UNIQUE INDEX IF NOT EXISTS uk_role_org_name ON role (org_id, name);
CREATE INDEX IF NOT EXISTS idx_role_org_id ON role (org_id);
CREATE INDEX IF NOT EXISTS idx_role_org_updated_at ON role (org_id, updated_at);

CREATE TABLE IF NOT EXISTS user_role (
    id         BIGSERIAL    NOT NULL,
    uid        BIGINT       NOT NULL,
    role_id    BIGINT       NOT NULL,
    extra      JSONB        NULL,
    version    INT          NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id)
);
COMMENT ON TABLE user_role IS '用户角色关联表';
COMMENT ON COLUMN user_role.id IS '主键ID';
COMMENT ON COLUMN user_role.uid IS '用户业务ID';
COMMENT ON COLUMN user_role.role_id IS '角色业务ID';
COMMENT ON COLUMN user_role.extra IS '扩展信息';
COMMENT ON COLUMN user_role.version IS '乐观锁版本号';
COMMENT ON COLUMN user_role.created_at IS '创建时间';
COMMENT ON COLUMN user_role.updated_at IS '更新时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_user_role_uid_role_id ON user_role (uid, role_id);
CREATE INDEX IF NOT EXISTS idx_user_role_role_id ON user_role (role_id);

CREATE TABLE IF NOT EXISTS role_permission (
    id              BIGSERIAL    NOT NULL,
    role_id         BIGINT       NOT NULL,
    permission_code VARCHAR(100) NOT NULL,
    extra           JSONB        NULL,
    version         INT          NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at      TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id)
);
COMMENT ON TABLE role_permission IS '角色权限关联表';
COMMENT ON COLUMN role_permission.id IS '主键ID';
COMMENT ON COLUMN role_permission.role_id IS '角色业务ID';
COMMENT ON COLUMN role_permission.permission_code IS '权限编码，对应permission.code';
COMMENT ON COLUMN role_permission.extra IS '扩展信息';
COMMENT ON COLUMN role_permission.version IS '乐观锁版本号';
COMMENT ON COLUMN role_permission.created_at IS '创建时间';
COMMENT ON COLUMN role_permission.updated_at IS '更新时间';

CREATE UNIQUE INDEX IF NOT EXISTS uk_role_permission_role_code
    ON role_permission (role_id, permission_code);
CREATE INDEX IF NOT EXISTS idx_role_permission_code ON role_permission (permission_code);

-- updated_at 自动刷新，等价于 MySQL 的 ON UPDATE CURRENT_TIMESTAMP(3)。
DO $$
DECLARE
    target TEXT;
BEGIN
    FOREACH target IN ARRAY ARRAY[
        'organization', 'admin_user', 'permission', 'role', 'user_role', 'role_permission'
    ]
    LOOP
        EXECUTE FORMAT('DROP TRIGGER IF EXISTS trg_%s_updated_at ON %I', target, target);
        EXECUTE FORMAT(
            'CREATE TRIGGER trg_%s_updated_at BEFORE UPDATE ON %I '
            'FOR EACH ROW EXECUTE FUNCTION pyadmin_set_updated_at()',
            target, target
        );
    END LOOP;
END;
$$;
