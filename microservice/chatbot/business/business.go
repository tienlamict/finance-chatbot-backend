// finance-chatbot-backend/microservice/chatbot/business/business.go
package business

import (
	"context"

	"finance-chatbot/microservice/chatbot/entity"
	mysqlrepo "finance-chatbot/microservice/chatbot/repository/mysql"
	rpcrepo "finance-chatbot/microservice/chatbot/repository/rpc"
)

type ChatUsecase interface {
	SendMessage(ctx context.Context, req entity.SendMessageRequest, userID string) (*entity.SendMessageResponse, error)
	ListMessages(ctx context.Context, req entity.ListMessagesRequest) (*entity.ListMessagesResponse, error)
}

type chatUsecase struct {
	sql interface {
		mysqlrepo.ConversationStore
		mysqlrepo.MessageStore
		mysqlrepo.AttachmentStore
	}
	ai      rpcrepo.AIClient
	storage StorageProvider
}

type StorageProvider interface {
	UploadFile(ctx context.Context, objectName string, data []byte, contentType string) (string, error)
	DownloadFile(ctx context.Context, objectName string) ([]byte, error)
	DeleteFile(ctx context.Context, objectName string) error
}

func NewChatBusiness(sqlRepo *mysqlrepo.MySQLRepo, aiClient rpcrepo.AIClient, storage StorageProvider) ChatUsecase {
	return &chatUsecase{
		sql:     sqlRepo,
		ai:      aiClient,
		storage: storage,
	}
}
