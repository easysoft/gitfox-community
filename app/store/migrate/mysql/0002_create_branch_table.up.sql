-- 创建代码库分支表
CREATE TABLE IF NOT EXISTS `ops_branch` (
  `id` VARCHAR(64) NOT NULL COMMENT '分支唯一标识',
  `name` VARCHAR(255) NOT NULL COMMENT '分支名称',
  `repo_id` VARCHAR(64) NOT NULL COMMENT '所属代码库ID',
  `creator_id` VARCHAR(64) NOT NULL COMMENT '创建者ID',
  `source` VARCHAR(32) NOT NULL DEFAULT 'manual' COMMENT '来源: manual(手动), auto(自动), webhook(Webhook触发)',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `protection` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否受保护: 0-否, 1-是',
  PRIMARY KEY (`id`),
  INDEX `idx_repo_id` (`repo_id`),
  INDEX `idx_creator_id` (`creator_id`),
  INDEX `idx_name` (`name`),
  INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='代码库分支表';
