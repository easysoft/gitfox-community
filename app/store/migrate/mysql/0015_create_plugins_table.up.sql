CREATE TABLE IF NOT EXISTS `ops_plugins` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '插件ID，自增主键',
	`groupID` int NOT NULL COMMENT '插件分组ID',
	`name` varchar(255) NOT NULL COMMENT '插件名称（唯一标识符）',
	`desc` mediumtext NOT NULL COMMENT '插件描述信息',
	`logo` mediumtext NOT NULL COMMENT '插件Logo（SVG格式）',
	`yaml` longtext NOT NULL COMMENT '插件规范配置（YAML格式）',
	`json` longtext NOT NULL COMMENT '插件规范配置（JSON格式）',
	`type` varchar(255) NOT NULL DEFAULT ''  COMMENT '插件类型：step/stage',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='CI插件表';

CREATE UNIQUE INDEX `uk_plugin_name` ON `ops_plugins` (`name`);
CREATE INDEX `idx_plugin_type` ON `ops_plugins` (`type`);
