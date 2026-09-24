CREATE DATABASE IF NOT EXISTS `hercules_admin`
  DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci;
USE `hercules_admin`;

CREATE TABLE IF NOT EXISTS `organization` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `org_id` BIGINT NOT NULL COMMENT '组织业务ID，最多8位数字',
  `name` VARCHAR(100) NOT NULL COMMENT '组织名称',
  `parent_id` BIGINT NULL COMMENT '直接上级组织主键ID，无上级时为空',
  `description` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '组织描述',
  `status` INT NOT NULL DEFAULT 1 COMMENT '状态：1正常，2禁用',
  `extra` JSON NULL COMMENT '扩展信息',
  `version` INT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
    ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_organization_org_id` (`org_id`),
  KEY `idx_organization_parent_id` (`parent_id`),
  KEY `idx_organization_name` (`name`),
  KEY `idx_organization_updated_at` (`updated_at`)
) ENGINE=InnoDB COMMENT='组织表';

CREATE TABLE IF NOT EXISTS `user` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `uid` BIGINT NOT NULL COMMENT '全局唯一用户ID，最多10位数字',
  `username` VARCHAR(64) NOT NULL COMMENT '用户名',
  `nickname` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '昵称',
  `phone_num` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '手机号',
  `email` VARCHAR(254) NOT NULL DEFAULT '' COMMENT '邮箱',
  `password` VARCHAR(255) NOT NULL COMMENT '密码哈希',
  `org_id` BIGINT NOT NULL COMMENT '所属组织业务ID',
  `status` INT NOT NULL DEFAULT 1 COMMENT '状态：1正常，2禁用',
  `extra` JSON NULL COMMENT '扩展信息',
  `version` INT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
    ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_uid` (`uid`),
  UNIQUE KEY `uk_user_username` (`username`),
  KEY `idx_user_org_id` (`org_id`),
  KEY `idx_user_org_updated_at` (`org_id`, `updated_at`)
) ENGINE=InnoDB COMMENT='用户表';

CREATE TABLE IF NOT EXISTS `permission` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `code` VARCHAR(100) NOT NULL COMMENT '权限业务编码',
  `name` VARCHAR(100) NOT NULL COMMENT '权限名称',
  `description` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '权限描述',
  `org_id` BIGINT NOT NULL COMMENT '所属组织业务ID',
  `extra` JSON NULL COMMENT '扩展信息',
  `version` INT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
    ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_permission_org_code` (`org_id`, `code`),
  UNIQUE KEY `uk_permission_org_name` (`org_id`, `name`),
  KEY `idx_permission_org_id` (`org_id`),
  KEY `idx_permission_org_updated_at` (`org_id`, `updated_at`)
) ENGINE=InnoDB COMMENT='权限表';

CREATE TABLE IF NOT EXISTS `role` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `role_id` BIGINT NOT NULL COMMENT '角色业务ID',
  `name` VARCHAR(100) NOT NULL COMMENT '角色名称',
  `description` VARCHAR(500) NOT NULL DEFAULT '' COMMENT '角色描述',
  `org_id` BIGINT NOT NULL COMMENT '所属组织业务ID',
  `extra` JSON NULL COMMENT '扩展信息',
  `version` INT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
    ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_role_id` (`role_id`),
  UNIQUE KEY `uk_role_org_name` (`org_id`, `name`),
  KEY `idx_role_org_id` (`org_id`),
  KEY `idx_role_org_updated_at` (`org_id`, `updated_at`)
) ENGINE=InnoDB COMMENT='角色表';

CREATE TABLE IF NOT EXISTS `user_role` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `uid` BIGINT NOT NULL COMMENT '用户业务ID',
  `role_id` BIGINT NOT NULL COMMENT '角色业务ID',
  `extra` JSON NULL COMMENT '扩展信息',
  `version` INT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
    ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_role_uid_role_id` (`uid`, `role_id`),
  KEY `idx_user_role_role_id` (`role_id`)
) ENGINE=InnoDB COMMENT='用户角色关联表';

CREATE TABLE IF NOT EXISTS `role_permission` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `role_id` BIGINT NOT NULL COMMENT '角色业务ID',
  `permission_code` VARCHAR(100) NOT NULL COMMENT '权限编码，对应permission.code',
  `extra` JSON NULL COMMENT '扩展信息',
  `version` INT NOT NULL DEFAULT 1 COMMENT '乐观锁版本号',
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3)
    ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_permission_role_code` (`role_id`, `permission_code`),
  KEY `idx_role_permission_code` (`permission_code`)
) ENGINE=InnoDB COMMENT='角色权限关联表';
