CREATE TABLE IF NOT EXISTS `ops_tokens` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `account` varchar(30) NOT NULL COMMENT '关联用户账号',
  `type` varchar(50) NOT NULL COMMENT 'Token类型: session, pat',
  `identifier` varchar(255) NOT NULL COMMENT 'Token唯一标识',
  `expires_date` datetime NULL COMMENT '过期时间',
  `token` varchar(255) NOT NULL COMMENT 'Token值',
  `created_by` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
  `created_date` datetime NULL COMMENT '创建时间',
  `edited_by` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
  `edited_date` datetime NULL COMMENT '更新时间',
  `deleted` tinyint NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_identifier` (`identifier`),
  KEY `idx_deleted` (`deleted`)
) ENGINE=InnoDB COMMENT='令牌表';