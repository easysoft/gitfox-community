CREATE TABLE IF NOT EXISTS `ops_scan_issues` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`issueKey` varchar(255) NOT NULL COMMENT '问题唯一标识（SHA1）',
	`ruleID` int unsigned NOT NULL COMMENT '关联规则ID',
	`message` text NULL DEFAULT NULL COMMENT '问题描述信息',
	`path` varchar(500) NOT NULL DEFAULT '' COMMENT '文件路径',
	`line` bigint NOT NULL DEFAULT 0 COMMENT '行号',
	`repoID` int unsigned NOT NULL COMMENT '仓库ID',
	`repoBranch` varchar(255) NOT NULL DEFAULT '' COMMENT '仓库分支',
	`createdByTaskID` int unsigned NOT NULL DEFAULT 0 COMMENT '创建该问题的任务ID',
	`updatedByTaskID` int unsigned NOT NULL DEFAULT 0 COMMENT '最后更新该问题的任务ID',
	`status` varchar(20) NOT NULL DEFAULT 'wait' COMMENT '问题状态（wait/todo/solving/solved/closed/ignore）',
	`scanMethod` varchar(20) NOT NULL DEFAULT '' COMMENT '扫描方法（check/smell）',
	`payload` text NULL DEFAULT NULL COMMENT '扩展数据（JSON）',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='扫描问题表';

CREATE UNIQUE INDEX uk_issueKey ON ops_scan_issues (`issueKey`);
CREATE INDEX idx_repoID_repoBranch ON ops_scan_issues (`repoID`, `repoBranch`);
CREATE INDEX idx_ruleID ON ops_scan_issues (`ruleID`);
CREATE INDEX idx_status ON ops_scan_issues (`status`);
CREATE INDEX idx_scanMethod ON ops_scan_issues (`scanMethod`);
CREATE INDEX idx_deleted ON ops_scan_issues (`deleted`);

CREATE TABLE IF NOT EXISTS `ops_scan_issue_task_binds` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`taskID` int unsigned NOT NULL COMMENT '扫描任务ID',
	`issueID` int unsigned NOT NULL COMMENT '扫描问题ID',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='扫描问题与任务绑定关系表';

CREATE INDEX idx_taskID ON ops_scan_issue_task_binds (`taskID`);
CREATE INDEX idx_issueID ON ops_scan_issue_task_binds (`issueID`);
CREATE UNIQUE INDEX uk_taskID_issueID ON ops_scan_issue_task_binds (`taskID`, `issueID`);
