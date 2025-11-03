-- 空间表
-- 注意：达梦数据库中使用双引号包裹的标识符可以保留大小写
CREATE TABLE "ops_space" (
  "id" INT IDENTITY(1,1) NOT NULL PRIMARY KEY,
  "name" VARCHAR(200) NOT NULL,
  "owner" VARCHAR(30) NOT NULL DEFAULT '',
  "acl" VARCHAR(30) NOT NULL DEFAULT 'open',
  "desc" VARCHAR(500) NOT NULL DEFAULT '',
  "createdBy" VARCHAR(30) NOT NULL DEFAULT '',
  "createdDate" DATETIME DEFAULT NULL,
  "editedBy" VARCHAR(30) NOT NULL DEFAULT '',
  "editedDate" DATETIME DEFAULT NULL,
  "deleted" TINYINT NOT NULL DEFAULT 0,
  CONSTRAINT "uk_name" UNIQUE ("name")
);



