CREATE TABLE IF NOT EXISTS `ops_branch_type` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
	`repoID` int unsigned NOT NULL COMMENT '仓库ID，0为系统级别分支类型',
	`name` varchar(255) NOT NULL COMMENT '分支类型名称',
	`key` varchar(255) NOT NULL DEFAULT '' COMMENT '分支类型键值，仓库下唯一',
	`prefix` varchar(255) NOT NULL DEFAULT '' COMMENT '分支前缀，仓库下唯一',
	`desc` varchar(500) NOT NULL DEFAULT '' COMMENT '描述', 
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB COMMENT='分支类型表';
CREATE UNIQUE INDEX uk_repo_branch_key ON ops_branch_type (repoID, `key`);
CREATE UNIQUE INDEX uk_repo_branch_prefix ON ops_branch_type (repoID, prefix);
CREATE INDEX idx_deleted ON ops_branch_type (deleted);
