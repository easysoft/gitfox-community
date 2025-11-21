CREATE TABLE IF NOT EXISTS `ops_repository` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '代码库ID',
	`spaceID` int NOT NULL COMMENT '空间ID',
	`identifier` varchar(255) NOT NULL COMMENT '代码库唯一标识',
	`desc` varchar(500) NOT NULL DEFAULT '' COMMENT '代码库描述',
	`gitUID` char(42) NOT NULL COMMENT 'Git仓库唯一标识',
	`defaultBranch` varchar(255) NOT NULL COMMENT '默认分支名',
	`forkID` int unsigned NULL DEFAULT NULL COMMENT '派生来源代码库ID',
	`empty` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否为空库',
	`acl` varchar(30) NOT NULL DEFAULT 'open' COMMENT '权限:private,open',
	`mirror` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否为镜像库（0=普通库，1=镜像库）',
	`status` tinyint(1) NOT NULL DEFAULT 0 COMMENT '代码库状态',
	`mutex` int NOT NULL DEFAULT 0 COMMENT '乐观锁版本号',    
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB COMMENT='代码库表';
CREATE UNIQUE INDEX uk_repo_space_id ON ops_repository (identifier, spaceID);
CREATE UNIQUE INDEX uk_gituid ON ops_repository (gitUID);
CREATE INDEX idx_deleted ON ops_repository (deleted);