CREATE TABLE IF NOT EXISTS `ops_scan_ruleset_rules` (
	`id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`rulesetID` INT NOT NULL COMMENT '规则集主键ID',
	`ruleID` INT NOT NULL COMMENT '规则主键ID',
    `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '',
	`editedDate` datetime NULL COMMENT '更新时间',
    `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE,
)
ENGINE=InnoDB COMMENT="扫描集与扫描规则关联表";
CREATE UNIQUE INDEX uk_ruleset_rules ON ops_scan_ruleset_rules (rulesetID, ruleID, deleted);
CREATE INDEX idx_rulesetID ON ops_scan_ruleset_rules (rulesetID);
CREATE INDEX idx_ruleID ON ops_scan_ruleset_rules (ruleID);
CREATE INDEX idx_deleted ON ops_scan_ruleset_rules (deleted);
CREATE INDEX idx_ruleset_rules ON ops_scan_ruleset_rules (rulesetID, ruleID, deleted);


