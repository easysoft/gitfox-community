-- ============================================================
-- 0037: 重建制品库相关表
--   废弃 0034/0036 形态，按 app/store/migrate/mysql/制品库设计.md 重新建表。
--   注意：这是破坏性迁移，会丢弃 ops_artifact / ops_artifact_groups / ops_artifact_packages /
--        ops_artifact_versions / ops_artifact_assets / ops_artifact_blobs /
--        ops_artifact_tree_nodes / ops_artifact_meta_assets 全部数据。
-- ============================================================

-- 1. 全量清表（按外键无约束，顺序无强制要求；先末梢后枝干以便阅读）
DROP TABLE IF EXISTS `ops_artifact_meta_assets`;
DROP TABLE IF EXISTS `ops_artifact_blobs`;
DROP TABLE IF EXISTS `ops_artifact_assets`;
DROP TABLE IF EXISTS `ops_artifact_versions`;
DROP TABLE IF EXISTS `ops_artifact_packages`;
DROP TABLE IF EXISTS `ops_artifact_groups`;
DROP TABLE IF EXISTS `ops_artifact_tree_nodes`;
DROP TABLE IF EXISTS `ops_artifact_libs`;
DROP TABLE IF EXISTS `ops_artifact`;


-- 2. 制品库主表
CREATE TABLE IF NOT EXISTS `ops_artifact_libs` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `spaceID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属空间ID，0表示全局级',
    `repoID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属代码库ID，0表示非代码库级',
    `name` varchar(100) NOT NULL COMMENT '制品库名称',
    `desc` varchar(500) NOT NULL DEFAULT '' COMMENT '制品库描述',
    `type` varchar(50) NOT NULL COMMENT '制品类型:file/docker/raw/helm/pypi/npm/maven/composer',
    `scope` varchar(30) NOT NULL DEFAULT '' COMMENT '制品库类型:global/space/repo',
    `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
    `createdDate` datetime DEFAULT NULL COMMENT '创建时间',
    `editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
    `editedDate` datetime DEFAULT NULL COMMENT '更新时间',
    `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_artifact_spaceID_repoID_name` (`spaceID`, `repoID`, `name`),
    KEY `idx_artifact_spaceID_deleted` (`spaceID`, `deleted`),
    KEY `idx_artifact_repoID_deleted` (`repoID`, `deleted`),
    KEY `idx_artifact_spaceID_repoID_deleted` (`spaceID`, `repoID`, `deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='制品库主表（支持全局/空间/代码库三级作用域）';


-- 3. 制品树节点表
CREATE TABLE IF NOT EXISTS `ops_artifact_tree_nodes` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `artifactLibID` int unsigned NOT NULL COMMENT '所属制品库ID',
    `parentID` int unsigned NOT NULL DEFAULT 0 COMMENT '父级节点ID，0表示根节点',
    `nodeName` varchar(255) NOT NULL DEFAULT '' COMMENT '节点名称',
    `nodePath` varchar(255) NOT NULL DEFAULT '' COMMENT '节点路径',
    `linkTable` varchar(20) NOT NULL DEFAULT '' COMMENT '关联表',
    `linkRecord` int unsigned DEFAULT NULL COMMENT '关联表ID',
    `type` varchar(50) NOT NULL DEFAULT '' COMMENT '制品类型:file/docker/raw/helm/pypi/npm/maven/composer',
    `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
    `createdDate` datetime DEFAULT NULL COMMENT '创建时间',
    `editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
    `editedDate` datetime DEFAULT NULL COMMENT '更新时间',
    `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_nodes_artifactLibID_nodePath_type` (`artifactLibID`, `nodePath`, `type`),
    KEY `idx_nodes_parentID_deleted` (`parentID`, `deleted`),
    KEY `idx_nodes_artifactLibID_linkTable_linkRecord_deleted` (`artifactLibID`, `linkTable`, `linkRecord`, `deleted`),
    KEY `idx_nodes_artifactLibID_type_deleted` (`artifactLibID`, `type`, `deleted`),
    KEY `idx_nodes_deleted` (`deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='制品树节点表';


-- 4. 制品分组表
CREATE TABLE IF NOT EXISTS `ops_artifact_groups` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `artifactLibID` int unsigned NOT NULL COMMENT '所属制品库ID',
    `parentID` int unsigned NOT NULL DEFAULT 0 COMMENT '父级分组ID，0表示根分组',
    `name` varchar(255) NOT NULL DEFAULT '' COMMENT '分组名称',
    `type` varchar(50) NOT NULL DEFAULT '' COMMENT '制品类型:file/docker/raw/helm/pypi/npm/maven/composer',
    `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
    `createdDate` datetime DEFAULT NULL COMMENT '创建时间',
    `editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
    `editedDate` datetime DEFAULT NULL COMMENT '更新时间',
    `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_groups_artifactLibID_deleted` (`artifactLibID`, `deleted`),
    KEY `idx_groups_artifactLibID_parentID_deleted` (`artifactLibID`, `parentID`, `deleted`),
    KEY `idx_groups_parentID_deleted` (`parentID`, `deleted`),
    KEY `idx_groups_artifactLibID_parentID_name` (`artifactLibID`, `parentID`, `name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='制品分组表';


-- 5. 制品包表
CREATE TABLE IF NOT EXISTS `ops_artifact_packages` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `artifactLibID` int unsigned NOT NULL COMMENT '所属制品库ID',
    `groupID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属分组ID',
    `namespace` varchar(255) NOT NULL DEFAULT '' COMMENT '命名空间',
    `name` varchar(100) NOT NULL DEFAULT '' COMMENT '制品包名称',
    `type` varchar(50) NOT NULL DEFAULT '' COMMENT '制品类型:file/docker/raw/helm/pypi/npm/maven/composer',
    `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
    `createdDate` datetime DEFAULT NULL COMMENT '创建时间',
    `editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
    `editedDate` datetime DEFAULT NULL COMMENT '更新时间',
    `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_pkgs_artifactLibID_name_namespace_type` (`artifactLibID`, `name`, `namespace`, `type`),
    KEY `idx_pkgs_artifactLibID_type_deleted` (`artifactLibID`, `type`, `deleted`),
    KEY `idx_pkgs_artifactLibID_namespace_deleted` (`artifactLibID`, `namespace`, `deleted`),
    KEY `idx_pkgs_groupID_deleted` (`groupID`, `deleted`),
    KEY `idx_pkgs_deleted` (`deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='制品包表';


-- 6. 制品版本表
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
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_vers_artifactLibID_packageID_version` (`artifactLibID`, `packageID`, `version`),
    KEY `idx_vers_artifactLibID_deleted` (`artifactLibID`, `deleted`),
    KEY `idx_vers_packageID_deleted` (`packageID`, `deleted`),
    KEY `idx_vers_createdDate` (`createdDate`),
    KEY `idx_vers_editedDate` (`editedDate`),
    KEY `idx_vers_deleted` (`deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='制品版本表';


-- 7. 制品资源表
CREATE TABLE IF NOT EXISTS `ops_artifact_assets` (
    `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `artifactLibID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属制品库ID',
    `groupID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属分组ID，0表示无分组',
    `packageID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属制品包ID',
    `versionID` int unsigned NOT NULL DEFAULT 0 COMMENT '所属版本ID，0表示无版本关联',
    `blobID` int unsigned NOT NULL DEFAULT 0 COMMENT '关联文件对象ID',
    `type` varchar(50) NOT NULL DEFAULT '' COMMENT '制品类型:file/docker/raw/helm/pypi/npm/maven/composer',
    `basename` varchar(255) NOT NULL DEFAULT '' COMMENT '制品名称',
    `contentType` varchar(100) NOT NULL DEFAULT '' COMMENT '内容类型',
    `kind` varchar(50) NOT NULL DEFAULT '' COMMENT '制品类型',
    `checksum` varchar(500) NOT NULL DEFAULT '' COMMENT '校验摘要',
    `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
    `createdDate` datetime DEFAULT NULL COMMENT '创建时间',
    `editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
    `editedDate` datetime DEFAULT NULL COMMENT '更新时间',
    `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
    PRIMARY KEY (`id`) USING BTREE,
    KEY `idx_assets_artifactID_type_basename_deleted` (`artifactLibID`, `type`, `basename`, `deleted`),
    KEY `idx_assets_blobID` (`blobID`),
    KEY `idx_assets_deleted` (`deleted`),
    KEY `idx_assets_groupID_basename_deleted` (`groupID`, `basename`, `deleted`),
    KEY `idx_assets_versionID_basename_deleted` (`versionID`, `basename`, `deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='制品资源表';


-- 8. 制品文件对象表
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
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE KEY `uk_blobs_ref_storageID` (`ref`, `storageID`),
    KEY `idx_blobs_assetID_deleted` (`assetID`, `deleted`),
    KEY `idx_blobs_deleted` (`deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='制品文件对象表';

-- 9. 制品元数据资源表
CREATE TABLE IF NOT EXISTS `ops_artifact_meta_assets` (
	`id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
	`artifactLibID` int unsigned NOT NULL COMMENT '所属制品库ID',
	`type` varchar(50) NOT NULL DEFAULT '' COMMENT '制品类型:file/docker/raw/helm/pypi/npm/maven/composer',
	`path` varchar(255) NOT NULL DEFAULT '' COMMENT '元数据路径',
	`contentType` varchar(100) NOT NULL DEFAULT '' COMMENT '内容类型',
	`kind` varchar(50) NOT NULL DEFAULT '' COMMENT '元数据类型',
	`blobID` int unsigned NOT NULL DEFAULT 0 COMMENT '关联文件对象ID',
	`checkSum` varchar(500) NOT NULL DEFAULT '' COMMENT '校验和',
	`createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
	`editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='制品元数据资源表';

CREATE UNIQUE INDEX `uk_meta_artifactID_type_path` ON `ops_artifact_meta_assets` (`artifactLibID`, `type`, `path`);
CREATE INDEX `idx_meta_artifactID_deleted` ON `ops_artifact_meta_assets` (`artifactLibID`, `deleted`);
CREATE INDEX `idx_meta_blobID` ON `ops_artifact_meta_assets` (`blobID`);
CREATE INDEX `idx_meta_deleted` ON `ops_artifact_meta_assets` (`deleted`);