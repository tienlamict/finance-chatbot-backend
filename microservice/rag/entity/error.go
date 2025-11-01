package entity

import (
	"finance-chatbot/addon/core"
	"net/http"
)

var (
	ErrRAGDocumentNotFound = &core.DefaultError{
		IDField:     "ERR_RAG_DOC_NOT_FOUND",
		CodeField:   http.StatusNotFound,
		StatusField: http.StatusText(http.StatusNotFound),
		ErrorField:  "Document not found in RAG system",
	}

	ErrRAGUploadFailed = &core.DefaultError{
		IDField:     "ERR_RAG_UPLOAD_FAILED",
		CodeField:   http.StatusBadGateway,
		StatusField: http.StatusText(http.StatusBadGateway),
		ErrorField:  "Failed to upload document to RAG service",
	}

	ErrRAGQueryFailed = &core.DefaultError{
		IDField:     "ERR_RAG_QUERY_FAILED",
		CodeField:   http.StatusBadGateway,
		StatusField: http.StatusText(http.StatusBadGateway),
		ErrorField:  "Failed to query RAG service",
	}

	ErrRAGInvalidResponse = &core.DefaultError{
		IDField:     "ERR_RAG_INVALID_RESPONSE",
		CodeField:   http.StatusBadGateway,
		StatusField: http.StatusText(http.StatusBadGateway),
		ErrorField:  "Invalid response from RAG service",
	}

	ErrRAGTimeout = &core.DefaultError{
		IDField:     "ERR_RAG_TIMEOUT",
		CodeField:   http.StatusGatewayTimeout,
		StatusField: http.StatusText(http.StatusGatewayTimeout),
		ErrorField:  "RAG service timeout",
	}

	ErrRAGDocumentExists = &core.DefaultError{
		IDField:     "ERR_RAG_DOC_EXISTS",
		CodeField:   http.StatusConflict,
		StatusField: http.StatusText(http.StatusConflict),
		ErrorField:  "Document already exists in RAG system",
	}

	ErrMinioDownloadFailed = &core.DefaultError{
		IDField:     "ERR_MINIO_DOWNLOAD_FAILED",
		CodeField:   http.StatusBadGateway,
		StatusField: http.StatusText(http.StatusBadGateway),
		ErrorField:  "Failed to download file from MinIO",
	}

	ErrInvalidCollection = &core.DefaultError{
		IDField:     "ERR_INVALID_COLLECTION",
		CodeField:   http.StatusBadRequest,
		StatusField: http.StatusText(http.StatusBadRequest),
		ErrorField:  "Invalid RAG collection name",
	}
)
