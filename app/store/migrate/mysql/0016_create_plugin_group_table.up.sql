CREATE TABLE IF NOT EXISTS `ops_plugin_group` (
  `id` int unsigned NOT NULL AUTO_INCREMENT COMMENT '插件分组ID，自增主键',
  `name` varchar(255) NOT NULL COMMENT '插件分组名称',
  `desc` varchar(255) NOT NULL DEFAULT '' COMMENT '插件分组描述信息',
  `deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '删除标识，0-未删除，1-已删除',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB COMMENT='插件分组基本信息表';

INSERT INTO `ops_plugin_group` (`name`, `desc`) VALUES   
    ('default', '默认'),                                                                              
    ('build', '构建'),                                                                                                                    
    ('test', '测试'),                                                                                                                     
    ('deploy', '部署'),                                                                                                                   
    ('artifact', '制品');
