CREATE TABLE IF NOT EXISTS `ops_scan_plans` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`repoID` int unsigned NOT NULL COMMENT '仓库ID',
	`name` varchar(255) NOT NULL COMMENT '扫描计划名称',
	`desc` text NULL DEFAULT NULL COMMENT '扫描计划描述',
	`scanType` varchar(20) NOT NULL DEFAULT 'full' COMMENT '扫描类型（full/incremental）',
	`branches` text NULL DEFAULT NULL COMMENT '扫描分支范围',
	`files` text NULL DEFAULT NULL COMMENT '扫描文件范围',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='扫描计划表';

CREATE UNIQUE INDEX uk_name_repo_deleted ON ops_scan_plans (`repoID`, `name`, `deleted`);
CREATE INDEX idx_repoID_deleted ON ops_scan_plans (`repoID`, `deleted`);
CREATE INDEX idx_name ON ops_scan_plans (`name`);
CREATE INDEX idx_deleted ON ops_scan_plans (`deleted`);
