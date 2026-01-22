CREATE TABLE IF NOT EXISTS `ops_pipeline` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '流水线ID，自增主键',
	`name` varchar(255) NOT NULL COMMENT '流水线名称',
	`type` varchar(255) NOT NULL COMMENT '流水线类型：系统/空间/仓库',
	`spaceID` int unsigned NOT NULL COMMENT '关联的空间ID',
	`repoID` int unsigned NOT NULL COMMENT '关联的代码仓库ID',
	`desc` varchar(255) NOT NULL DEFAULT '' COMMENT '流水线描述信息',
	`status` varchar(255) NOT NULL COMMENT '流水线状态：草稿/激活',
	`latestVersion` int NOT NULL DEFAULT 0 COMMENT '流水线最新版本',
	`defaultBranch` varchar(255) NOT NULL COMMENT '流水线默认分支',
	`yamlPath` varchar(255) NOT NULL DEFAULT '' COMMENT '流水线配置文件路径',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='流水线基本信息表';

CREATE UNIQUE INDEX `uk_space_repo_name` ON `ops_pipelines` (`spaceID`,`repoID`,`name`);


CREATE TABLE IF NOT EXISTS `ops_pipeline_content` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
	`pipelineID` int unsigned NOT NULL COMMENT '关联的流水线ID',
	`version` int unsigned NOT NULL DEFAULT 0 COMMENT '流水线版本',
	`data` longtext NOT NULL DEFAULT '' COMMENT '流水线内容',
	`variables` text NOT NULL DEFAULT '' COMMENT '流水线参数',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '',
	`editedDate` datetime NULL COMMENT '更新时间',
	PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='流水线配置表';

CREATE UNIQUE INDEX `uk_pipelineID_version` ON `ops_pipeline_contents` (`pipelineID`, `version`);