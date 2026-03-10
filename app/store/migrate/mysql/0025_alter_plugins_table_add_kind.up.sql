ALTER TABLE `ops_plugins`
	ADD COLUMN `kind` varchar(255) NOT NULL DEFAULT '' COMMENT '插件 kind' AFTER `type`;
