-- 0038: 将 ops_artifact_assets.kind 改名为 rank
ALTER TABLE `ops_artifact_assets`
    CHANGE COLUMN `kind` `rank` varchar(50) NOT NULL DEFAULT '' COMMENT '制品等级 main/sub';
