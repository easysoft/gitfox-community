CREATE TABLE IF NOT EXISTS `ops_tokens` (
	`id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'Token ID',
	`principal_id` bigint unsigned NOT NULL COMMENT '关联用户ID',
	`type` varchar(30) NOT NULL COMMENT '类型: session, pat',
	`identifier` varchar(255) NOT NULL COMMENT '令牌唯一标识',
	`expires_at` datetime NULL COMMENT '过期时间',
	`created_by` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`created_date` datetime NULL COMMENT '创建时间',
	`edited_by` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`edited_date` datetime NULL COMMENT '更新时间',
	`deleted` tinyint NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB COMMENT='令牌表';

CREATE UNIQUE INDEX `uk_token_identifier` ON `ops_tokens` (`identifier`);
CREATE INDEX `idx_principal_id` ON `ops_tokens` (`principal_id`);
CREATE INDEX `idx_deleted` ON `ops_tokens` (`deleted`);