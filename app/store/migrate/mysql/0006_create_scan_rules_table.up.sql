CREATE TABLE IF NOT EXISTS `ops_scan_rules` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`ruleKey` varchar(255) NOT NULL COMMENT '规则唯一标识',
	`name` varchar(255) NOT NULL COMMENT '规则名称',
	`desc` text NULL DEFAULT NULL COMMENT '规则描述（中文）',
	`descEn` text NULL DEFAULT NULL COMMENT '规则描述（英文）',
	`lang` varchar(50) NOT NULL COMMENT '语言类型（如：java、python等）',
	`plugin` varchar(50) NOT NULL COMMENT '所属插件名称（标识规则归属的扫描插件）',
	`tag` varchar(50) NOT NULL COMMENT '规则标签（用于分类筛选，如：安全、规范、性能等）',
	`priority` varchar(20) NOT NULL  COMMENT '规则优先级（low-低、medium-中、high-高）',
	`type` varchar(20) NOT NULL COMMENT '规则类型（如：defect(缺陷)、security(安全)、compliance(合规)、optimize(优化)等）',
	`content` text NOT NULL COMMENT '规则内容（中文，如扫描正则、检查逻辑描述等）',
	`contentEn` text NOT NULL COMMENT '规则内容（英文，如扫描正则、检查逻辑描述等）',
	`isCustom` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否自定义规则（1=自定义规则，0=系统默认规则）',
	`status` varchar(20) NOT NULL DEFAULT 'enabled' COMMENT '规则状态（enabled-启用、disabled-禁用，默认启用）',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='扫描规则表';

CREATE UNIQUE INDEX uk_ruleKey ON ops_scan_rules (ruleKey);
CREATE UNIQUE INDEX uk_ruleKey_lang_deleted_unique ON ops_scan_rules (ruleKey, lang, deleted);
CREATE INDEX idx_name ON ops_scan_rules (name);
CREATE INDEX idx_type ON ops_scan_rules (type);
CREATE INDEX idx_plugin_tag ON ops_scan_rules (plugin, tag);
CREATE INDEX idx_priority_status ON ops_scan_rules (priority, status);
CREATE INDEX idx_deleted ON ops_scan_rules (deleted);
CREATE INDEX idx_lang ON ops_scan_rules (lang);

