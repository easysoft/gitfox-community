CREATE TABLE IF NOT EXISTS `ops_triggers` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '触发器ID，自增主键',
	`pipelineID` int unsigned NOT NULL DEFAULT 0 COMMENT '关联的流水线ID',
	`trigger` varchar(500) NOT NULL DEFAULT '' COMMENT '触发动作',
	`cron` varchar(255) NOT NULL DEFAULT '' COMMENT '定时配置',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='流水线触发器表';

CREATE INDEX `idx_trigger_pipelineID` ON `ops_triggers` (`pipelineID`);
