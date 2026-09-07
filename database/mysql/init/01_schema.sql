-- =============================================================
-- 简聊 SimpleChat 数据库初始化脚本
-- 存储引擎: InnoDB  字符集: utf8mb4  排序规则: utf8mb4_unicode_ci
-- 说明: 所有外键采用逻辑外键(应用层保证), 不使用物理外键约束
-- =============================================================
USE simplechat;

SET NAMES utf8mb4;

-- 1. 用户表
CREATE TABLE IF NOT EXISTS `user` (
  `user_id`       BIGINT       NOT NULL COMMENT '用户ID(雪花算法)',
  `account`       VARCHAR(32)  NOT NULL COMMENT '登录账号(字母数字, 唯一)',
  `phone`         VARCHAR(20)  DEFAULT NULL COMMENT '手机号',
  `email`         VARCHAR(100) DEFAULT NULL COMMENT '邮箱',
  `password_hash` VARCHAR(128) NOT NULL COMMENT 'bcrypt密码哈希',
  `nickname`      VARCHAR(20)  NOT NULL COMMENT '昵称',
  `avatar_url`    VARCHAR(255) DEFAULT NULL COMMENT '头像URL',
  `signature`     VARCHAR(50)  DEFAULT NULL COMMENT '个性签名',
  `status`        TINYINT      NOT NULL DEFAULT 0 COMMENT '在线状态 0=离线 1=在线 2=忙碌',
  `created_at`    DATETIME     NOT NULL COMMENT '注册时间',
  `last_active`   DATETIME     DEFAULT NULL COMMENT '最后活跃时间',
  `is_verified`   TINYINT      NOT NULL DEFAULT 0 COMMENT '是否已验证 0=否 1=是',
  `fail_count`    INT          NOT NULL DEFAULT 0 COMMENT '连续登录失败次数',
  `lock_until`    DATETIME     DEFAULT NULL COMMENT '锁定截止时间',
  PRIMARY KEY (`user_id`),
  UNIQUE KEY `uk_account` (`account`),
  UNIQUE KEY `uk_phone` (`phone`),
  UNIQUE KEY `uk_email` (`email`),
  KEY `idx_nickname` (`nickname`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 2. 好友关系表
CREATE TABLE IF NOT EXISTS `friendship` (
  `id`          BIGINT   NOT NULL AUTO_INCREMENT COMMENT '关系ID',
  `from_id`     BIGINT   NOT NULL COMMENT '发起方用户ID',
  `to_id`       BIGINT   NOT NULL COMMENT '接收方用户ID',
  `remark`      VARCHAR(64) DEFAULT NULL COMMENT '备注名',
  `status`      TINYINT  NOT NULL DEFAULT 0 COMMENT '0=申请中 1=正常 2=已删除 3=拉黑',
  `black`       TINYINT  NOT NULL DEFAULT 0 COMMENT '是否黑名单 0=否 1=是',
  `create_time` BIGINT   NOT NULL COMMENT '创建时间戳(ms)',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_from_to` (`from_id`, `to_id`),
  KEY `idx_from_id` (`from_id`),
  KEY `idx_to_id` (`to_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友关系表';

-- 3. 群组表
CREATE TABLE IF NOT EXISTS `group` (
  `group_id`     BIGINT       NOT NULL COMMENT '群组ID(雪花算法)',
  `group_name`   VARCHAR(30)  NOT NULL COMMENT '群名称',
  `creator_id`   BIGINT       NOT NULL COMMENT '群主ID',
  `avatar_url`   VARCHAR(255) DEFAULT NULL COMMENT '群头像URL',
  `announcement` TEXT         COMMENT '群公告(<=200字符)',
  `member_count` INT          NOT NULL DEFAULT 0 COMMENT '成员数量(冗余)',
  `created_at`   DATETIME     NOT NULL COMMENT '创建时间',
  `is_dismissed` TINYINT      NOT NULL DEFAULT 0 COMMENT '是否已解散 0=否 1=是',
  PRIMARY KEY (`group_id`),
  KEY `idx_creator_id` (`creator_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组表';

-- 4. 群组成员表
CREATE TABLE IF NOT EXISTS `group_member` (
  `id`        BIGINT   NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `group_id`  BIGINT   NOT NULL COMMENT '群组ID',
  `user_id`   BIGINT   NOT NULL COMMENT '用户ID',
  `role`      TINYINT  NOT NULL DEFAULT 0 COMMENT '0=普通成员 1=管理员 2=群主',
  `joined_at` DATETIME NOT NULL COMMENT '加入时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_user` (`group_id`, `user_id`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组成员表';

-- 5. 会话表
CREATE TABLE IF NOT EXISTS `conversation` (
  `conversation_id`   BIGINT      NOT NULL COMMENT '会话ID(雪花算法)',
  `type`              TINYINT     NOT NULL COMMENT '0=单聊 1=群聊',
  `peer_user_id`      BIGINT      DEFAULT NULL COMMENT '单聊对方用户ID',
  `group_id`          BIGINT      DEFAULT NULL COMMENT '群聊对应群ID',
  `last_msg_time`     DATETIME    DEFAULT NULL COMMENT '最后一条消息时间',
  `last_msg_id`       BIGINT      DEFAULT NULL COMMENT '最后一条消息ID',
  `last_msg_preview`  VARCHAR(100) DEFAULT NULL COMMENT '最后一条消息预览',
  `created_at`        DATETIME    NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`conversation_id`),
  KEY `idx_group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会话表';

-- 6. 会话成员表
CREATE TABLE IF NOT EXISTS `conversation_member` (
  `id`                BIGINT   NOT NULL AUTO_INCREMENT COMMENT '关联ID',
  `conversation_id`   BIGINT   NOT NULL COMMENT '会话ID',
  `user_id`           BIGINT   NOT NULL COMMENT '用户ID',
  `unread_count`      INT      NOT NULL DEFAULT 0 COMMENT '未读消息数',
  `is_top`            TINYINT  NOT NULL DEFAULT 0 COMMENT '是否置顶 0=否 1=是',
  `is_mute`           TINYINT  NOT NULL DEFAULT 0 COMMENT '是否免打扰 0=否 1=是',
  `joined_at`         DATETIME NOT NULL COMMENT '加入时间',
  `last_read_msg_id`  BIGINT   DEFAULT NULL COMMENT '最后已读消息ID',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_conversation_user` (`conversation_id`, `user_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会话成员表';

-- 7. 消息表 (按月范围分区预留)
CREATE TABLE IF NOT EXISTS `message` (
  `message_id`     BIGINT      NOT NULL COMMENT '消息ID(雪花算法)',
  `conversation_id` BIGINT      NOT NULL COMMENT '所属会话ID',
  `sender_id`      BIGINT      NOT NULL COMMENT '发送者ID',
  `type`           TINYINT     NOT NULL COMMENT '0=文本 1=图片 2=文件 3=表情 4=系统 5=撤回',
  `content`        TEXT        COMMENT '消息内容(文本消息)',
  `media_url`      VARCHAR(255) DEFAULT NULL COMMENT '媒体文件URL(图片/文件)',
  `file_id`        BIGINT      DEFAULT NULL COMMENT '关联文件记录ID',
  `sent_time`      DATETIME    NOT NULL COMMENT '发送时间',
  `status`         TINYINT     NOT NULL DEFAULT 0 COMMENT '0=发送中 1=已发送 2=已送达 3=已读',
  `is_recalled`    TINYINT     NOT NULL DEFAULT 0 COMMENT '是否已撤回 0=否 1=是',
  PRIMARY KEY (`message_id`),
  KEY `idx_conversation_id` (`conversation_id`),
  KEY `idx_sender_id` (`sender_id`),
  KEY `idx_sent_time` (`sent_time`),
  KEY `idx_conversation_time` (`conversation_id`, `sent_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息表';

-- 8. 消息状态表
CREATE TABLE IF NOT EXISTS `message_status` (
  `id`         BIGINT   NOT NULL AUTO_INCREMENT COMMENT '状态ID',
  `message_id` BIGINT   NOT NULL COMMENT '消息ID',
  `user_id`    BIGINT   NOT NULL COMMENT '接收方用户ID',
  `status`     TINYINT  NOT NULL DEFAULT 0 COMMENT '0=未读 1=已读',
  `read_time`  DATETIME DEFAULT NULL COMMENT '阅读时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_message_user` (`message_id`, `user_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_message_id` (`message_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息状态表';

-- 9. 离线消息表
CREATE TABLE IF NOT EXISTS `offline_message` (
  `id`          BIGINT   NOT NULL AUTO_INCREMENT COMMENT '离线消息ID',
  `user_id`     BIGINT   NOT NULL COMMENT '接收方用户ID',
  `message_id`  BIGINT   NOT NULL COMMENT '消息ID',
  `del_flag`    TINYINT  NOT NULL DEFAULT 0 COMMENT '0=未拉取 1=已拉取',
  `create_time` DATETIME NOT NULL COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_del_flag` (`del_flag`),
  UNIQUE KEY `uk_user_message` (`user_id`, `message_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='离线消息表';

-- 10. 文件记录表
CREATE TABLE IF NOT EXISTS `file_record` (
  `file_id`     BIGINT      NOT NULL COMMENT '文件ID(雪花算法)',
  `file_name`   VARCHAR(255) NOT NULL COMMENT '原始文件名',
  `file_size`   BIGINT      NOT NULL COMMENT '文件大小(字节)',
  `file_url`    VARCHAR(255) NOT NULL COMMENT '存储URL',
  `file_md5`    VARCHAR(32) DEFAULT NULL COMMENT '文件MD5(用于去重)',
  `mime_type`   VARCHAR(50) DEFAULT NULL COMMENT '文件MIME类型',
  `uploader_id` BIGINT      NOT NULL COMMENT '上传者ID',
  `upload_time` DATETIME    NOT NULL COMMENT '上传时间',
  PRIMARY KEY (`file_id`),
  KEY `idx_file_md5` (`file_md5`),
  KEY `idx_uploader_id` (`uploader_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件记录表';