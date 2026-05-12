-- 1. 删除 ops_artifact_groups 的唯一索引，改为普通索引。
-- 原因：软删记录（deleted=1）不应阻止同名 group 的重新创建。
DROP INDEX `uk_groups_artifactID_parentID_name` ON `ops_artifact_groups`;
CREATE INDEX `idx_groups_artifactID_parentID_name` ON `ops_artifact_groups` (`artifactID`, `parentID`, `name`);

-- 2. 删除 ops_artifact_assets 的唯一索引，改为普通索引。
-- 原因：同一 version 下同名 asset 软删后允许重新上传。
DROP INDEX `uk_assets_versionID_path_deleted` ON `ops_artifact_assets`;
CREATE INDEX `idx_assets_versionID_path_deleted` ON `ops_artifact_assets` (`versionID`, `path`, `deleted`);
