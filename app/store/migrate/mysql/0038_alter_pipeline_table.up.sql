-- 重构 ops_pipeline：
-- 新增 engine（默认 gitfox）、providerID（默认 0）
-- 将 type 重命名为 scope 并调整类型为 varchar(30)
-- 调整 status 类型为 varchar(30)
-- 新增 customParam、lastExec、lastResult
ALTER TABLE `ops_pipeline`
    ADD COLUMN `engine` varchar(30) NOT NULL DEFAULT 'gitfox' COMMENT '引擎: gitlab, jenkins, gitfox' AFTER `name`,
    ADD COLUMN `providerID` int unsigned NOT NULL DEFAULT 0 COMMENT '关联的外部服务ID' AFTER `engine`,
    CHANGE COLUMN `type` `scope` varchar(30) NOT NULL COMMENT '流水线类型：系统/空间/仓库',
    MODIFY COLUMN `status` varchar(30) NOT NULL COMMENT '流水线状态：草稿/激活',
    ADD COLUMN `customParam` TEXT DEFAULT NULL COMMENT 'Jenkins,GitLab 流水线执行传参' AFTER `yamlPath`,
    ADD COLUMN `externalPipeline` varchar(128) NOT NULL DEFAULT '' COMMENT 'Jenkins/GitLab pipeline' AFTER `customParam`,
    ADD COLUMN `lastExec` datetime DEFAULT NULL COMMENT '最后一次执行时间' AFTER `customParam`,
    ADD COLUMN `lastResult` varchar(30) NOT NULL DEFAULT '' COMMENT '最后一次执行状态' AFTER `lastExec`;

-- 重构 ops_pipeline_executions：
-- 将 started 重命名为 startedDate
-- 将 finished 重命名为 finishedDate
-- 新增 queue（Jenkins 队列号）、logs（流水线日志）
ALTER TABLE `ops_pipeline_executions`
    CHANGE COLUMN `started` `startedDate` datetime DEFAULT NULL COMMENT '执行开始时间',
    CHANGE COLUMN `finished` `finishedDate` datetime DEFAULT NULL COMMENT '执行结束时间',
    ADD COLUMN `queue` int unsigned NOT NULL DEFAULT 0 COMMENT 'Jenkins队列号，用于后续状态查询和去重' AFTER `error`,
    ADD COLUMN `logs` longtext DEFAULT NULL COMMENT 'Jenkins,GitLab的流水线日志，支持 longtext（最大 4GB）' AFTER `queue`;

-- 重构 ops_triggers：
-- 将 trigger 重命名为 event
-- 新增 comment（branch_updated 事件时的提交关键词匹配）
ALTER TABLE `ops_triggers`
    CHANGE COLUMN `trigger` `event` varchar(500) NOT NULL DEFAULT '' COMMENT '事件触发动作',
    ADD COLUMN `comment` varchar(255) NOT NULL DEFAULT '' COMMENT '当事件为 branch_updated 时进行提交关键词匹配' AFTER `event`;
