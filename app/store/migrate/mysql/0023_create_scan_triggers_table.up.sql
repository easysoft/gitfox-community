CREATE TABLE IF NOT EXISTS `ops_scan_triggers` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`planID` int unsigned NOT NULL COMMENT '扫描计划ID',
	`repoID` int unsigned NOT NULL COMMENT '仓库ID',
	`name` varchar(255) NOT NULL COMMENT '触发器名称',
	`desc` text NULL DEFAULT NULL COMMENT '触发器描述',
	`type` varchar(20) NOT NULL COMMENT '触发器类型（cron/manual/event）',
	`cron` varchar(255) NOT NULL DEFAULT '' COMMENT 'cron 表达式',
	`cronBranch` varchar(255) NOT NULL DEFAULT '' COMMENT 'cron 扫描分支',
	`keywords` varchar(500) NOT NULL DEFAULT '' COMMENT '关键字过滤条件',
	`actions` varchar(500) NOT NULL DEFAULT '' COMMENT '触发动作配置',
	`disabled` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否禁用，0-启用，1-禁用',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='扫描计划触发器表';

CREATE INDEX idx_planID ON ops_scan_triggers (`planID`);
CREATE INDEX idx_repo_plan_deleted ON ops_scan_triggers (`repoID`, `planID`, `deleted`);
CREATE INDEX idx_deleted ON ops_scan_triggers (`deleted`);
