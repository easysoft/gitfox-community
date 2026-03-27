ALTER TABLE `ops_scan_issues`
    ADD COLUMN `resolution` varchar(50) NULL DEFAULT NULL COMMENT '问题解决方案（bydesign/duplicate/external/fixed/notrepro/postponed/willnotfix/tostory）';

CREATE INDEX idx_resolution ON ops_scan_issues (`resolution`);
