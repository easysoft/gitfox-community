CREATE TABLE IF NOT EXISTS `ops_space` (
  `id` smallint UNSIGNED NOT NULL  COMMENT 'ID' AUTO_INCREMENT,
  `name` varchar(200) NOT NULL COMMENT '名称',
  `owner` varchar(30) NOT NULL DEFAULT '' COMMENT '负责人',
  `acl` varchar(30) NOT NULL DEFAULT 'open' COMMENT '权限:private,open',
  `desc` TEXT NULL DEFAULT NULL COMMENT '描述',
  `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
  `createdDate` datetime NULL COMMENT '创建时间',
  `editedBy` varchar(30) NOT NULL DEFAULT '',
  `editedDate` datetime NULL COMMENT '更新时间',
  `deleted` tinyint NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='空间表';