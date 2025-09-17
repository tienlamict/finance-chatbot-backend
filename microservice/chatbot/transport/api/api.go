package api

import (
	"context"
	sctx "finance-chatbot/addon/sctx"
	"finance-chatbot/microservice/chatbot/entity"
)

// Interface hợp nhất hạ tầng (DB) và nghiệp vụ (Business)
// => nếu struct nào implement interface này, nó sẽ vừa có capability của ServiceContext (DI) vừa có logic nghiệp vụ
type ServiceContext interface {
	sctx.ServiceContext
	Business
}

// Định nghĩa các hành vi nghiệp vụ mà transport cần.
type Business interface {
	CreateNewChat(ctx context.Context, data *entity.ChatDataCreation) error
	//GetMessageByUserId(ctx context.Context, id int) (*entity.ChatMessage, error) // => chưa implement sẽ báo lỗi trong composer
	//ListMessages(ctx context.Context, userID string, limit int) ([]entity.ChatMessage, error)
}

// Struct implement tầng API
type api struct {
	serviceCtx sctx.ServiceContext
	business   Business
}

// Constructor function để tạo api
// Được gọi trong service_composer.go
func NewAPI(serviceCtx sctx.ServiceContext, business Business) *api {
	return &api{
		serviceCtx: serviceCtx,
		business:   business,
	}
}
