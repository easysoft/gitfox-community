ALTER TABLE `ops_scan_issues`
    ADD COLUMN `resolved` datetime NULL COMMENT '问题解决时间',
    ADD COLUMN `closed` datetime NULL COMMENT '问题关闭时间',
    ADD COLUMN `ignored` bigint NOT NULL DEFAULT -1 COMMENT '问题忽略到期时间';