CREATE DATABASE IF NOT EXISTS rm_search;

USE rm_search;

CREATE TABLE IF NOT EXISTS `bbs_post`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '帖子ID',
    `code`        int             NOT NULL DEFAULT 0 COMMENT '状态码',
    `message`     varchar(255)    NOT NULL DEFAULT '' COMMENT '状态信息',
    `success`     boolean         NOT NULL DEFAULT FALSE COMMENT '是否成功',
    `data`        json            NOT NULL COMMENT '数据',
    `create_time` timestamp(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `update_time` timestamp(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_code` (`code`)
) COMMENT '论坛帖子';

CREATE TABLE IF NOT EXISTS `announce`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '公告ID',
    `found`       boolean         NOT NULL DEFAULT FALSE COMMENT '是否找到',
    `title`       varchar(255)    NOT NULL DEFAULT '' COMMENT '标题',
    `date`        date            NOT NULL DEFAULT '0001-01-01' COMMENT '日期',
    `context`     mediumtext      NOT NULL COMMENT '上下文',
    `content`     mediumtext      NOT NULL COMMENT '内容',
    `attachments` json            NOT NULL COMMENT '附件',
    `create_time` timestamp(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `update_time` timestamp(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_found` (`found`),
    KEY `idx_date` (`date`)
) COMMENT '公告';
