CREATE TABLE IF NOT EXISTS `ops_webhooks` (
	`id` INT unsigned NOT NULL AUTO_INCREMENT COMMENT 'Web钩子主键ID，自增',
	`spaceID` INT unsigned NULL DEFAULT NULL COMMENT '空间ID，关联spaces表id',
	`repoID` INT unsigned NULL DEFAULT NULL COMMENT '仓库ID，关联repository表id',
	`displayName` VARCHAR(255) NOT NULL COMMENT 'Web钩子显示名称',
	`desc` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Web钩子描述信息',
	`url` VARCHAR(255) NOT NULL COMMENT 'Web钩子回调地址',
	`secret` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Web钩子签名密钥',
	`isSSL` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '是否需要SSL证书验证：0=忽略，1=验证',
	`triggers` VARCHAR(500) NOT NULL COMMENT '触发条件等',
	`internal` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '是否为内部钩子：1=是，0=否，默认0',
	`authMethod` VARCHAR(90) NOT NULL DEFAULT '' COMMENT '认证方式（如none/basic/bearer等），默认空',
	`authHeader` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '认证请求头（如Basic Auth、Bearer Token等），默认空',
    `latestExecResult` VARCHAR(255) NULL DEFAULT NULL COMMENT '最近一次执行结果（成功/失败/异常等）',
    `enabled` TINYINT(1) NOT NULL DEFAULT '1' COMMENT '是否启用Web钩子：1=启用，0=禁用',
    `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
	`createdDate` datetime NULL COMMENT '创建时间',
    `editedBy` varchar(30) NOT NULL DEFAULT '',
	`editedDate` datetime NULL COMMENT '更新时间',
	`deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY (`id`) USING BTREE COMMENT '主键索引（Web钩子ID）'
)
ENGINE=InnoDB COMMENT='Web钩子配置表：存储Web钩子的基本信息、触发规则、认证方式、执行状态等';



CREATE TABLE IF NOT EXISTS `ops_webhook_executions` (
	`id`            INT unsigned     NOT NULL AUTO_INCREMENT COMMENT 'Web钩子执行记录主键ID，自增',
	`webhookID`     INT unsigned     NOT NULL COMMENT '关联的Web钩子ID',
	`triggerType`   VARCHAR(255)     NOT NULL COMMENT '触发类型（如push、merge、release等）',
	`eventID`       VARCHAR(255)     NOT NULL COMMENT '触发事件唯一标识,内部事件ID',
	`result`        TEXT             DEFAULT NULL COMMENT '执行结果（成功/失败/超时等状态描述）',
	`duration`      BIGINT           NOT NULL DEFAULT '0' COMMENT '执行耗时（毫秒），从请求发送到接收响应的总时长',
	`error`         TEXT             DEFAULT NULL COMMENT '执行错误信息（无错误则为空字符串）',
	`reqUrl`        VARCHAR(255)     NOT NULL COMMENT '回调请求URL',
	`reqHeaders`    TEXT     		 DEFAULT NULL COMMENT '回调请求头（JSON格式文本存储）',
	`reqBody`       TEXT             DEFAULT NULL COMMENT '回调请求体（JSON/表单等格式文本）',
	`respStatusCode` INT             NOT NULL COMMENT '回调响应状态码（如200、400、500等）',
	`respStatus`    VARCHAR(255)     NOT NULL COMMENT '回调响应状态描述（如OK、Bad Request等）',
	`respHeaders`   TEXT     		 DEFAULT NULL COMMENT '回调响应头（JSON格式文本存储）',
	`respBody`      TEXT             DEFAULT NULL COMMENT '回调响应体（JSON/文本等格式）',
	`retriggerID`   INT unsigned     NULL DEFAULT NULL COMMENT '关联重触发的执行记录ID（若为重试则指向原执行记录）',
	`retriggerable` TINYINT(1)       NOT NULL COMMENT '是否可重触发：1=可重触发，0=不可重触发',
	`createdDate`   datetime         NULL COMMENT '创建时间',
	`editedDate`    datetime         NULL COMMENT '更新时间',
	PRIMARY KEY (`id`) USING BTREE
)
ENGINE=InnoDB COMMENT='Web钩子执行记录表：存储每次Web钩子回调的请求、响应、执行状态、耗时等全量信息';

CREATE INDEX `idx_webhookID` ON `ops_webhook_executions` (`webhookID`) USING BTREE COMMENT 'Web钩子ID索引（优化按钩子维度查询执行记录）';
CREATE INDEX `idx_createdDate` ON `ops_webhook_executions` (`createdDate`) USING BTREE COMMENT '执行时间索引（优化按时间范围查询执行记录）';