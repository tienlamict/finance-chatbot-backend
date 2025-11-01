package business

import (
	"context"
	"testing"
	"time"

	"finance-chatbot/microservice/rag/entity"
)

// Mock implementations for testing
type mockRAGRepo struct{}

func (m *mockRAGRepo) CreateDocument(ctx context.Context, doc *entity.RAGDocument) error {
	return nil
}

func (m *mockRAGRepo) GetDocumentByID(ctx context.Context, docID string) (*entity.RAGDocument, error) {
	return &entity.RAGDocument{
		DocID:      docID,
		Collection: "test_collection",
		StoredPath: "./temp_uploads/test.pdf",
	}, nil
}

func (m *mockRAGRepo) GetDocumentByStoredPath(ctx context.Context, storedPath string) (*entity.RAGDocument, error) {
	return &entity.RAGDocument{
		DocID:      "test-doc-1",
		StoredPath: storedPath,
		Collection: "test_collection",
	}, nil
}

func (m *mockRAGRepo) ListDocumentsByCollection(ctx context.Context, collection string, limit, offset int) ([]*entity.RAGDocument, error) {
	return []*entity.RAGDocument{}, nil
}

func (m *mockRAGRepo) ListDocumentsByUser(ctx context.Context, userID string, limit, offset int) ([]*entity.RAGDocument, error) {
	return []*entity.RAGDocument{}, nil
}

func (m *mockRAGRepo) DeleteDocument(ctx context.Context, docID string) error {
	return nil
}

func (m *mockRAGRepo) CreateQueryLog(ctx context.Context, log *entity.RAGQueryLog) error {
	return nil
}

func (m *mockRAGRepo) GetRecentQueries(ctx context.Context, userID string, limit int) ([]*entity.RAGQueryLog, error) {
	return []*entity.RAGQueryLog{}, nil
}

type mockRAGClient struct{}

func (m *mockRAGClient) UploadDocument(ctx context.Context, docID, filePath, collection string) (string, error) {
	return "./temp_uploads/test.pdf", nil
}

func (m *mockRAGClient) QueryDocuments(ctx context.Context, query string, k int, collection string) ([]entity.RAGQueryDoc, error) {
	return []entity.RAGQueryDoc{
		{
			Source: "./temp_uploads/test.pdf",
			Metadata: map[string]interface{}{
				"title":       "Test Document",
				"total_pages": float64(10),
			},
			Matches: []entity.RAGMatch{
				{
					PageContent: "This is a test snippet",
					Score:       0.95,
					Page:        1,
				},
			},
		},
	}, nil
}

type mockStorage struct{}

func (m *mockStorage) UploadFile(ctx context.Context, objectName string, data []byte, contentType string) (string, error) {
	return "bucket/" + objectName, nil
}

func (m *mockStorage) DownloadFile(ctx context.Context, objectName string) ([]byte, error) {
	return []byte("mock file content"), nil
}

func (m *mockStorage) DeleteFile(ctx context.Context, objectName string) error {
	return nil
}

func (m *mockStorage) GetPresignedURL(ctx context.Context, objectName string, expires time.Duration) (string, error) {
	return "https://example.com/presigned-url", nil
}

// TestQueryRAG tests the query functionality
func TestQueryRAG(t *testing.T) {
	repo := &mockRAGRepo{}
	client := &mockRAGClient{}
	storage := &mockStorage{}

	biz := NewRAGBusiness(repo, client, storage, "")

	req := &entity.QueryRAGRequest{
		Query:      "test query",
		K:          5,
		Collection: "test_collection",
	}

	results, err := biz.QueryRAG(context.Background(), req, "test-user", "test-conv")
	if err != nil {
		t.Fatalf("QueryRAG failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected at least one result")
	}

	if results[0].DocumentID == "" {
		t.Error("Expected document ID to be set")
	}

	if len(results[0].Snippets) == 0 {
		t.Error("Expected at least one snippet")
	}
}

// TestMapRAGDocToResult tests the mapping function
func TestMapRAGDocToResult(t *testing.T) {
	repo := &mockRAGRepo{}
	client := &mockRAGClient{}
	storage := &mockStorage{}

	biz := NewRAGBusiness(repo, client, storage, "").(*ragBusiness)

	doc := entity.RAGQueryDoc{
		Source: "./temp_uploads/test.pdf",
		Metadata: map[string]interface{}{
			"title":       "Test Document",
			"author":      "Test Author",
			"total_pages": float64(10),
		},
		Matches: []entity.RAGMatch{
			{
				PageContent: "Test content",
				Score:       0.95,
				Page:        1,
			},
		},
	}

	result := biz.mapRAGDocToResult(context.Background(), doc)

	if result.Title != "Test Document" {
		t.Errorf("Expected title 'Test Document', got '%s'", result.Title)
	}

	if result.Author != "Test Author" {
		t.Errorf("Expected author 'Test Author', got '%s'", result.Author)
	}

	if result.TotalPages != 10 {
		t.Errorf("Expected 10 pages, got %d", result.TotalPages)
	}

	if len(result.Snippets) != 1 {
		t.Fatalf("Expected 1 snippet, got %d", len(result.Snippets))
	}

	if result.Snippets[0].Text != "Test content" {
		t.Errorf("Expected snippet text 'Test content', got '%s'", result.Snippets[0].Text)
	}
}
