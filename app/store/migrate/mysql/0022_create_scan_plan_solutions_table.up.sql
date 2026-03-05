CREATE TABLE IF NOT EXISTS `ops_scan_plan_solutions` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`planID` int unsigned NOT NULL COMMENT '扫描计划ID',
	`solutionID` int unsigned NOT NULL COMMENT '扫描方案ID',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='扫描计划与扫描方案关联表';

CREATE UNIQUE INDEX uk_plan_solution_deleted ON ops_scan_plan_solutions (`planID`, `solutionID`, `deleted`);
CREATE INDEX idx_planID ON ops_scan_plan_solutions (`planID`);
CREATE INDEX idx_solutionID ON ops_scan_plan_solutions (`solutionID`);
CREATE INDEX idx_deleted ON ops_scan_plan_solutions (`deleted`);
