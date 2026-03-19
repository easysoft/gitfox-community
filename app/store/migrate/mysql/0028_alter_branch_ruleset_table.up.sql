ALTER TABLE `ops_branch_ruleset`
	ADD COLUMN `ppmCreateUser` varchar(500) NOT NULL COMMENT '创建推送合并请求人员' AFTER `forcePushUser`,
	ADD COLUMN `ppmHandleUser` varchar(500) NOT NULL COMMENT '处理推送合并请求人员' AFTER `ppmCreateUser`,
	ADD COLUMN `commitLine` smallint unsigned NOT NULL DEFAULT 0 COMMENT 'commit限制行数' AFTER `targetBranch`,
	ADD COLUMN `pushLine` smallint unsigned NOT NULL DEFAULT 0 COMMENT 'push行数限制行数' AFTER `commitLine`,
	ADD COLUMN `forceReview` tinyint(1) unsigned NOT NULL DEFAULT 0 COMMENT '是否开启强制评审' AFTER `pushLine`,
	ADD COLUMN `reviewFlowID` int unsigned NOT NULL DEFAULT 0 COMMENT '评审流程ID' AFTER `forceReview`;


ALTER TABLE `ops_repo`
	ADD COLUMN `branchArchivable` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否开启分支归档' AFTER `synced`;