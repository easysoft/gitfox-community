-- 代码库表（达梦适配版：小驼峰命名）
CREATE TABLE IF NOT EXISTS "ops_repo" (
	"id" INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '代码库ID',
	"spaceID" INT NOT NULL COMMENT '空间ID',
	"identifier" VARCHAR(255) NOT NULL COMMENT '代码库唯一标识',
	"desc" VARCHAR(500) NOT NULL DEFAULT '' COMMENT '代码库描述',
	"gitUID" CHAR(42) NOT NULL COMMENT 'Git仓库唯一标识',
	"defaultBranch" VARCHAR(255) NOT NULL COMMENT '默认分支名',
	"forkID" INT UNSIGNED NULL DEFAULT NULL COMMENT '派生来源代码库ID',
	"empty" TINYINT NOT NULL DEFAULT 0 COMMENT '是否为空库',
	"acl" VARCHAR(30) NOT NULL DEFAULT 'open' COMMENT '权限:private,open',
	"mirror" TINYINT NOT NULL DEFAULT 0 COMMENT '是否为镜像库（0=普通库，1=镜像库）',
	"status" TINYINT NOT NULL DEFAULT 0 COMMENT '代码库状态',
	"mutex" INT NOT NULL DEFAULT 0 COMMENT '乐观锁版本号',
	"createdBy" VARCHAR(30) NOT NULL DEFAULT '' COMMENT '创建人',
	"createdDate" DATETIME NULL COMMENT '创建时间',
	"editedBy" VARCHAR(30) NOT NULL DEFAULT '' COMMENT '更新人',
	"editedDate" DATETIME NULL COMMENT '更新时间',
	"deleted" TINYINT NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
	PRIMARY KEY ("id") USING BTREE
) ENGINE INNODB CHARACTER SET UTF8MB4 COMMENT '代码库表';

-- 唯一索引：同一空间内代码库标识唯一（小驼峰索引名+字段名）
CREATE UNIQUE INDEX "uk_repo_space_id" ON "ops_repo" ("identifier", "spaceID");

-- 唯一索引：Git仓库唯一标识（修正原笔误：uid→gitUID，小驼峰字段名）
CREATE UNIQUE INDEX "uk_gituid" ON "ops_repo" ("gitUID");

-- 普通索引：软删除字段（小驼峰索引名+字段名）
CREATE INDEX "idx_deleted" ON "ops_repo" ("deleted");