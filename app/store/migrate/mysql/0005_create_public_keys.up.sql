CREATE TABLE IF NOT EXISTS `ops_public_keys` (
  `id` int unsigned NOT NULL  COMMENT 'ID' AUTO_INCREMENT,
  `account` varchar(30) NOT NULL COMMENT '关联用户账号',
  `identifier` varchar(255) NOT NULL COMMENT '公钥唯一标识',
  `verifiedDate` datetime NULL COMMENT '公钥验证时间',
  `fingerprint` varchar(255) NOT NULL COMMENT '公钥指纹',
  `content` text NOT NULL COMMENT '公钥原始内容',
  `comment` varchar(500) NOT NULL DEFAULT '' COMMENT '公钥注释部分',
  `type` varchar(50) NOT NULL COMMENT '公钥算法类型',
  `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
  `createdDate` datetime NULL COMMENT '创建时间',
  `editedBy` varchar(30) NOT NULL DEFAULT '',
  `editedDate` datetime NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='SSH公钥表';

CREATE UNIQUE INDEX `uk_fingerprint` ON `ops_public_keys` (`fingerprint`);
CREATE UNIQUE INDEX `uk_identifier_account` ON `ops_public_keys` (`identifier`, `account`);
