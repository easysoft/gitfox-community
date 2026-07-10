ALTER TABLE `ops_space`
	ADD COLUMN `code` VARCHAR(50) NOT NULL DEFAULT '' COMMENT '空间唯一标识' AFTER `name`;

CREATE UNIQUE INDEX `uk_code` ON `ops_space`(`code`);
DROP INDEX `uk_name` ON `ops_space`;
