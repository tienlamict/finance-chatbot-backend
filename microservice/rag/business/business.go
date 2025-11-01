package business

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"finance-chatbot/addon/common"
	"finance-chatbot/microservice/rag/entity"
	"finance-chatbot/microservice/rag/repository/mysql"
	"finance-chatbot/microservice/rag/repository/rpc"
)

// RAGBusiness defines the business logic interface for RAG operations
type RAGBusiness interface {
	// UploadDocumentForRAG downloads from MinIO and uploads to RAG service
	UploadDocumentForRAG(ctx context.Context, req *entity.CreateRAGDocumentRequest, userID string) (*entity.RAGDocument, error)

	// UploadDocumentDirect uploads a file directly to RAG service (multipart)
	UploadDocumentDirect(ctx context.Context, filePath string, docID, collection string, userID string) (*entity.RAGDocument, error)

	// QueryRAG queries the RAG service and returns mapped results
	QueryRAG(ctx context.Context, req *entity.QueryRAGRequest, userID, conversationID string) ([]*entity.RAGQueryResult, error)

	// GetDocument retrieves a document by ID
	GetDocument(ctx context.Context, docID string) (*entity.RAGDocument, error)

	// ListUserDocuments lists documents uploaded by a user
	ListUserDocuments(ctx context.Context, userID string, limit, offset int) ([]*entity.RAGDocument, error)

	// DeleteDocument removes a document from tracking (doesn't delete from RAG service)
	DeleteDocument(ctx context.Context, docID, userID string) error
}

type ragBusiness struct {
	repo      mysql.RAGRepository
	ragClient rpc.RAGClient
	storage   common.StorageProvider
	tempDir   string
}

// NewRAGBusiness creates a new RAG business layer
func NewRAGBusiness(
	repo mysql.RAGRepository,
	ragClient rpc.RAGClient,
	storage common.StorageProvider,
	tempDir string,
) RAGBusiness {
	// Ensure temp directory exists
	if tempDir == "" {
		tempDir = filepath.Join(os.TempDir(), "rag_uploads")
	}
	os.MkdirAll(tempDir, 0700) // Secure permissions

	return &ragBusiness{
		repo:      repo,
		ragClient: ragClient,
		storage:   storage,
		tempDir:   tempDir,
	}
}

// UploadDocumentForRAG handles the complete upload flow
func (b *ragBusiness) UploadDocumentForRAG(
	ctx context.Context,
	req *entity.CreateRAGDocumentRequest,
	userID string,
) (*entity.RAGDocument, error) {
	// Validate collection name
	if req.Collection == "" {
		req.Collection = entity.DefaultCollection
	}

	// Check if document already exists
	existing, err := b.repo.GetDocumentByID(ctx, req.DocID)
	if err == nil && existing != nil {
		return nil, entity.ErrRAGDocumentExists
	}

	// Step 1: Download file from MinIO to temp location
	if b.storage == nil {
		return nil, entity.ErrMinioDownloadFailed.WithError("MinIO storage not configured")
	}

	fmt.Printf("[RAG Business] Downloading from MinIO: key=%s\n", req.MinioKey)

	// Extract object name from storage key (format: bucket/objectName)
	objectName := req.MinioKey
	if idx := len(req.MinioKey) - 1; idx >= 0 {
		for i := 0; i < len(req.MinioKey); i++ {
			if req.MinioKey[i] == '/' {
				objectName = req.MinioKey[i+1:]
			}
		}
	}

	data, err := b.storage.DownloadFile(ctx, objectName)
	if err != nil {
		fmt.Printf("[RAG Business] MinIO download failed: %v\n", err)
		return nil, entity.ErrMinioDownloadFailed
	}

	// Step 2: Write to secure temp file
	tempFile := filepath.Join(b.tempDir, fmt.Sprintf("%s_%d_%s", req.DocID, time.Now().UnixNano(), req.Filename))

	if err := os.WriteFile(tempFile, data, 0600); err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	// Ensure temp file is cleaned up
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			fmt.Printf("[RAG Business] Warning: failed to remove temp file %s: %v\n", tempFile, err)
		}
	}()

	fmt.Printf("[RAG Business] Created temp file: %s, size=%d bytes\n", tempFile, len(data))

	// Step 3: Upload to RAG service
	storedPath, err := b.ragClient.UploadDocument(ctx, req.DocID, tempFile, req.Collection)
	if err != nil {
		fmt.Printf("[RAG Business] RAG upload failed: %v\n", err)
		return nil, entity.ErrRAGUploadFailed
	}

	// Step 4: Persist document metadata to database
	doc := &entity.RAGDocument{
		DocID:      req.DocID,
		Collection: req.Collection,
		StoredPath: storedPath,
		Filename:   req.Filename,
		MimeType:   req.MimeType,
		SHA256:     req.SHA256,
		UploadedBy: userID,
		MinioKey:   req.MinioKey,
		UploadedAt: time.Now(),
	}

	if err := b.repo.CreateDocument(ctx, doc); err != nil {
		fmt.Printf("[RAG Business] Failed to persist document metadata: %v\n", err)
		// Note: Document is already in RAG service, so we log but don't fail completely
		// Consider adding a cleanup/reconciliation job
	}

	fmt.Printf("[RAG Business] Document uploaded successfully: doc_id=%s, stored_path=%s\n",
		doc.DocID, doc.StoredPath)

	return doc, nil
}

// UploadDocumentDirect uploads a file directly to RAG service (for multipart uploads)
func (b *ragBusiness) UploadDocumentDirect(
	ctx context.Context,
	filePath string,
	docID, collection, userID string,
) (*entity.RAGDocument, error) {
	// Validate collection name
	if collection == "" {
		collection = entity.DefaultCollection
	}

	// Check if document already exists
	existing, err := b.repo.GetDocumentByID(ctx, docID)
	if err == nil && existing != nil {
		return nil, entity.ErrRAGDocumentExists
	}

	// Upload to RAG service
	storedPath, err := b.ragClient.UploadDocument(ctx, docID, filePath, collection)
	if err != nil {
		fmt.Printf("[RAG Business] Direct upload to RAG failed: %v\n", err)
		return nil, entity.ErrRAGUploadFailed
	}

	// Get file info for metadata
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Printf("[RAG Business] Failed to stat file: %v\n", err)
	}

	filename := filepath.Base(filePath)
	mimeType := ""
	if fileInfo != nil {
		// Try to detect MIME type
		data, err := os.ReadFile(filePath)
		if err == nil && len(data) > 0 {
			mimeType = detectMimeTypeFromContent(data)
		}
	}

	// Persist document metadata to database
	doc := &entity.RAGDocument{
		DocID:      docID,
		Collection: collection,
		StoredPath: storedPath,
		Filename:   filename,
		MimeType:   mimeType,
		UploadedBy: userID,
		UploadedAt: time.Now(),
	}

	if err := b.repo.CreateDocument(ctx, doc); err != nil {
		fmt.Printf("[RAG Business] Failed to persist document metadata: %v\n", err)
	}

	fmt.Printf("[RAG Business] Direct upload success: doc_id=%s, stored_path=%s\n",
		doc.DocID, doc.StoredPath)

	return doc, nil
}

// QueryRAG queries the RAG service and maps results to internal models
func (b *ragBusiness) QueryRAG(
	ctx context.Context,
	req *entity.QueryRAGRequest,
	userID, conversationID string,
) ([]*entity.RAGQueryResult, error) {
	// Validate parameters
	if req.K <= 0 || req.K > entity.MaxQueryK {
		req.K = entity.DefaultQueryK
	}
	if req.Collection == "" {
		req.Collection = entity.DefaultCollection
	}

	fmt.Printf("[RAG Business] Querying RAG: query=%s, k=%d, collection=%s\n",
		req.Query, req.K, req.Collection)

	start := time.Now()

	// Step 1: Query RAG service
	docs, err := b.ragClient.QueryDocuments(ctx, req.Query, req.K, req.Collection)
	if err != nil {
		fmt.Printf("[RAG Business] Query failed: %v\n", err)
		return nil, entity.ErrRAGQueryFailed
	}

	latency := int(time.Since(start).Milliseconds())

	// Step 2: Map results to internal models
	results := make([]*entity.RAGQueryResult, 0, len(docs))
	docIDs := make([]string, 0, len(docs))

	for _, doc := range docs {
		result := b.mapRAGDocToResult(ctx, doc)
		results = append(results, result)
		docIDs = append(docIDs, result.DocumentID)
	}

	// Step 3: Log query for analytics
	docIDsJSON, _ := json.Marshal(docIDs)
	queryLog := &entity.RAGQueryLog{
		QueryText:      req.Query,
		UserID:         userID,
		ConversationID: conversationID,
		Collection:     req.Collection,
		K:              req.K,
		ReturnedDocs:   string(docIDsJSON),
		LatencyMs:      latency,
		CreatedAt:      time.Now(),
	}

	// Log asynchronously to avoid blocking
	go func() {
		if err := b.repo.CreateQueryLog(context.Background(), queryLog); err != nil {
			fmt.Printf("[RAG Business] Failed to log query: %v\n", err)
		}
	}()

	fmt.Printf("[RAG Business] Query success: found %d results, latency=%dms\n", len(results), latency)

	return results, nil
}

// mapRAGDocToResult maps RAG API response to internal model
func (b *ragBusiness) mapRAGDocToResult(ctx context.Context, doc entity.RAGQueryDoc) *entity.RAGQueryResult {
	result := &entity.RAGQueryResult{
		Source:   doc.Source,
		Metadata: doc.Metadata,
		Snippets: make([]entity.RAGSnippet, 0, len(doc.Matches)),
	}

	// Extract metadata fields
	if title, ok := doc.Metadata["title"].(string); ok {
		result.Title = title
	}
	if author, ok := doc.Metadata["author"].(string); ok {
		result.Author = author
	}
	if pages, ok := doc.Metadata["total_pages"].(float64); ok {
		result.TotalPages = int(pages)
	}

	// Try to find document ID from database by stored_path
	dbDoc, err := b.repo.GetDocumentByStoredPath(ctx, doc.Source)
	if err == nil && dbDoc != nil {
		result.DocumentID = dbDoc.DocID
	} else {
		// Fallback: use source path as ID
		result.DocumentID = doc.Source
	}

	// Convert matches to snippets
	for _, match := range doc.Matches {
		result.Snippets = append(result.Snippets, entity.RAGSnippet{
			Page:  match.Page,
			Score: match.Score,
			Text:  match.PageContent,
		})
	}

	return result
}

// GetDocument retrieves a document by ID
func (b *ragBusiness) GetDocument(ctx context.Context, docID string) (*entity.RAGDocument, error) {
	return b.repo.GetDocumentByID(ctx, docID)
}

// ListUserDocuments lists documents uploaded by a user
func (b *ragBusiness) ListUserDocuments(ctx context.Context, userID string, limit, offset int) ([]*entity.RAGDocument, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return b.repo.ListDocumentsByUser(ctx, userID, limit, offset)
}

// DeleteDocument removes a document from tracking
func (b *ragBusiness) DeleteDocument(ctx context.Context, docID, userID string) error {
	// Verify ownership
	doc, err := b.repo.GetDocumentByID(ctx, docID)
	if err != nil {
		return err
	}

	if doc.UploadedBy != userID {
		return entity.ErrRAGDocumentNotFound // Don't reveal existence
	}

	return b.repo.DeleteDocument(ctx, docID)
}

// detectMimeTypeFromContent detects MIME type from file content
func detectMimeTypeFromContent(data []byte) string {
	// Check magic bytes for common file types
	if len(data) < 4 {
		return "application/octet-stream"
	}

	// PDF
	if len(data) >= 4 && string(data[0:4]) == "%PDF" {
		return "application/pdf"
	}

	// JPEG
	if len(data) >= 4 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}

	// PNG
	if len(data) >= 8 && data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}

	// GIF
	if len(data) >= 6 && string(data[0:4]) == "GIF8" {
		return "image/gif"
	}

	// ZIP (also covers DOCX, XLSX, PPTX which are ZIP files)
	if len(data) >= 4 && data[0] == 0x50 && data[1] == 0x4B && data[2] == 0x03 && data[3] == 0x04 {
		return "application/zip"
	}

	// TXT - check if mostly ASCII
	if isText(data) {
		return "text/plain"
	}

	return "application/octet-stream"
}

// isText checks if data appears to be text
func isText(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// Count non-printable characters
	nonPrintable := 0
	for _, b := range data {
		if b < 32 && b != 9 && b != 10 && b != 13 {
			nonPrintable++
		}
	}

	// If more than 30% non-printable, likely binary
	return float64(nonPrintable)/float64(len(data)) < 0.3
}
