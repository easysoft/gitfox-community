-- 1. 制品库主表
CREATE TABLE IF NOT EXISTS `ops_artifact_libs` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `spaceID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属空间ID，0表示全局级',
    `repoID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属代码库ID，0表示非代码库级',
    `name` varchar(100) NOT NULL COMMENT '制品库名称',
    `desc` varchar(500) NOT NULL DEFAULT '' COMMENT '制品库描述',
    `type` varchar(50) NOT NULL COMMENT '制品类型:file/container/raw/helm/pypi/npm/maven/composer',
    `scope` varchar(30) NOT NULL DEFAULT '' COMMENT '制品库类型:global/space/repo',
    `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
    `createdDate` datetime DEFAULT NULL COMMENT '创建时间',
    `editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
    `editedDate` datetime DEFAULT NULL COMMENT '更新时间',
    `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB COMMENT= '制品库主表（支持全局/空间/代码库三级作用域）';

CREATE INDEX `idx_artifact_spaceID_repoID_name` ON `ops_artifact_libs` (`spaceID`, `repoID`, `name`);
CREATE INDEX `idx_artifact_spaceID_deleted` ON `ops_artifact_libs` (`spaceID`, `deleted`);
CREATE INDEX `idx_artifact_repoID_deleted` ON `ops_artifact_libs` (`repoID`, `deleted`);
CREATE INDEX `idx_artifact_spaceID_repoID_deleted` ON `ops_artifact_libs` (`spaceID`, `repoID`, `deleted`);


-- 2. 制品树节点表
CREATE TABLE IF NOT EXISTS `ops_artifact_tree_nodes` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `artifactLibID` int unsigned NOT NULL COMMENT '所属制品库ID',
    `parentID` int unsigned NOT NULL DEFAULT 0 COMMENT '父级节点ID，0表示根节点',
    `nodeName` varchar(255) NOT NULL DEFAULT '' COMMENT '节点名称',
    `nodePath` varchar(255) NOT NULL DEFAULT '' COMMENT '节点路径',
    `linkTable` varchar(20) NOT NULL DEFAULT '' COMMENT '关联表',
    `linkRecord` int unsigned DEFAULT NULL COMMENT '关联表ID',
    `type` varchar(50) NOT NULL DEFAULT '' COMMENT '制品类型:file/container/raw/helm/pypi/npm/maven/composer',
    `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
    `createdDate` datetime DEFAULT NULL COMMENT '创建时间',
    `editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
    `editedDate` datetime DEFAULT NULL COMMENT '更新时间',
    `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB  COMMENT= '制品树节点表';

CREATE UNIQUE INDEX `uk_nodes_artifactLibID_nodePath_type` ON `ops_artifact_tree_nodes` (`artifactLibID`, `nodePath`, `type`);
CREATE INDEX `idx_nodes_parentID_deleted` ON `ops_artifact_tree_nodes` (`parentID`, `deleted`);
CREATE INDEX `idx_nodes_artifactLibID_linkTable_linkRecord_deleted` ON `ops_artifact_tree_nodes` (`artifactLibID`, `linkTable`, `linkRecord`, `deleted`);
CREATE INDEX `idx_nodes_artifactLibID_type_deleted` ON `ops_artifact_tree_nodes` (`artifactLibID`, `type`, `deleted`);
CREATE INDEX `idx_nodes_deleted` ON `ops_artifact_tree_nodes` (`deleted`);


-- 3. 制品分组表
CREATE TABLE IF NOT EXISTS `ops_artifact_groups` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `artifactLibID` int unsigned NOT NULL COMMENT '所属制品库ID',
    `parentID` int unsigned NOT NULL DEFAULT 0 COMMENT '父级分组ID，0表示根分组',
    `name` varchar(255) NOT NULL DEFAULT '' COMMENT '分组名称',
    `type` varchar(50) NOT NULL DEFAULT '' COMMENT '制品类型:file/container/raw/helm/pypi/npm/maven/composer',
    `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
    `createdDate` datetime DEFAULT NULL COMMENT '创建时间',
    `editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
    `editedDate` datetime DEFAULT NULL COMMENT '更新时间',
    `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB  COMMENT= '制品分组表';

CREATE INDEX `idx_groups_artifactLibID_deleted` ON `ops_artifact_groups` (`artifactLibID`, `deleted`);
CREATE INDEX `idx_groups_artifactLibID_parentID_deleted` ON `ops_artifact_groups` (`artifactLibID`, `parentID`, `deleted`);
CREATE INDEX `idx_groups_parentID_deleted` ON `ops_artifact_groups` (`parentID`, `deleted`);
CREATE INDEX `idx_groups_artifactLibID_parentID_name` ON `ops_artifact_groups` (`artifactLibID`, `parentID`, `name`);


-- 4. 制品包表
CREATE TABLE IF NOT EXISTS `ops_artifact_packages` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `artifactLibID` int unsigned NOT NULL COMMENT '所属制品库ID',
    `groupID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属分组ID',
    `namespace` varchar(255) NOT NULL DEFAULT '' COMMENT '命名空间',
    `name` varchar(100) NOT NULL DEFAULT '' COMMENT '制品包名称',
    `type` varchar(50) NOT NULL DEFAULT '' COMMENT '制品类型:file/container/raw/helm/pypi/npm/maven/composer',
    `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
    `createdDate` datetime DEFAULT NULL COMMENT '创建时间',
    `editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
    `editedDate` datetime DEFAULT NULL COMMENT '更新时间',
    `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB  COMMENT= '制品包表';

CREATE UNIQUE INDEX `uk_pkgs_artifactLibID_name_namespace_type` ON `ops_artifact_packages` (`artifactLibID`, `name`, `namespace`, `type`);
CREATE INDEX `idx_pkgs_artifactLibID_type_deleted` ON `ops_artifact_packages` (`artifactLibID`, `type`, `deleted`);
CREATE INDEX `idx_pkgs_artifactLibID_namespace_deleted` ON `ops_artifact_packages` (`artifactLibID`, `namespace`, `deleted`);
CREATE INDEX `idx_pkgs_groupID_deleted` ON `ops_artifact_packages` (`groupID`, `deleted`);
CREATE INDEX `idx_pkgs_deleted` ON `ops_artifact_packages` (`deleted`);


-- 5. 制品版本表
CREATE TABLE IF NOT EXISTS `ops_artifact_versions` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `artifactLibID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属制品库ID',
    `packageID` int unsigned NOT NULL COMMENT '所属制品包ID',
    `version` varchar(100) NOT NULL DEFAULT '' COMMENT '版本号',
    `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
    `createdDate` datetime DEFAULT NULL COMMENT '创建时间',
    `editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
    `editedDate` datetime DEFAULT NULL COMMENT '更新时间',
    `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB  COMMENT= '制品版本表';

CREATE UNIQUE INDEX `uk_vers_artifactLibID_packageID_version` ON `ops_artifact_versions` (`artifactLibID`, `packageID`, `version`);
CREATE INDEX `idx_vers_artifactLibID_deleted` ON `ops_artifact_versions` (`artifactLibID`, `deleted`);
CREATE INDEX `idx_vers_packageID_deleted` ON `ops_artifact_versions` (`packageID`, `deleted`);
CREATE INDEX `idx_vers_createdDate` ON `ops_artifact_versions` (`createdDate`);
CREATE INDEX `idx_vers_editedDate` ON `ops_artifact_versions` (`editedDate`);
CREATE INDEX `idx_vers_deleted` ON `ops_artifact_versions` (`deleted`);


-- 6. 制品资源表
CREATE TABLE IF NOT EXISTS `ops_artifact_assets` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `artifactLibID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属制品库ID',
    `groupID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属分组ID，0表示无分组',
    `packageID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属制品包ID',
    `versionID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属版本ID，0表示无版本关联',
    `blobID` int unsigned NOT NULL DEFAULT 0 COMMENT '关联文件对象ID',
    `type` varchar(50) NOT NULL DEFAULT '' COMMENT '制品类型:file/container/raw/helm/pypi/npm/maven/composer',
    `basename` varchar(255) NOT NULL DEFAULT '' COMMENT '制品名称',
    `contentType` varchar(100) NOT NULL DEFAULT '' COMMENT '内容类型',
    `rank` varchar(50) NOT NULL DEFAULT '' COMMENT '制品等级 main/sub',
    `checksum` varchar(500) NOT NULL DEFAULT '' COMMENT '校验摘要',
    `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
    `createdDate` datetime DEFAULT NULL COMMENT '创建时间',
    `editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
    `editedDate` datetime DEFAULT NULL COMMENT '更新时间',
    `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB  COMMENT='制品资源表';

CREATE INDEX `idx_assets_artifactID_type_basename_deleted` ON `ops_artifact_assets` (`artifactLibID`, `type`, `basename`, `deleted`);
CREATE INDEX `idx_assets_blobID` ON `ops_artifact_assets` (`blobID`);
CREATE INDEX `idx_assets_deleted` ON `ops_artifact_assets` (`deleted`);
CREATE INDEX `idx_assets_groupID_basename_deleted` ON `ops_artifact_assets` (`groupID`, `basename`, `deleted`);
CREATE INDEX `idx_assets_versionID_basename_deleted` ON `ops_artifact_assets` (`versionID`, `basename`, `deleted`);


-- 7. 制品文件对象表
CREATE TABLE IF NOT EXISTS `ops_artifact_blobs` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `storageID` int unsigned NOT NULL DEFAULT 0 COMMENT '存储后端ID',
    `assetID` int unsigned NOT NULL DEFAULT 0 COMMENT '关联制品资源ID',
    `ref` varchar(255) NOT NULL DEFAULT '' COMMENT '存储引用',
    `size` bigint unsigned NOT NULL DEFAULT 0 COMMENT '文件大小',
    `downloads` int unsigned NOT NULL DEFAULT 0 COMMENT '下载次数',
    `metadata` varchar(500) NOT NULL DEFAULT '' COMMENT '文件元数据',
    `checksum` varchar(500) NOT NULL DEFAULT '' COMMENT '校验摘要',
    `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
    `createdDate` datetime DEFAULT NULL COMMENT '创建时间',
    `editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
    `editedDate` datetime DEFAULT NULL COMMENT '更新时间',
    `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
    PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB  COMMENT= '制品文件对象表';

CREATE UNIQUE INDEX `uk_blobs_ref_storageID` ON `ops_artifact_blobs` (`ref`, `storageID`);
CREATE INDEX `idx_blobs_assetID_deleted` ON `ops_artifact_blobs` (`assetID`, `deleted`);
CREATE INDEX `idx_blobs_deleted` ON `ops_artifact_blobs` (`deleted`);

-- 8. 制品元数据资源表
CREATE TABLE IF NOT EXISTS `ops_artifact_meta_assets` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`artifactLibID` int unsigned NOT NULL COMMENT '所属制品库ID',
	`type` varchar(50) NOT NULL DEFAULT '' COMMENT '制品类型:file/container/raw/helm/pypi/npm/maven/composer',
	`path` varchar(255) NOT NULL DEFAULT '' COMMENT '元数据路径',
	`contentType` varchar(100) NOT NULL DEFAULT '' COMMENT '内容类型',
	`rank` varchar(50) NOT NULL DEFAULT '' COMMENT '制品等级 main/sub',
	`blobID` int unsigned NOT NULL DEFAULT 0 COMMENT '关联文件对象ID',
	`checkSum` varchar(500) NOT NULL DEFAULT '' COMMENT '校验和',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB  COMMENT= '制品元数据资源表';

CREATE UNIQUE INDEX `uk_meta_artifactID_type_path` ON `ops_artifact_meta_assets` (`artifactLibID`, `type`, `path`);
CREATE INDEX `idx_meta_artifactID_deleted` ON `ops_artifact_meta_assets` (`artifactLibID`, `deleted`);
CREATE INDEX `idx_meta_blobID` ON `ops_artifact_meta_assets` (`blobID`);
CREATE INDEX `idx_meta_deleted` ON `ops_artifact_meta_assets` (`deleted`);