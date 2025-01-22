CREATE DATABASE IF NOT EXISTS bbs_search;

USE bbs_search;

CREATE TABLE IF NOT EXISTS `post_resp`
(
    `id`          bigint unsigned NOT NULL AUTO_INCREMENT COMMENT '帖子ID',
    `code`        int             NOT NULL DEFAULT 0 COMMENT '状态码',
    `message`     varchar(255)    NOT NULL DEFAULT '' COMMENT '状态信息',
    `success`     boolean         NOT NULL DEFAULT FALSE COMMENT '是否成功',
    `data`        json            NOT NULL COMMENT '数据',
    `create_time` timestamp(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT '创建时间',
    `update_time` timestamp(3)    NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_code` (`code`),
    KEY `idx_create_time` (`create_time`),
    KEY `idx_update_time` (`update_time`)
) COMMENT '帖子响应';
