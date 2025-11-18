CREATE TABLE IF NOT EXISTS `ops_scan_solution_rulesets` (
	`id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`solutionID` INT NOT NULL COMMENT '扫描方案主键ID',
	`rulesetID` INT NOT NULL  COMMENT '扫描规则集主键ID',
    `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '',
	`editedDate` datetime NULL COMMENT '更新时间',
    `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE,
)
ENGINE=InnoDB COMMENT='扫描方案与规则集关联表';
CREATE UNIQUE INDEX uk_ruleset_rules ON ops_scan_solution_rulesets (solutionID, rulesetID, deleted);
CREATE INDEX idx_rulesetID ON ops_scan_solution_rulesets (rulesetID);
CREATE INDEX idx_solutionID ON ops_scan_solution_rulesets (solutionID);
CREATE INDEX idx_deleted ON ops_scan_solution_rulesets (deleted);