CREATE TABLE IF NOT EXISTS `ops_artifact` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`spaceID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属空间ID，0表示全局级',
	`repoID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属代码库ID，0表示非代码库级',
	`name` varchar(100) NOT NULL COMMENT '视图名称',
	`desc` text NULL DEFAULT NULL COMMENT '视图描述',
	`format` varchar(50) NOT NULL COMMENT '制品格式',
	`type` varchar(30) NOT NULL DEFAULT '' COMMENT '视图类型:global/space/repo',
	`isDefault` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否默认视图，0-否，1-是',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='制品视图表（支持全局/空间/代码库三级作用域）';

-- 注意：artifact 表不使用严格唯一索引，因为回收站机制需要在保留期内允许同名软删记录存在。
-- 名称唯一性由应用层在 CreateView 中保证（详见 service/artifact/artifact.go）。
CREATE INDEX `idx_artifact_spaceID_repoID_name` ON `ops_artifact` (`spaceID`, `repoID`, `name`);
CREATE INDEX `idx_artifact_spaceID_deleted` ON `ops_artifact` (`spaceID`, `deleted`);
CREATE INDEX `idx_artifact_repoID_deleted` ON `ops_artifact` (`repoID`, `deleted`);
CREATE INDEX `idx_artifact_spaceID_repoID_deleted` ON `ops_artifact` (`spaceID`, `repoID`, `deleted`);
CREATE INDEX `idx_artifact_spaceID_repoID_isDefault_deleted` ON `ops_artifact` (`spaceID`, `repoID`, `isDefault`, `deleted`);

CREATE TABLE IF NOT EXISTS `ops_artifact_groups` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`artifactID` int unsigned NOT NULL COMMENT '所属视图ID',
	`parentID` int unsigned NOT NULL DEFAULT 0 COMMENT '父级分组ID，0表示根分组',
	`name` varchar(255) NOT NULL COMMENT '分组名称',
	`format` varchar(50) NOT NULL COMMENT '制品格式',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='制品分组表';

CREATE UNIQUE INDEX `uk_groups_artifactID_parentID_name` ON `ops_artifact_groups` (`artifactID`, `parentID`, `name`);
CREATE INDEX `idx_groups_artifactID_deleted` ON `ops_artifact_groups` (`artifactID`, `deleted`);
CREATE INDEX `idx_groups_artifactID_format_parentID_deleted` ON `ops_artifact_groups` (`artifactID`, `format`, `parentID`, `deleted`);
CREATE INDEX `idx_groups_parentID_deleted` ON `ops_artifact_groups` (`parentID`, `deleted`);

CREATE TABLE IF NOT EXISTS `ops_artifact_packages` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`artifactID` int unsigned NOT NULL COMMENT '所属视图ID',
	`name` varchar(100) NOT NULL COMMENT '制品包名称',
	`groupID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属分组ID，0-无分组',
	`namespace` varchar(255) NOT NULL DEFAULT '' COMMENT '命名空间',
	`format` varchar(50) NOT NULL COMMENT '制品格式',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='制品包表';

CREATE UNIQUE INDEX `uk_pkgs_artifactID_name_namespace_format` ON `ops_artifact_packages` (`artifactID`, `name`, `namespace`, `format`);
CREATE INDEX `idx_pkgs_artifactID_format_deleted` ON `ops_artifact_packages` (`artifactID`, `format`, `deleted`);
CREATE INDEX `idx_pkgs_artifactID_namespace_deleted` ON `ops_artifact_packages` (`artifactID`, `namespace`, `deleted`);
CREATE INDEX `idx_pkgs_groupID_deleted` ON `ops_artifact_packages` (`groupID`, `deleted`);
CREATE INDEX `idx_pkgs_deleted` ON `ops_artifact_packages` (`deleted`);

CREATE TABLE IF NOT EXISTS `ops_artifact_versions` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`packageID` int unsigned NOT NULL COMMENT '所属制品包ID',
	`version` varchar(100) NOT NULL COMMENT '版本号',
	`metadata` text NULL DEFAULT NULL COMMENT '版本元数据',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='制品版本表';

CREATE UNIQUE INDEX `uk_vers_packageID_version` ON `ops_artifact_versions` (`packageID`, `version`);
CREATE INDEX `idx_vers_packageID_deleted` ON `ops_artifact_versions` (`packageID`, `deleted`);
CREATE INDEX `idx_vers_createdDate` ON `ops_artifact_versions` (`createdDate`);
CREATE INDEX `idx_vers_editedDate` ON `ops_artifact_versions` (`editedDate`);
CREATE INDEX `idx_vers_deleted` ON `ops_artifact_versions` (`deleted`);

CREATE TABLE IF NOT EXISTS `ops_artifact_assets` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`versionID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属版本ID，0表示无版本关联',
	`artifactID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属视图ID',
	`groupID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属分组ID，0表示无分组',
	`format` varchar(50) NOT NULL DEFAULT '' COMMENT '制品格式',
	`path` varchar(255) NOT NULL COMMENT '制品路径',
	`contentType` varchar(100) NOT NULL DEFAULT '' COMMENT '内容类型',
	`kind` varchar(50) NOT NULL DEFAULT '' COMMENT '制品类型',
	`metadata` text NULL DEFAULT NULL COMMENT '制品元数据',
	`blobID` int unsigned NOT NULL DEFAULT 0 COMMENT '关联文件对象ID',
	`checkSum` text NULL COMMENT '校验和',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='制品资源表';

CREATE UNIQUE INDEX `uk_assets_versionID_path_deleted` ON `ops_artifact_assets` (`versionID`, `path`, `deleted`);
CREATE INDEX `idx_assets_artifactID_format_path_deleted` ON `ops_artifact_assets` (`artifactID`, `format`, `path`, `deleted`);
CREATE INDEX `idx_assets_groupID_path_deleted` ON `ops_artifact_assets` (`groupID`, `path`, `deleted`);
CREATE INDEX `idx_assets_blobID` ON `ops_artifact_assets` (`blobID`);
CREATE INDEX `idx_assets_deleted` ON `ops_artifact_assets` (`deleted`);

CREATE TABLE IF NOT EXISTS `ops_artifact_meta_assets` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`artifactID` int unsigned NOT NULL COMMENT '所属视图ID',
	`format` varchar(50) NOT NULL COMMENT '制品格式',
	`path` varchar(255) NOT NULL COMMENT '元数据路径',
	`contentType` varchar(100) NOT NULL DEFAULT '' COMMENT '内容类型',
	`kind` varchar(50) NOT NULL DEFAULT '' COMMENT '元数据类型',
	`blobID` int unsigned NOT NULL DEFAULT 0 COMMENT '关联文件对象ID',
	`checkSum` text NULL COMMENT '校验和',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='制品元数据资源表';

CREATE UNIQUE INDEX `uk_meta_artifactID_format_path` ON `ops_artifact_meta_assets` (`artifactID`, `path`, `format`);
CREATE INDEX `idx_meta_artifactID_deleted` ON `ops_artifact_meta_assets` (`artifactID`, `deleted`);
CREATE INDEX `idx_meta_blobID` ON `ops_artifact_meta_assets` (`blobID`);
CREATE INDEX `idx_meta_deleted` ON `ops_artifact_meta_assets` (`deleted`);

CREATE TABLE IF NOT EXISTS `ops_artifact_blobs` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`storageID` int unsigned NOT NULL DEFAULT 0 COMMENT '存储后端ID',
	`ref` varchar(255) NOT NULL DEFAULT '' COMMENT '存储引用',
	`size` bigint unsigned NOT NULL DEFAULT 0 COMMENT '文件大小',
	`downloads` bigint unsigned NOT NULL DEFAULT 0 COMMENT '下载次数',
	`metadata` text NULL DEFAULT NULL COMMENT '文件元数据',
	`assetID` int unsigned NOT NULL DEFAULT 0 COMMENT '关联制品资源ID',
	`checkSum` text NULL COMMENT '校验和',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='制品文件对象表';

CREATE UNIQUE INDEX `uk_blobs_ref_storageID` ON `ops_artifact_blobs` (`ref`, `storageID`);
CREATE INDEX `idx_blobs_assetID_deleted` ON `ops_artifact_blobs` (`assetID`, `deleted`);
CREATE INDEX `idx_blobs_deleted` ON `ops_artifact_blobs` (`deleted`);

CREATE TABLE IF NOT EXISTS `ops_artifact_tree_nodes` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`artifactID` int unsigned NOT NULL COMMENT '所属视图ID',
	`parentID` int unsigned NOT NULL DEFAULT 0 COMMENT '父级节点ID，0表示根节点',
	`name` varchar(255) NOT NULL COMMENT '节点名称',
	`path` varchar(255) NOT NULL COMMENT '节点路径',
	`type` varchar(20) NOT NULL COMMENT '节点类型',
	`format` varchar(50) NOT NULL COMMENT '制品格式',
	`targetID` int unsigned NULL DEFAULT NULL COMMENT '关联目标ID',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='制品树节点表';

CREATE UNIQUE INDEX `uk_nodes_artifactID_path_format` ON `ops_artifact_tree_nodes` (`artifactID`, `path`, `format`);
CREATE INDEX `idx_nodes_parentID_deleted` ON `ops_artifact_tree_nodes` (`parentID`, `deleted`);
CREATE INDEX `idx_nodes_artifactID_type_targetID_deleted` ON `ops_artifact_tree_nodes` (`artifactID`, `type`, `targetID`, `deleted`);
CREATE INDEX `idx_nodes_artifactID_format_deleted` ON `ops_artifact_tree_nodes` (`artifactID`, `format`, `deleted`);
CREATE INDEX `idx_nodes_deleted` ON `ops_artifact_tree_nodes` (`deleted`);
