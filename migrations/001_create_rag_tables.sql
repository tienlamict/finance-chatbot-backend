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

