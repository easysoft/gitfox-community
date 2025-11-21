CREATE TABLE IF NOT EXISTS `ops_public_keys` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `account` varchar(30) NOT NULL COMMENT '关联用户账号',
  `verified` datetime NULL COMMENT '公钥验证时间',
  `identifier` varchar(255) NOT NULL COMMENT '公钥唯一标识',
  `fingerprint` varchar(255) NOT NULL COMMENT '公钥指纹',
  `content` text NOT NULL COMMENT '公钥原始内容',
  `comment` varchar(500) NOT NULL DEFAULT '' COMMENT '公钥注释部分',
  `type` varchar(50) NOT NULL COMMENT '公钥算法类型',
  `created_by` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
  `created_date` datetime NULL COMMENT '创建时间',
  `edited_by` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
  `edited_date` datetime NULL COMMENT '更新时间',
  `deleted` tinyint NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uk_fingerprint` (`fingerprint`),
  UNIQUE KEY `uk_identifier` (`identifier`),
  KEY `idx_deleted` (`deleted`)
) ENGINE=InnoDB COMMENT='SSH公钥表';
