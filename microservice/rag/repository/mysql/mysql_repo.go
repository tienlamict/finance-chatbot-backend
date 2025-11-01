package mysql

import (
	"context"
	"finance-chatbot/microservice/rag/entity"

	"gorm.io/gorm"
)

// RAGRepository defines the interface for RAG document persistence
type RAGRepository interface {
	// Document operations
	CreateDocument(ctx context.Context, doc *entity.RAGDocument) error
	GetDocumentByID(ctx context.Context, docID string) (*entity.RAGDocument, error)
	GetDocumentByStoredPath(ctx context.Context, storedPath string) (*entity.RAGDocument, error)
	ListDocumentsByCollection(ctx context.Context, collection string, limit, offset int) ([]*entity.RAGDocument, error)
	ListDocumentsByUser(ctx context.Context, userID string, limit, offset int) ([]*entity.RAGDocument, error)
	DeleteDocument(ctx context.Context, docID string) error

	// Query log operations
	CreateQueryLog(ctx context.Context, log *entity.RAGQueryLog) error
	GetRecentQueries(ctx context.Context, userID string, limit int) ([]*entity.RAGQueryLog, error)
}

type ragRepo struct {
	db *gorm.DB
}

// NewRAGRepository creates a new RAG repository
func NewRAGRepository(db *gorm.DB) RAGRepository {
	return &ragRepo{db: db}
}

// CreateDocument creates a new RAG document record
func (r *ragRepo) CreateDocument(ctx context.Context, doc *entity.RAGDocument) error {
	if err := r.db.WithContext(ctx).Create(doc).Error; err != nil {
		return err
	}
	return nil
}

// GetDocumentByID retrieves a document by its doc_id
func (r *ragRepo) GetDocumentByID(ctx context.Context, docID string) (*entity.RAGDocument, error) {
	var doc entity.RAGDocument
	if err := r.db.WithContext(ctx).Where("doc_id = ? AND status = 1", docID).First(&doc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entity.ErrRAGDocumentNotFound
		}
		return nil, err
	}
	return &doc, nil
}

// GetDocumentByStoredPath retrieves a document by its stored_path (from RAG API)
func (r *ragRepo) GetDocumentByStoredPath(ctx context.Context, storedPath string) (*entity.RAGDocument, error) {
	var doc entity.RAGDocument
	if err := r.db.WithContext(ctx).Where("stored_path = ? AND status = 1", storedPath).First(&doc).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entity.ErrRAGDocumentNotFound
		}
		return nil, err
	}
	return &doc, nil
}

// ListDocumentsByCollection retrieves documents by collection
func (r *ragRepo) ListDocumentsByCollection(ctx context.Context, collection string, limit, offset int) ([]*entity.RAGDocument, error) {
	var docs []*entity.RAGDocument
	query := r.db.WithContext(ctx).
		Where("collection = ? AND status = 1", collection).
		Order("uploaded_at DESC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

// ListDocumentsByUser retrieves documents uploaded by a specific user
func (r *ragRepo) ListDocumentsByUser(ctx context.Context, userID string, limit, offset int) ([]*entity.RAGDocument, error) {
	var docs []*entity.RAGDocument
	query := r.db.WithContext(ctx).
		Where("uploaded_by = ? AND status = 1", userID).
		Order("uploaded_at DESC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

// DeleteDocument soft deletes a document
func (r *ragRepo) DeleteDocument(ctx context.Context, docID string) error {
	result := r.db.WithContext(ctx).
		Model(&entity.RAGDocument{}).
		Where("doc_id = ?", docID).
		Update("status", 0)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entity.ErrRAGDocumentNotFound
	}
	return nil
}

// CreateQueryLog creates a new query log entry
func (r *ragRepo) CreateQueryLog(ctx context.Context, log *entity.RAGQueryLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetRecentQueries retrieves recent queries for a user
func (r *ragRepo) GetRecentQueries(ctx context.Context, userID string, limit int) ([]*entity.RAGQueryLog, error) {
	var logs []*entity.RAGQueryLog
	query := r.db.WithContext(ctx).
		Where("user_id = ? AND status = 1", userID).
		Order("created_at DESC").
		Limit(limit)

	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
