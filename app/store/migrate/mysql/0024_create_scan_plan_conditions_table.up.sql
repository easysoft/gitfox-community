CREATE TABLE IF NOT EXISTS `ops_scan_plan_conditions` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`planID` int unsigned NOT NULL COMMENT '扫描计划ID',
	`triggerID` int unsigned NOT NULL DEFAULT 0 COMMENT '触发器ID',
	`repoID` int unsigned NOT NULL COMMENT '仓库ID',
	`priority` tinyint(1) NOT NULL COMMENT '优先级（0-low,1-medium,2-high）',
	`type` varchar(50) NOT NULL COMMENT '条件类型（defect/security/compliance/optimize）',
	`unit` varchar(20) NOT NULL COMMENT '阈值单位（count/percent）',
	`threshold` bigint NOT NULL COMMENT '阈值',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='扫描计划触发条件表';

CREATE INDEX idx_triggerID_deleted ON ops_scan_plan_conditions (`triggerID`, `deleted`);
CREATE INDEX idx_repo_plan_trigger_deleted ON ops_scan_plan_conditions (`repoID`, `planID`, `triggerID`, `deleted`);
CREATE INDEX idx_deleted ON ops_scan_plan_conditions (`deleted`);
