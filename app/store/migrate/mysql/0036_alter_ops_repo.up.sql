-- 重构 ops_repo：
-- 移除已废弃字段（empty/mutex）
-- 将 status 由 tinyint 改为 varchar(30) 并重置为 'active'
-- 调整字段顺序
ALTER TABLE `ops_repo`
    DROP COLUMN `empty`,
    DROP COLUMN `mutex`,
    DROP COLUMN `status`,
    MODIFY COLUMN `product` varchar(255) NOT NULL DEFAULT '' COMMENT '关联产品' AFTER `spaceID`,
    MODIFY COLUMN `forkID` int unsigned NULL DEFAULT NULL COMMENT '派生来源代码库ID' AFTER `gitUID`,
    MODIFY COLUMN `mirror` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否为镜像库（0=普通库，1=镜像库）' AFTER `forkID`,
    MODIFY COLUMN `defaultBranch` varchar(255) NOT NULL COMMENT '默认分支名' AFTER `mirror`,
    MODIFY COLUMN `acl` varchar(30) NOT NULL DEFAULT 'open' COMMENT '权限:private,open' AFTER `defaultBranch`,
    ADD COLUMN `status` varchar(30) NOT NULL DEFAULT 'active' COMMENT '代码库状态' AFTER `acl`,
    ADD COLUMN `providerID` int unsigned NOT NULL DEFAULT 0 COMMENT '源代码提供商ID' AFTER `mirror`,
    ADD COLUMN `connector` text COMMENT '连接器信息' AFTER `providerID`;
