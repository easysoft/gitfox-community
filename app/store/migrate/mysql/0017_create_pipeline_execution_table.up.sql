CREATE TABLE IF NOT EXISTS `ops_pipeline_executions` (
	`id` int unsigned  NOT NULL AUTO_INCREMENT COMMENT '执行记录ID，自增主键',
	`pipelineID` int unsigned NOT NULL COMMENT '关联的流水线ID',
	`parent` int unsigned NOT NULL DEFAULT 0 COMMENT '父执行记录ID（用于子流水线/重试场景，无则为0）',
	`trigger` varchar(50) NOT NULL COMMENT '触发方式：manual,branch_updated',
	`commit` char(40) NOT NULL DEFAULT '' COMMENT 'CommitID',
	`ref` varchar(255) NOT NULL DEFAULT '' COMMENT '分支/tag/commit，如refs/heads/main,refs/tags/v1.0,commit',
	`params` text NOT NULL DEFAULT '' COMMENT '执行入参（JSON格式）',
	`started` datetime NULL COMMENT '执行开始时间',
	`finished` datetime NULL COMMENT '执行结束时间',
	`duration` int NOT NULL DEFAULT 0 COMMENT '执行时长(s)',
	`status` varchar(50) NOT NULL DEFAULT '' COMMENT '执行状态',
	`error` varchar(500) NOT NULL DEFAULT '' COMMENT '执行错误信息（无错误则为空字符串）',
	`createdBy` varchar(30) NOT NULL COMMENT '执行记录创建人',
	`createdDate` datetime NULL COMMENT '记录创建时间',
	`editedDate` datetime NULL COMMENT '记录更新时间',
	PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='流水线执行记录表';

CREATE INDEX `idx_createdBy` ON `ops_pipeline_executions` (`createdBy`);
CREATE INDEX `idx_status` ON `ops_pipeline_executions` (`status`);


CREATE TABLE IF NOT EXISTS `ops_stage_executions` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '阶段ID，自增主键',
	`repoID` int unsigned NOT NULL DEFAULT 0 COMMENT '关联的代码仓库ID,并发控制用',
	`executionID` int unsigned NOT NULL COMMENT '关联的流水线执行记录ID',
	`parent` int unsigned NOT NULL DEFAULT 0 COMMENT '父执行记录ID（用于阶段重试场景，无则为0）',
	`name` varchar(255) NOT NULL DEFAULT '' COMMENT '显示名称',
	`code` varchar(255) NOT NULL COMMENT '阶段内部标识名称',
	`sn` int unsigned NOT NULL COMMENT 'serial number阶段序号',
	`type` varchar(50) NOT NULL DEFAULT '' COMMENT '类型:agent',
	`host` varchar(255) NOT NULL DEFAULT '' COMMENT 'machine运行机器标识',
	`os` varchar(255) NOT NULL DEFAULT '' COMMENT '阶段运行的操作系统（如：linux、windows、macos）',
	`arch` varchar(255) NOT NULL DEFAULT '' COMMENT '阶段运行的CPU架构（如：amd64、arm64、x86）',
	`labels` varchar(255) NOT NULL DEFAULT '' COMMENT '阶段标签，runtime.selector',
	`depends` varchar(255) NOT NULL DEFAULT '' COMMENT '依赖的前置阶段code',
	`runPolicy` varchar(50) NOT NULL DEFAULT 'onSuccess' COMMENT '执行策略：onSuccess,onFailure,always,custom',
	`customPolicy` varchar(255) NOT NULL DEFAULT '' COMMENT '自定义执行策略',
	`failPolicy` varchar(50) NOT NULL DEFAULT 'fail' COMMENT '失败策略：fail,ignore,warning',
	`started` datetime NULL COMMENT '阶段开始执行时间',
	`finished` datetime NULL COMMENT '阶段结束执行时间',
	`duration` int NULL COMMENT '执行时长(s)',
	`status` varchar(50) NOT NULL COMMENT '阶段执行状态',
	`errorType` varchar(50) NOT NULL COMMENT '失败类型',
	`errorCode` int NOT NULL COMMENT '阶段执行退出码（0=成功，非0=失败）',
	`errorLog` text NOT NULL COMMENT '阶段执行错误信息（无错误则为空字符串）',
	PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='流水线阶段执行表';


CREATE TABLE IF NOT EXISTS `ops_step_executions` (
	`id` int unsigned  NOT NULL AUTO_INCREMENT COMMENT '步骤唯一主键',
	`stageID` int unsigned NOT NULL COMMENT '关联阶段stages主键ID',
	`name` varchar(255) NOT NULL DEFAULT '' COMMENT '显示名称',
	`code` varchar(100) NOT NULL COMMENT '步骤业务唯一标识',
	`sn` tinyint(1) unsigned NOT NULL COMMENT '步骤在所属阶段内的序号',
	`image` varchar(255) NOT NULL DEFAULT '' COMMENT '步骤执行所需的容器镜像',
	`depends` varchar(255) NOT NULL DEFAULT '' COMMENT '步骤依赖的其他步骤code',
	`groupID` int unsigned NOT NULL DEFAULT 0 COMMENT '步骤所属分组ID',
	`failPolicy` varchar(50) NOT NULL DEFAULT 'fail' COMMENT '失败策略：fail,ignore,warning',
	`started` datetime NULL COMMENT '步骤执行开始时间',
	`finished` datetime NULL COMMENT '步骤执行完成时间',
	`duration` int NULL COMMENT '执行时长(s)',
	`status` varchar(50) NOT NULL COMMENT '步骤执行状态（枚举）',
	`errorCode` int NOT NULL COMMENT '步骤执行退出码',
	`errorLog` varchar(500) NOT NULL COMMENT '步骤执行失败时的错误信息，成功时为空字符串',
	PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT '流水线步骤执行表';