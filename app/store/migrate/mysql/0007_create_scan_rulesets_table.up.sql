CREATE TABLE IF NOT EXISTS `ops_scan_rulesets` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`name` varchar(255) NOT NULL COMMENT '规则集名称',
	`desc` text NULL DEFAULT NULL COMMENT '规则集描述（中文）',
	`descEn` varchar(500) NULL DEFAULT NULL COMMENT '规则集描述（英文）',
	`type` varchar(20) NOT NULL COMMENT '规则集类型（如：代码扫描、依赖扫描、全量扫描等）',
	`lang` varchar(50) NULL DEFAULT NULL COMMENT '适配语言类型（java、python，NULL表示适配所有语言）',
	`plugin` varchar(50) NOT NULL COMMENT '所属插件名称（标识规则集归属的扫描插件）',
	`tag` varchar(50) NOT NULL COMMENT '规则集标签（用于分类筛选，如：安全合规、代码规范、性能优化等）',
	`isCustom` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否自定义规则集（1=自定义，0=系统默认规则集）',
	`status` varchar(20) NOT NULL DEFAULT 1 COMMENT '规则状态（enabled-启用、disabled-禁用，默认启用）',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='扫描规则集表';
CREATE UNIQUE INDEX uk_name_lang_deleted_unique ON ops_scan_rulesets (name, lang, deleted);
CREATE INDEX idx_name ON ops_scan_rulesets (name);
CREATE INDEX idx_type ON ops_scan_rulesets (type);
CREATE INDEX idx_status ON ops_scan_rulesets (status);
CREATE INDEX idx_plugin_tag ON ops_scan_rulesets (plugin, tag);
CREATE INDEX idx_deleted ON ops_scan_rulesets (deleted);
