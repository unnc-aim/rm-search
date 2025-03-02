CREATE DATABASE IF NOT EXISTS rm_search;

USE rm_search;

CREATE TABLE IF NOT EXISTS `bbs_post`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '帖子ID',
    `code`        int             NOT NULL DEFAULT 0 COMMENT '状态码',
    `message`     varchar(256)    NOT NULL DEFAULT '' COMMENT '状态信息',
    `success`     boolean         NOT NULL DEFAULT FALSE COMMENT '是否成功',
    `data`        json            NOT NULL COMMENT '数据',
    `create_time` timestamp(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `update_time` timestamp(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_code` (`code`),
    KEY `idx_create_time` (`create_time`),
    KEY `idx_update_time` (`update_time`)
) COMMENT '论坛帖子';

CREATE TABLE IF NOT EXISTS `announce`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '公告ID',
    `found`       boolean         NOT NULL DEFAULT FALSE COMMENT '是否找到',
    `title`       varchar(256)    NOT NULL DEFAULT '' COMMENT '标题',
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

CREATE TABLE IF NOT EXISTS `attachment`
(
    `id`            bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '附件ID',
    `url`           varchar(512)    NOT NULL DEFAULT '' COMMENT 'URL',
    `name`          varchar(256)    NOT NULL DEFAULT '' COMMENT '名称',
    `size`          bigint unsigned NOT NULL DEFAULT 0 COMMENT '大小',
    `type`          varchar(64)     NOT NULL DEFAULT '' COMMENT '类型',
    `sha256`        char(64)        NOT NULL DEFAULT '' COMMENT 'SHA256',
    `content`       mediumtext      NOT NULL COMMENT '内容',
    `last_modified` bigint unsigned NOT NULL DEFAULT 0 COMMENT '最后修改时间',
    `create_time`   timestamp(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `update_time`   timestamp(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_url` (`url`),
    UNIQUE KEY `idx_sha256` (`sha256`)
) COMMENT '附件';
