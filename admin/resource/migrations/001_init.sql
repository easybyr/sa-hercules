CREATE DATABASE IF NOT EXISTS `hercules_admin`
  DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
USE `hercules_admin`;

SET NAMES utf8mb4;

-- 1. 租户表
CREATE TABLE IF NOT EXISTS `sys_tenant` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `code` VARCHAR(64) NOT NULL COMMENT '租户编码，全局唯一',
  `parent_code` VARCHAR(64) NULL COMMENT '直接上级租户，无上级时为空',
  `name` VARCHAR(128) NOT NULL COMMENT '租户名称',
  `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
  `description` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '描述',
  `extra` JSON NULL COMMENT '扩展信息',
  `version` INT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenant_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户';

-- 2. 全局用户表：不带 tenant_id
CREATE TABLE IF NOT EXISTS `sys_user` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `uid` BIGINT NOT NULL COMMENT '全局唯一用户ID',
  `username` VARCHAR(64) NOT NULL COMMENT '全局登录名',
  `nickname` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '昵称',
  `password_hash` VARCHAR(255) NOT NULL,
  `email` VARCHAR(128) NULL COMMENT '邮箱',
  `phone` VARCHAR(32) NULL COMMENT '手机号',
  `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
  `extra` JSON NULL COMMENT '扩展信息',
  `version` INT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_uid` (`uid`),
  UNIQUE KEY `uk_user_username` (`username`),
  UNIQUE KEY `uk_user_email` (`email`),
  UNIQUE KEY `uk_user_phone` (`phone`),
  KEY `idx_user_updated_at` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='全局用户';

-- 3. 租户成员表：维护用户与租户的成员关系
CREATE TABLE IF NOT EXISTS `sys_tenant_member` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT UNSIGNED NOT NULL,
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户主键ID',
  `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1正常 0禁用',
  `joined_at` DATETIME(3) NULL COMMENT '加入租户时间',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tm_tenant_user` (`tenant_id`, `user_id`),
  UNIQUE KEY `uk_tm_tenant_id_id_user` (`tenant_id`, `id`, `user_id`),
  KEY `idx_tm_user` (`user_id`),
  KEY `idx_tm_tenant_status` (`tenant_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户成员';

-- 4. 租户级角色表
CREATE TABLE IF NOT EXISTS `sys_role` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT UNSIGNED NOT NULL,
  `code` VARCHAR(64) NOT NULL COMMENT '租户内角色编码',
  `name` VARCHAR(128) NOT NULL COMMENT '角色名称',
  `description` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '角色描述',
  `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
  `is_system` TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '是否系统内置角色',
  `extra` JSON NULL COMMENT '扩展信息',
  `version` INT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_tenant_code` (`tenant_id`, `code`),
  UNIQUE KEY `uk_role_tenant_id_id` (`tenant_id`, `id`),
  KEY `idx_role_tenant_status` (`tenant_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户级角色';

-- 5. 租户级权限定义表
CREATE TABLE IF NOT EXISTS `sys_permission` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT UNSIGNED NOT NULL,
  `code` VARCHAR(128) NOT NULL COMMENT '租户内权限码，例如 order.read',
  `name` VARCHAR(128) NOT NULL,
  `node_type` VARCHAR(32) NOT NULL COMMENT 'MENU/BUTTON/API/DATA',
  `parent_id` BIGINT UNSIGNED NULL COMMENT '父权限，可为空',
  `description` VARCHAR(255) NULL,
  `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
  `extra` JSON NULL COMMENT '扩展信息',
  `version` INT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_perm_tenant_code` (`tenant_id`, `code`),
  UNIQUE KEY `uk_perm_tenant_id_id` (`tenant_id`, `id`),
  KEY `idx_perm_tenant_parent` (`tenant_id`, `parent_id`),
  KEY `idx_perm_tenant_type_status` (`tenant_id`, `node_type`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户级权限定义';

-- 6. 角色-权限关系：带 tenant_id
CREATE TABLE IF NOT EXISTS `sys_role_permission` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT UNSIGNED NOT NULL,
  `role_id` BIGINT UNSIGNED NOT NULL,
  `permission_id` BIGINT UNSIGNED NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_rp_tenant_role_perm` (`tenant_id`, `role_id`, `permission_id`),
  KEY `idx_rp_tenant_perm` (`tenant_id`, `permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户内角色权限关系';

-- 7. 用户角色表：语义上是“租户成员角色”
-- 不直接只存 user_id + role_id，而是绑定 tenant_member_id，避免角色跨租户
CREATE TABLE IF NOT EXISTS `sys_user_role` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT UNSIGNED NOT NULL,
  `tenant_member_id` BIGINT UNSIGNED NOT NULL COMMENT '租户成员表主键ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '冗余，便于查询，用户主键ID',
  `role_id` BIGINT UNSIGNED NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_ur_tenant_member_role` (`tenant_id`, `tenant_member_id`, `role_id`),
  KEY `idx_ur_tenant_member` (`tenant_id`, `tenant_member_id`),
  KEY `idx_ur_tenant_role` (`tenant_id`, `role_id`),
  KEY `idx_ur_user` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户成员角色关系';

-- 8. 全局 API 权限映射表：不带 tenant_id
-- 只负责把 permission_code 映射到具体接口，不直接关联 permission 表
CREATE TABLE IF NOT EXISTS `sys_api_permission` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `permission_code` VARCHAR(128) NOT NULL COMMENT '对应某个租户 permission.code，不做全局外键',
  `service_name` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '服务名，单服务可留空',
  `http_method` VARCHAR(10) NOT NULL COMMENT 'GET/POST/PUT/DELETE/PATCH/*',
  `api_path` VARCHAR(512) NOT NULL COMMENT '接口路径模板，如 /api/orders/{id}',
  `path_type` VARCHAR(16) NOT NULL DEFAULT 'EXACT' COMMENT 'EXACT/PREFIX/REGEX',
  `description` VARCHAR(255) NULL COMMENT '描述',
  `status` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '1启用 0禁用',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_api_service_method_path` (`service_name`, `http_method`, `api_path`),
  KEY `idx_sys_api_permission_code` (`permission_code`),
  KEY `idx_api_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='全局API权限映射';

CREATE TABLE IF NOT EXISTS `sys_permission_oper_log` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `tenant_id` BIGINT NOT NULL COMMENT '租户ID',
  `operator_uid` BIGINT NOT NULL COMMENT '操作人用户业务ID',
  `target_type` VARCHAR(32) NOT NULL COMMENT '操作对象类型：permission、role、user',
  `target_id` BIGINT NOT NULL COMMENT '操作对象业务ID',
  `operation` VARCHAR(64) NOT NULL COMMENT '操作类型，例如：create、update、delete',
  `before_data` JSON NOT NULL COMMENT '变更前数据',
  `after_data` JSON NOT NULL COMMENT '变更后数据',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '操作时间',
  PRIMARY KEY (`id`),
  KEY `idx_permission_operation_tenant_created_at` (`tenant_id`, `created_at`),
  KEY `idx_permission_operation_target` (`target_type`, `target_id`, `created_at`),
  KEY `idx_permission_operation_operator` (`operator_uid`, `created_at`)
) ENGINE=InnoDB COMMENT='权限操作日志表';
