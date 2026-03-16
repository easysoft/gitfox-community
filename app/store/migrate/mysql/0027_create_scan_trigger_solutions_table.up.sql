CREATE TABLE IF NOT EXISTS `ops_scan_trigger_solutions` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`triggerID` int unsigned NOT NULL COMMENT '触发器ID',
	`solutionID` int unsigned NOT NULL COMMENT '解决方案ID',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	PRIMARY KEY (`id`) USING BTREE,
	UNIQUE INDEX `uk_trigger_solution` (`triggerID`, `solutionID`)
)
ENGINE=InnoDB COMMENT='扫描触发器关联解决方案表';

CREATE INDEX idx_triggerID ON ops_scan_trigger_solutions (`triggerID`);
