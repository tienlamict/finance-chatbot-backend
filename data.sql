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
  `mime_type`     VARCHAR(255)  NULL,
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

-- ==== RBAC tables ====
CREATE TABLE IF NOT EXISTS roles (
  id          INT AUTO_INCREMENT PRIMARY KEY,
  code        VARCHAR(64)  NOT NULL UNIQUE, -- 'sadmin','admin','user',...
  name        VARCHAR(128) NOT NULL,
  description VARCHAR(255) NULL,
  created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS permissions (
  id          INT AUTO_INCREMENT PRIMARY KEY,
  code        VARCHAR(128) NOT NULL UNIQUE, -- 'chat.send','chat.read','rbac.role.create',...
  name        VARCHAR(128) NOT NULL,
  description VARCHAR(255) NULL,
  created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS role_permissions (
  role_id       INT NOT NULL,
  permission_id INT NOT NULL,
  created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (role_id, permission_id),
  CONSTRAINT fk_rp_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
  CONSTRAINT fk_rp_perm FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS user_roles (
  user_id    INT NOT NULL,     -- map sang users.id (INT AUTO_INCREMENT)
  role_id    INT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, role_id),
  CONSTRAINT fk_ur_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_ur_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ==== seed cơ bản ====
INSERT INTO roles (code, name) VALUES
  ('sadmin','Super Admin'),
  ('admin','Admin'),
  ('user','User')
ON DUPLICATE KEY UPDATE name=VALUES(name);

INSERT INTO permissions (code, name) VALUES
  ('chat.send','Chat: send message'),
  ('chat.read','Chat: read history'),
  ('rbac.role.create','RBAC: create role'),
  ('rbac.role.read','RBAC: read roles'),
  ('rbac.role.update','RBAC: update role'),
  ('rbac.role.delete','RBAC: delete role'),
  ('rbac.permission.create','RBAC: create permission'),
  ('rbac.permission.read','RBAC: read permissions'),
  ('rbac.permission.assign','RBAC: assign permissions to role'),
  ('rbac.permission.revoke','RBAC: revoke permissions from role'),
  ('rbac.user.assign','RBAC: assign roles to user'),
  ('rbac.user.read','RBAC: read user roles')
ON DUPLICATE KEY UPDATE name=VALUES(name);

-- gán quyền mặc định cho role
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
WHERE r.code='user' AND p.code IN ('chat.send','chat.read');

INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r JOIN permissions p
WHERE r.code='admin' AND p.code IN ('chat.send','chat.read');

-- gán toàn bộ quyền cho sadmin
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.code='sadmin';

-- ==== seed superadmin user ====
-- Superadmin user (username: superadmin@system.local, password: superadmin)
INSERT INTO users (first_name, last_name, email, system_role, status)
VALUES ('Super', 'Admin', 'superadmin@system.local', 'sadmin', 'active')
ON DUPLICATE KEY UPDATE system_role='sadmin', status='active';

INSERT INTO auths (user_id, auth_type, email, salt, password)
SELECT u.id, 'email_password', 'superadmin@system.local', '3a26a16b04b2b1d8a910be76629a5a59', '$2a$10$tlyeAdbG1E4Kqw9itIG/tOLmOI4NBbd5D4BU8OMkYs/lDFs6v4yU6'
FROM users u WHERE u.email='superadmin@system.local'
ON DUPLICATE KEY UPDATE salt=VALUES(salt), password=VALUES(password);

-- gán role sadmin cho superadmin user
INSERT IGNORE INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON r.code='sadmin'
WHERE u.email='superadmin@system.local';

-- map người dùng vào role 'user' 
INSERT IGNORE INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON r.code='user';


-- Create rag_documents table
CREATE TABLE IF NOT EXISTS `rag_documents` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `doc_id` varchar(100) NOT NULL COMMENT 'Unique document identifier',
  `collection` varchar(100) NOT NULL COMMENT 'RAG collection name',
  `stored_path` varchar(500) NOT NULL COMMENT 'Path returned by RAG API',
  `filename` varchar(255) NOT NULL COMMENT 'Original filename',
  `mime_type` varchar(100) DEFAULT NULL COMMENT 'File MIME type',
  `sha256` varchar(64) DEFAULT NULL COMMENT 'File SHA-256 hash',
  `total_pages` int(11) DEFAULT NULL COMMENT 'Number of pages (for PDFs)',
  `title` varchar(500) DEFAULT NULL COMMENT 'Document title from metadata',
  `author` varchar(255) DEFAULT NULL COMMENT 'Document author from metadata',
  `uploaded_by` varchar(50) NOT NULL COMMENT 'User UID who uploaded',
  `minio_key` varchar(500) DEFAULT NULL COMMENT 'Original MinIO storage key',
  `uploaded_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Upload timestamp',
  `status` int(11) NOT NULL DEFAULT '1' COMMENT 'Status: 1=active, 0=deleted',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_doc_id` (`doc_id`),
  KEY `idx_collection` (`collection`),
  KEY `idx_uploaded_by` (`uploaded_by`),
  KEY `idx_sha256` (`sha256`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='RAG documents metadata';

-- Create rag_query_logs table
CREATE TABLE IF NOT EXISTS `rag_query_logs` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `query_text` text NOT NULL COMMENT 'Query text',
  `user_id` varchar(50) DEFAULT NULL COMMENT 'User UID who made query',
  `conversation_id` varchar(100) DEFAULT NULL COMMENT 'Associated conversation',
  `collection` varchar(100) DEFAULT NULL COMMENT 'Collection queried',
  `k` int(11) NOT NULL COMMENT 'Number of results requested',
  `returned_docs` json DEFAULT NULL COMMENT 'Array of returned doc_ids',
  `latency_ms` int(11) DEFAULT NULL COMMENT 'Query latency in milliseconds',
  `status` int(11) NOT NULL DEFAULT '1' COMMENT 'Status: 1=active, 0=deleted',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_conversation_id` (`conversation_id`),
  KEY `idx_collection` (`collection`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='RAG query logs for analytics';