ALTER TABLE `ops_space`
	ADD COLUMN `auth` VARCHAR(30) NOT NULL DEFAULT 'extend' COMMENT '空间人员权限定义：extend继承，reset重定义' AFTER `acl`;
