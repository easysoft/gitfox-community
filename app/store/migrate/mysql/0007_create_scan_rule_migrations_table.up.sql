CREATE TABLE IF NOT EXISTS `ops_scan_rule_migrations` (
	`name` VARCHAR(255) NOT NULL COMMENT '迁移名称',
	`version` VARCHAR(255) NOT NULL COMMENT '迁移版本',
    `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
    `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
)
ENGINE=InnoDB COMMENT='扫描规则迁移记录表'
;
