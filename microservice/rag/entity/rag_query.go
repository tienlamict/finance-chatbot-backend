package entity

// RAGMatch represents a single match from RAG query
type RAGMatch struct {
	PageContent string  `json:"page_content"`
	Score       float64 `json:"score"`
	Page        int     `json:"page"`
}

// RAGQueryDoc represents a document result from RAG query
type RAGQueryDoc struct {
	Source   string                 `json:"source"`
	Metadata map[string]interface{} `json:"metadata"`
	Matches  []RAGMatch             `json:"matches"`
}

// RAGQueryResult is the parsed and mapped result for internal use
type RAGQueryResult struct {
	DocumentID string                 `json:"document_id"` // Mapped from DB lookup
	Source     string                 `json:"source"`      // Original source path
	Title      string                 `json:"title"`       // From metadata
	Author     string                 `json:"author"`      // From metadata
	TotalPages int                    `json:"total_pages"` // From metadata
	Metadata   map[string]interface{} `json:"metadata"`    // Raw metadata
	Snippets   []RAGSnippet           `json:"snippets"`    // Formatted matches
}

// RAGSnippet represents a formatted snippet for AI context
type RAGSnippet struct {
	Page  int     `json:"page"`
	Score float64 `json:"score"`
	Text  string  `json:"text"`
}

// ToAIContextSnippet converts RAGQueryResult to AI context format
func (r *RAGQueryResult) ToAIContextSnippets() []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(r.Snippets))
	for _, snippet := range r.Snippets {
		result = append(result, map[string]interface{}{
			"source":      r.Source,
			"title":       r.Title,
			"page":        snippet.Page,
			"score":       snippet.Score,
			"text":        snippet.Text,
			"document_id": r.DocumentID,
		})
	}
	return result
}
