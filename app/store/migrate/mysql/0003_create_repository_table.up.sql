CREATE TABLE IF NOT EXISTS `ops_repo` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '代码库ID',
	`spaceID` int NOT NULL COMMENT '空间ID',
	`name` varchar(255) NOT NULL COMMENT '代码库唯一标识',
	`desc` varchar(500) NOT NULL DEFAULT '' COMMENT '代码库描述',
	`product` varchar(255) NOT NULL DEFAULT '' COMMENT '关联产品',
	`gitUID` char(42) NOT NULL COMMENT 'Git仓库唯一标识',
	`defaultBranch` varchar(255) NOT NULL COMMENT '默认分支名',
	`forkID` int unsigned NULL DEFAULT NULL COMMENT '派生来源代码库ID',
	`empty` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否为空库',
	`acl` varchar(30) NOT NULL DEFAULT 'open' COMMENT '权限:private,open',
	`mirror` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否为镜像库（0=普通库，1=镜像库）',
	`status` tinyint(1) NOT NULL DEFAULT 0 COMMENT '代码库状态',
	`mutex` int NOT NULL DEFAULT 0 COMMENT '乐观锁版本号',
	`synced` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '同步标识',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB COMMENT='代码库表';
CREATE UNIQUE INDEX `uk_repo_space_id` ON `ops_repo` (`name`, `spaceID`);
CREATE UNIQUE INDEX `uk_gituid` ON `ops_repo` (`gitUID`);
CREATE INDEX `idx_deleted` ON `ops_repo` (`deleted`);

/* 以下表为devops使用 */

CREATE TABLE IF NOT EXISTS `ops_repouser` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `repo` int unsigned NOT NULL DEFAULT 0 COMMENT '所属代码库',
  `account` varchar(30) NOT NULL DEFAULT '' COMMENT '用户帐号',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='代码库用户关联表';
CREATE UNIQUE INDEX `uk_repouser` ON `ops_repouser` (`repo`,`account`);

CREATE TABLE IF NOT EXISTS `ops_repohistory` (
	`id` int(10) unsigned NOT NULL AUTO_INCREMENT,
	`repo` int(10) unsigned NOT NULL DEFAULT '0',
	`revision` varchar(40) NOT NULL DEFAULT '',
	`commit` int(10) unsigned NOT NULL DEFAULT '0',
	`comment` text NULL DEFAULT NULL,
	`committer` varchar(100) NOT NULL DEFAULT '',
	`time` DATETIME NULL DEFAULT NULL,
	PRIMARY KEY (`id`)
)ENGINE=InnoDB;
CREATE INDEX `idx_repo` ON `ops_repohistory` (`repo`);
CREATE INDEX `idx_revision` ON `ops_repohistory` (`revision`);


CREATE TABLE IF NOT EXISTS `ops_repofiles` (
	`id` int(11) unsigned NOT NULL AUTO_INCREMENT,
	`repo` int(10) unsigned NOT NULL DEFAULT '0',
	`revision` int(10) unsigned NOT NULL DEFAULT '0',
	`path` varchar(255) NOT NULL DEFAULT '',
	`oldPath` varchar(255) NOT NULL DEFAULT '',
	`parent` varchar(255) NOT NULL DEFAULT '',
	`type` varchar(20) NOT NULL DEFAULT '',
	`action` char(1) NOT NULL DEFAULT '',
	PRIMARY KEY (`id`)
)ENGINE=InnoDB;
CREATE INDEX `idx_path` ON `ops_repofiles` (`path`);
CREATE INDEX `idx_parent` ON `ops_repofiles` (`parent`);
CREATE INDEX `idx_repo` ON `ops_repofiles` (`repo`);
CREATE INDEX `idx_revision` ON `ops_repofiles` (`revision`);


CREATE TABLE IF NOT EXISTS `ops_repobranch` (
	`id` int(10) unsigned NOT NULL AUTO_INCREMENT,
	`repo` int(10) unsigned NOT NULL DEFAULT '0',
	`revision` int(10) unsigned NOT NULL DEFAULT '0',
	`branch` varchar(100) NOT NULL DEFAULT '',
	PRIMARY KEY (`id`)
)ENGINE=InnoDB;
CREATE UNIQUE INDEX `uk_repo_revision_branch` ON `ops_repouser` (`repo`,`revision`,`branch`);
CREATE INDEX `idx_branch` ON `ops_repobranch` (`branch`);
CREATE INDEX `idx_revision` ON `ops_repobranch` (`revision`);