CREATE TABLE IF NOT EXISTS `ops_space` (
  `id` int unsigned NOT NULL  COMMENT 'ID' AUTO_INCREMENT,
  `name` varchar(200) NOT NULL COMMENT '名称',
  `owner` varchar(30) NOT NULL DEFAULT '' COMMENT '负责人',
  `acl` varchar(30) NOT NULL DEFAULT 'open' COMMENT '权限:private,open',
  `desc` varchar(500) NOT NULL DEFAULT '' COMMENT '描述',
  `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
  `createdDate` datetime NULL COMMENT '创建时间',
  `editedBy` varchar(30) NOT NULL DEFAULT '',
  `editedDate` datetime NULL COMMENT '更新时间',
  `deleted` tinyint NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='空间表';
CREATE UNIQUE INDEX `uk_name` ON `ops_space`(`name`); -- 单独创建唯一索引


CREATE TABLE IF NOT EXISTS `ops_spaceuser` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '编号',
  `space` int unsigned NOT NULL DEFAULT 0 COMMENT '所属空间',
  `role` varchar(10) NOT NULL DEFAULT '' COMMENT '角色',
  `account` varchar(30) NOT NULL DEFAULT '' COMMENT '用户帐号',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='空间用户关联表';
CREATE UNIQUE INDEX `uk_spaceuser` ON `ops_spaceuser` (`space`,`account`);