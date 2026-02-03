CREATE TABLE IF NOT EXISTS `ops_ppm` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '自增主键',
	`title` varchar(255) NOT NULL COMMENT 'PR 标题',
	`desc` TEXT NOT NULL COMMENT '拉取请求详细描述',
	`repoID` int unsigned NOT NULL COMMENT '关联zt_repo表的id',
	`sourceRepoID` int unsigned NOT NULL COMMENT '源仓库ID，关联repository表的id',
	`sourceBranch` varchar(255) NOT NULL COMMENT '源分支名称（待合并的分支）',
	`sourceSHA` char(40) NOT NULL COMMENT '源分支最新提交的SHA-1值',
	`targetRepoID` int unsigned NOT NULL COMMENT '目标仓库ID，关联repository表的id',
	`targetBranch` varchar(255) NOT NULL COMMENT '目标分支名称（合并目标分支）',
	`mergedBy` varchar(30) NULL DEFAULT NULL COMMENT '执行合并操作的用户（未合并则为NULL）',
	`mergeMethod` varchar(255) NULL DEFAULT NULL COMMENT '合并方法:merge/squash/rebase/fast-forward',
	`mergeTargetSHA` char(40) NULL DEFAULT NULL COMMENT '合并时目标分支的 SHA,用于验证目标分支未变更',
	`mergeBaseSHA` char(40) NOT NULL DEFAULT '' COMMENT '合并基准 SHA(源分支和目标分支的共同祖先)',
	`mergeSHA` char(40) NULL DEFAULT NULL COMMENT '合并后生成的 commit SHA',
	`mergeCheckStatus` varchar(255) NOT NULL COMMENT '合并前检查状态unchecked=未检查，conflict=有冲突，mergeable=无冲突）',
	`mergeConflicts` text NULL DEFAULT NULL COMMENT 'merge/squash 冲突文件列表,换行符分隔',
	`rebaseCheckStatus` varchar(255) NOT NULL DEFAULT 'unchecked' COMMENT '变基检查状态（unchecked=未检查，conflict=有冲突，mergeable=无冲突）',
	`rebaseConflicts` text NULL DEFAULT NULL COMMENT '变基时产生的冲突详情（文件路径、冲突内容，未变基/无冲突则为NULL）',
	`commitCount` int unsigned NULL DEFAULT NULL COMMENT 'PR 包含的 commit 数量',
	`fileCount` int unsigned NULL DEFAULT NULL COMMENT 'PR 修改的文件数量',
	`additions` int unsigned NULL DEFAULT NULL COMMENT '新增的代码行数',
	`deletions` int unsigned NULL DEFAULT NULL COMMENT '删除的代码行数',
	`flow` tinyint(1) NOT NULL DEFAULT '0' COMMENT '拉取请求对应的流程标识（0=默认流程，其他值对应自定义流程）',
	`reviewFlowID` int unsigned NOT NULL DEFAULT '0' COMMENT '关联的评审流程ID',
	`status` varchar(255) NOT NULL COMMENT '拉取请求状态（如open/closed/merged/draft等）',
	`mergedDate` datetime NULL COMMENT '合并时间',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '',
	`editedDate` datetime NULL COMMENT '更新时间',
	`approval` int unsigned NOT NULL DEFAULT '0' COMMENT '审批实例',
	`reviewers` varchar(255) NOT NULL DEFAULT '' COMMENT '最新节点评审人',
	`approvalflow` int unsigned NOT NULL DEFAULT 0 COMMENT '审批流模板' ,
	`reviewStatus` varchar(20) NOT NULL DEFAULT '' COMMENT '审批状态',
	PRIMARY KEY (`id`)
	) ENGINE=InnoDB COMMENT='合并请求主表';

CREATE INDEX `idx_createdBy` ON ops_ppm (`createdBy`);
CREATE INDEX `idx_sourceRepoID_sourceBranch_targetRepoID_targetBranch` ON ops_ppm(`sourceRepoID`, `sourceBranch`, `targetRepoID`, `targetBranch`);

CREATE TABLE IF NOT EXISTS `ops_branch_type` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `repo` int unsigned NOT NULL COMMENT '仓库ID，0为系统级别分支类型',
  `name` varchar(255) NOT NULL COMMENT '分支类型名称',
  `key` varchar(255) NOT NULL DEFAULT '' COMMENT '分支类型键值，仓库下唯一',
  `prefix` varchar(255) NOT NULL DEFAULT '' COMMENT '分支前缀，仓库下唯一',
  `desc` varchar(500) NOT NULL DEFAULT '' COMMENT '描述',
  `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
  `createdDate` datetime DEFAULT NULL COMMENT '创建时间',
  `editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '编辑人',
  `editedDate` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted` tinyint NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='分支类型表';
CREATE UNIQUE INDEX `uk_repo_branch_key` ON `ops_branch_type` (`repo`, `key`);
CREATE UNIQUE INDEX `uk_repo_branch_prefix` ON `ops_branch_type` (`repo`, `prefix`);

INSERT INTO `ops_branch_type` (`id`, `repo`, `name`, `key`, `prefix`, `desc`, `createdBy`, `createdDate`, `editedBy`, `editedDate`, `deleted`) VALUES
(1, 0, '主干分支', 'main', 'main/,master/', '存放随时可以发布/部署的稳定代码。', 'system', NOW(), '', NULL, 0),
(2, 0, '开发分支', 'develop', 'develop/,dev/,story/', '日常开发的集成分支，完成的功能会合并到这里做集成测试。', 'system', NOW(), '', NULL, 0),
(3, 0, '特性分支', 'feature', 'feature/', '为实现某个新功能或任务而创建的短期分支。', 'system', NOW(), '', NULL, 0),
(4, 0, '预发布分支', 'release', 'release/', '用于准备一次正式发布的分支。', 'system', NOW(), '', NULL, 0),
(5, 0, '缺陷修复分支', 'bugfix', 'bugfix/,bug/', '用于修复非紧急的缺陷。', 'system', NOW(), '', NULL, 0),
(6, 0, '热修复分支', 'hotfix', 'hotfix/', '用于线上紧急修复。', 'system', NOW(), '', NULL, 0);


CREATE TABLE IF NOT EXISTS `ops_branch_ruleset` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `repo` int unsigned NOT NULL COMMENT '仓库ID',
  `branchType` int unsigned NOT NULL COMMENT '分支类型ID',
  `branchName` varchar(255) NOT NULL DEFAULT '' COMMENT '分支名',
  `createUser` varchar(500) NOT NULL COMMENT '创建分支权限人员',
  `deleteUser` varchar(500) NOT NULL COMMENT '删除分支权限人员',
  `updateUser` varchar(500) NOT NULL COMMENT '更新分支权限人员',
  `forcePushUser` varchar(500) NOT NULL COMMENT '强制推送权限人员',
  `sourceBranch` varchar(500) NOT NULL DEFAULT '' COMMENT '允许合并来源类型，为空为全部分支',
  `targetBranch` varchar(500) NOT NULL DEFAULT '' COMMENT '允许合并的目标分支类型，为空为全部分支',
  `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
  `createdDate` datetime DEFAULT NULL COMMENT '创建时间',
  `editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '编辑人',
  `editedDate` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted` tinyint NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='分支规则表';
CREATE INDEX `idx_repo` ON `ops_branch_ruleset` (`repo`);


CREATE TABLE IF NOT EXISTS `ops_review_flow` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `repo` int unsigned NOT NULL DEFAULT 0 COMMENT '仓库ID',
  `branchType` varchar(255) NOT NULL DEFAULT '' COMMENT '分支类型ID，0-全部分支类型，具体用逗号分隔',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '评审流程名称',
  `desc` varchar(500) NOT NULL DEFAULT '' COMMENT '评审流程描述',
  `definition` TEXT NOT NULL DEFAULT '' COMMENT '评审规则，json定义',
  `status` varchar(20) NOT NULL DEFAULT '' COMMENT 'enable-启用，disable-停用',
  `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
  `createdDate` datetime DEFAULT NULL COMMENT '创建时间',
  `editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '编辑人',
  `editedDate` datetime DEFAULT NULL COMMENT '更新时间',
  `deleted` tinyint NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='评审流程表';
CREATE INDEX `idx_repo` ON `ops_review_flow` (`repo`);


CREATE TABLE IF NOT EXISTS `ops_request_reviewers` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `requestID` int unsigned  NOT NULL DEFAULT 0 COMMENT '关联合并请求ID',
  `repoID` int unsigned NOT NULL DEFAULT 0 COMMENT '关联仓库ID',
  `decision` varchar(255) NOT NULL DEFAULT '' COMMENT '最新审核决策（如：approve-批准、reject-拒绝、pending-待审核等）',
  `sha` varchar(40) NOT NULL DEFAULT '' COMMENT '审核对应的代码提交SHA校验值',
  `account` varchar(30) NOT NULL DEFAULT '' COMMENT '评审人',
  `opinion` mediumtext DEFAULT NULL COMMENT '评审意见',
  `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
  `createdDate` datetime NULL COMMENT '创建时间',
  `editedDate` datetime NULL COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='评审人员表';
CREATE INDEX `idx_requestID` ON `ops_request_reviewers` (`requestID`);
CREATE INDEX `idx_account` ON `ops_request_reviewers` (`account`);
