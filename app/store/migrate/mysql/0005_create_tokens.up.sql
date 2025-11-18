CREATE TABLE IF NOT EXISTS `ops_tokens` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `principal_id` bigint unsigned NOT NULL COMMENT '关联用户ID',
  `type` varchar(50) NOT NULL COMMENT 'Token类型: session, pat',
  `identifier` varchar(255) NOT NULL COMMENT 'Token唯一标识',
  `expires_at` datetime NULL COMMENT '过期时间(毫秒 Unix 时间)',
  `created_by` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
  `created_date` datetime NULL COMMENT '创建时间',
  `edited_by` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
  `edited_date` datetime NULL COMMENT '更新时间',
  `deleted` tinyint NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_identifier` (`identifier`),
  KEY `idx_principal_id` (`principal_id`),
  KEY `idx_deleted` (`deleted`)
) ENGINE=InnoDB COMMENT='令牌表';