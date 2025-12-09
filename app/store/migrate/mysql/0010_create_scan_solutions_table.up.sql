CREATE TABLE IF NOT EXISTS `ops_scan_solutions` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`name` varchar(255) NOT NULL COMMENT '扫描解决方案名称',
	`desc` varchar(500) NULL DEFAULT NULL COMMENT '解决方案描述（中文）',
	`descEn` varchar(500) NULL DEFAULT NULL COMMENT '解决方案描述（英文）',
	`status` varchar(20) NOT NULL DEFAULT 'enabled' COMMENT '规则状态（enabled-启用、disabled-禁用，默认启用）',
	`tag` varchar(50) NULL DEFAULT NULL COMMENT '解决方案标签（用于分类筛选，如：安全加固、漏洞修复、合规整改等）',
	`isCustom` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否自定义解决方案（1=自定义，0=系统默认解决方案）',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='扫描方案表';

CREATE UNIQUE INDEX uk_name_deleted ON ops_scan_solutions (name,deleted);
CREATE INDEX idx_name ON ops_scan_solutions (name);
CREATE INDEX idx_status ON ops_scan_solutions (status);
CREATE INDEX idx_deleted ON ops_scan_solutions (deleted);