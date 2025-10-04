CREATE TABLE `auths` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `auth_type` enum('email_password','gmail','facebook') DEFAULT 'email_password',
  `email` varchar(255) CHARACTER SET utf8mb4 NOT NULL,
  `salt` varchar(40) CHARACTER SET utf8mb4 DEFAULT NULL,
  `password` varchar(100) CHARACTER SET utf8mb4 DEFAULT NULL,
  `facebook_id` varchar(35) CHARACTER SET utf8mb4 DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`) USING BTREE,
  KEY `user_id` (`user_id`) USING BTREE,
  KEY `facebook_id` (`facebook_id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `users` (
  `id` int NOT NULL AUTO_INCREMENT,
  `first_name` varchar(30) CHARACTER SET utf8mb4 NOT NULL,
  `last_name` varchar(30) CHARACTER SET utf8mb4 NOT NULL,
  `email` varchar(255) CHARACTER SET utf8mb4 NOT NULL,
  `phone` varchar(30) DEFAULT NULL,
  `avatar` json DEFAULT NULL,
  `gender` enum('male','female','unknown') DEFAULT 'unknown',
  `dob` date DEFAULT NULL,
  `system_role` enum('sadmin','admin','user') DEFAULT 'user',
  `status` enum('active','waiting_verify','banned') DEFAULT 'active',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `email` (`email`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `chat_messages` (
    `id`        int NOT NULL AUTO_INCREMENT,
    `user_id`   VARCHAR(50) NOT NULL,
    `role`      ENUM('user','assistant') NOT NULL,
    `content`   TEXT NOT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE conversations (
  `id`            CHAR(36) NOT NULL,
  `user_id`       VARCHAR(50) NOT NULL,
  `org_id`        VARCHAR(50) NULL, -- multi-tenant
  `title`         VARCHAR(255) NULL,
  `status`        ENUM('active','archived','deleted') DEFAULT 'active',
  `created_at`    DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3), -- 3 độ chinh xác mili giây phục mục đích tính độ trễn (latency)
  `updated_at`    DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  INDEX idx_conv_user_created (user_id, created_at),
  INDEX idx_conv_org_created  (org_id, created_at)
);

CREATE TABLE messages (
  `id`               CHAR(36) PRIMARY KEY,
  `conversation_id`  CHAR(36) NOT NULL,
  `parent_id`        CHAR(36) NULL,                     -- cho thread/fork
  `role`             ENUM('user','assistant','tool') NOT NULL,
  `content`          MEDIUMTEXT NULL,                   -- text đã hiển thị
  `content_type`     ENUM('text','markdown','json') DEFAULT 'text',
  `meta`             JSON NULL,                         -- {language, safety_flags...}
  `tokens_input`     INT DEFAULT 0,
  `tokens_output`    INT DEFAULT 0,
  `latency_ms`       INT DEFAULT 0,                     -- độ trễ (latency) tính từ lúc nhận được message đến lúc hoàn thành
  `model_name`       VARCHAR(128) NULL,                 -- model AI
  `error_code`       VARCHAR(64) NULL,                  -- nếu fail
  `created_at`       DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  FOREIGN KEY (conversation_id) REFERENCES conversations(id),
  INDEX idx_msg_conv_created (conversation_id, created_at),
  INDEX idx_msg_role_created (role, created_at)
);

CREATE TABLE tool_runs ( -- các tool (Retriever, Crawler, external API) được gọi trong quá trình xử lý message
  `id`            CHAR(36) PRIMARY KEY,
  `message_id`    CHAR(36) NOT NULL,
  `tool_name`     VARCHAR(64) NOT NULL,                 -- retriever, calculator, crawler...
  `input_json`    JSON NULL,
  `output_json`   JSON NULL,
  `duration_ms`   INT DEFAULT 0,
  `success`       BOOLEAN DEFAULT TRUE,
  `created_at`    DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  FOREIGN KEY (message_id) REFERENCES messages(id),
  INDEX idx_tool_msg (message_id),
  INDEX idx_tool_name_created (tool_name, created_at)
);

-- file đính kèm trong message
CREATE TABLE message_attachments (
  `id`            CHAR(36) PRIMARY KEY,
  `message_id`    CHAR(36) NOT NULL,
  `filename`      VARCHAR(255) NOT NULL,
  `storage_key`   VARCHAR(512) NOT NULL,                -- S3/MinIO path
  `mime_type`     VARCHAR(64)  NULL,
  `sha256`        CHAR(64) NULL,
  `pages`         INT NULL,                              -- nếu PDF
  `meta`          JSON NULL,
  `created_at`    DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  FOREIGN KEY (message_id) REFERENCES messages(id),
  INDEX idx_att_msg (message_id),
  INDEX idx_att_sha (sha256)
);

-- sự kiện liên quan đến message
CREATE TABLE chat_events (
  `id`            CHAR(36) PRIMARY KEY,
  `conversation_id` CHAR(36) NOT NULL,
  `message_id`    CHAR(36) NULL,
  `type`          VARCHAR(64) NOT NULL, -- 'retry', 'rate_limited', 'timeout', 'parse_error'...
  `payload`       JSON NULL,
  `created_at`    DATETIME(3) DEFAULT CURRENT_TIMESTAMP(3),
  INDEX idx_evt_conv_created (conversation_id, created_at),
  INDEX idx_evt_type_created (type, created_at)
);