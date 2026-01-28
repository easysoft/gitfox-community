CREATE TABLE IF NOT EXISTS `ops_schedule_jobs` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '任务ID，自增主键',
	`name` varchar(255) NOT NULL COMMENT '任务唯一标识',
	`type` varchar(255) NOT NULL COMMENT '任务类型',
	`priority` int NOT NULL DEFAULT 0 COMMENT '任务优先级',
	`data` text NOT NULL COMMENT '任务数据（JSON格式）',
	`result` text NOT NULL COMMENT '任务结果（JSON格式）',
	`maxDurationSeconds` int NOT NULL DEFAULT 0 COMMENT '最大执行时长（秒）',
	`maxRetries` int NOT NULL DEFAULT 0 COMMENT '最大重试次数',
	`state` varchar(255) NOT NULL COMMENT '任务状态',
	`scheduled` datetime NULL COMMENT '计划执行时间',
	`totalExecutions` int NOT NULL DEFAULT 0 COMMENT '总执行次数',
	`runBy` varchar(255) NOT NULL DEFAULT '' COMMENT '执行者',
	`runDeadline` datetime NULL COMMENT '执行截止时间',
	`runProgress` int NOT NULL DEFAULT 0 COMMENT '执行进度',
	`lastExecuted` datetime NULL COMMENT '最后执行时间',
	`isRecurring` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否为周期性任务，0-否，1-是',
	`recurringCron` varchar(255) NOT NULL DEFAULT '' COMMENT '周期性任务Cron表达式',
	`consecutiveFailures` int NOT NULL DEFAULT 0 COMMENT '连续失败次数',
	`lastFailureError` text NOT NULL COMMENT '最后失败错误信息',
	`groupID` varchar(255) NOT NULL DEFAULT '' COMMENT '任务组ID',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='后台任务表';

CREATE UNIQUE INDEX `uk_job_name` ON `ops_schedule_jobs` (`name`);
CREATE INDEX `idx_job_type` ON `ops_schedule_jobs` (`type`);
CREATE INDEX `idx_job_state` ON `ops_schedule_jobs` (`state`);
CREATE INDEX `idx_job_groupID` ON `ops_schedule_jobs` (`groupID`);
