CREATE TABLE IF NOT EXISTS `ops_scan_tasks` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`planID` int unsigned NOT NULL COMMENT '扫描计划ID',
	`executionID` int unsigned NOT NULL COMMENT '执行ID',
	`repoNumber` bigint NOT NULL DEFAULT 0 COMMENT '仓库编号',
	`planNumber` bigint NOT NULL DEFAULT 0 COMMENT '计划编号',
	`triggerID` int unsigned NULL COMMENT '触发器ID',
	`triggerType` varchar(20) NOT NULL DEFAULT '' COMMENT '触发类型（manual/cron/action）',
	`status` varchar(20) NOT NULL DEFAULT '' COMMENT '任务状态（in_progress/success/failed）',
	`result` varchar(20) NULL DEFAULT NULL COMMENT '任务结果（pass/no-pass/-）',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedDate` datetime NULL COMMENT '更新时间',
	`scanType` varchar(20) NOT NULL DEFAULT 'full' COMMENT '扫描类型（full/incremental）',
	`issueNumber` bigint NOT NULL DEFAULT 0 COMMENT '问题编号',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='扫描任务表';

CREATE INDEX idx_planID_deleted ON ops_scan_tasks (`planID`, `deleted`);
CREATE INDEX idx_executionID ON ops_scan_tasks (`executionID`);
CREATE INDEX idx_status ON ops_scan_tasks (`status`);
CREATE INDEX idx_deleted ON ops_scan_tasks (`deleted`);