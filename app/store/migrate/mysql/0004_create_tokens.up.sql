CREATE TABLE IF NOT EXISTS `ops_tokens` (
  `id` int unsigned NOT NULL  COMMENT 'ID' AUTO_INCREMENT,
  `account` varchar(30) NOT NULL COMMENT '关联用户账号',
  `identifier` varchar(255) NOT NULL COMMENT 'Token唯一标识',
  `type` varchar(50) NOT NULL COMMENT 'Token类型: pat,app',
  `expirationDate` datetime NULL COMMENT '过期时间',
  `token` text NOT NULL COMMENT 'Token值',
  `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
  `createdDate` datetime NULL COMMENT '创建时间',
  `editedBy` varchar(30) NOT NULL DEFAULT '',
  `editedDate` datetime NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='令牌表';

CREATE UNIQUE INDEX `uk_identifier_account` ON `ops_tokens` (`identifier`, `account`);