CREATE TABLE `ops_provider` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `type` varchar(30) NOT NULL DEFAULT '' COMMENT 'GitLab,Gitea,Gogs,Subversion,GitHub,Jenkins',
  `name` varchar(50) NOT NULL DEFAULT '' COMMENT '外部服务名称',
  `url` varchar(255) NOT NULL DEFAULT '' COMMENT '外部服务地址',
  `token` varchar(255) NOT NULL DEFAULT '' COMMENT '外部服务调用Token，如果 type=Jenkins 则存入 base64_encode(username:token)',
  `createdBy` varchar(30) NOT NULL DEFAULT '' COMMENT '创建人',
  `createdDate` datetime NULL COMMENT '创建时间',
  `editedBy` varchar(30) NOT NULL DEFAULT '' COMMENT '更新人',
  `editedDate` datetime NULL COMMENT '更新时间',
  `deleted` tinyint unsigned NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='外部服务表';