package composer

import (
	"context"
	"finance-chatbot/addon/common"
	authBusiness "finance-chatbot/microservice/auth/business"
	authSQLRepository "finance-chatbot/microservice/auth/repository/mysql"
	authUserRPC "finance-chatbot/microservice/auth/repository/rpc"
	authAPI "finance-chatbot/microservice/auth/transport/api"
	authRPC "finance-chatbot/microservice/auth/transport/rpc"
	"os"

	taskBusiness "finance-chatbot/microservice/task/business"
	taskSQLRepository "finance-chatbot/microservice/task/repository/mysql"
	taskUserRPC "finance-chatbot/microservice/task/repository/rpc"
	taskAPI "finance-chatbot/microservice/task/transport/api"

	userBusiness "finance-chatbot/microservice/user/business"
	userSQLRepository "finance-chatbot/microservice/user/repository/mysql"
	userApi "finance-chatbot/microservice/user/transport/api"
	userRPC "finance-chatbot/microservice/user/transport/rpc"

	chatBusiness "finance-chatbot/microservice/chatbot/business"
	chatmysql "finance-chatbot/microservice/chatbot/repository/mysql"
	chatrpc "finance-chatbot/microservice/chatbot/repository/rpc"
	chatAPI "finance-chatbot/microservice/chatbot/transport/api"

	"github.com/gin-gonic/gin"

	"finance-chatbot/proto/pb"

	sctx "finance-chatbot/addon/sctx"
)

type TaskService interface {
	CreateTaskHdl() func(*gin.Context)
	GetTaskHdl() func(*gin.Context)
	ListTaskHdl() func(*gin.Context)
	UpdateTaskHdl() func(*gin.Context)
	DeleteTaskHdl() func(*gin.Context)
}

type UserService interface {
	GetUserProfileHdl() func(*gin.Context)
}

type AuthService interface {
	LoginHdl() func(*gin.Context)
	RegisterHdl() func(*gin.Context)
}

type ChatbotService interface {
	SendMessageHandler() func(*gin.Context)
	ListMessagesHandler() func(*gin.Context)
}

func ComposeUserAPIService(serviceCtx sctx.ServiceContext) UserService {
	db := serviceCtx.MustGet(common.KeyCompMySQL).(common.GormComponent)

	userRepo := userSQLRepository.NewMySQLRepository(db.GetDB())
	biz := userBusiness.NewBusiness(userRepo)
	userService := userApi.NewAPI(biz)

	return userService
}

// Đây là một hàm factory, nhận vào ServiceContext (container chứa các dependency chung như DB, logger, config…)
// Trả về một TaskService (tức API service của module task)
func ComposeTaskAPIService(serviceCtx sctx.ServiceContext) TaskService {
	// Lấy component kết nối MySQL từ serviceCtx
	// Ép kiểu về GormComponent để sử dụng GORM
	db := serviceCtx.MustGet(common.KeyCompMySQL).(common.GormComponent)

	// Tạo client RPC để gọi sang User microservice
	// Hàm composeUserRPCClient được định nghĩa trong cùng package composer
	// composeUserRPCClient sẽ dựng kết nối gRPC tới service user, sau đó taskUserRPC.NewClient(...) bọc lại thành client chuyên biệt.
	userClient := taskUserRPC.NewClient(composeUserRPCClient(serviceCtx))

	// Tạo repository layer cho task, implement bằng MySQL.
	// db.GetDB() trả về con trỏ *gorm.DB và được wrap trong taskSQLRepository.
	taskRepo := taskSQLRepository.NewMySQLRepository(db.GetDB())

	// Tạo business/usecase layer cho task.
	// biz sẽ chứa logic chính: tạo task, xoá, update…
	// Nó phụ thuộc vào taskRepo (đọc/ghi DB) và userClient (gọi sang user service để xác thực/kiểm tra user).
	biz := taskBusiness.NewBusiness(taskRepo, userClient)

	// Tạo transport layer cho task (HTTP/gRPC handler).
	// taskAPI.NewAPI nhận biz để xử lý request, đồng thời dùng serviceCtx để đăng ký middleware/logging…
	serviceAPI := taskAPI.NewAPI(serviceCtx, biz)

	return serviceAPI
}

func ComposeAuthAPIService(serviceCtx sctx.ServiceContext) AuthService {
	db := serviceCtx.MustGet(common.KeyCompMySQL).(common.GormComponent)
	jwtComp := serviceCtx.MustGet(common.KeyCompJWT).(common.JWTProvider)

	authRepo := authSQLRepository.NewMySQLRepository(db.GetDB())
	hasher := new(common.Hasher)

	userClient := authUserRPC.NewClient(composeUserRPCClient(serviceCtx))
	biz := authBusiness.NewBusiness(authRepo, userClient, jwtComp, hasher)
	serviceAPI := authAPI.NewAPI(serviceCtx, biz)

	return serviceAPI
}

// ComposeRBACClient exposes a lightweight adapter for permission checks in middleware layer.
type rbacClient struct {
	repo userSQLRepository.RBACRepository
}

func (r *rbacClient) HasAnyPermission(userID int, permCodes ...string) (bool, error) {
	for _, code := range permCodes {
		ok, err := r.repo.HasPermission(context.Background(), userID, code)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func ComposeRBACClient(serviceCtx sctx.ServiceContext) *rbacClient {
	db := serviceCtx.MustGet(common.KeyCompMySQL).(common.GormComponent)
	repo := userSQLRepository.NewRBACRepository(db.GetDB())
	return &rbacClient{repo: repo}
}

func ComposeUserGRPCService(serviceCtx sctx.ServiceContext) pb.UserServiceServer {
	db := serviceCtx.MustGet(common.KeyCompMySQL).(common.GormComponent)

	userRepo := userSQLRepository.NewMySQLRepository(db.GetDB())
	userBiz := userBusiness.NewBusiness(userRepo)
	userService := userRPC.NewService(userBiz)

	return userService
}

func ComposeAuthGRPCService(serviceCtx sctx.ServiceContext) pb.AuthServiceServer {
	db := serviceCtx.MustGet(common.KeyCompMySQL).(common.GormComponent)
	jwtComp := serviceCtx.MustGet(common.KeyCompJWT).(common.JWTProvider)

	authRepo := authSQLRepository.NewMySQLRepository(db.GetDB())
	hasher := new(common.Hasher)

	// In Auth GRPC service, user repository is unnecessary
	biz := authBusiness.NewBusiness(authRepo, nil, jwtComp, hasher)
	authService := authRPC.NewService(biz)

	return authService
}

// Chọn AI client theo ENV: AI_PROTOCOL=rest|grpc (mặc định grpc)
func chooseAIClient(serviceCtx sctx.ServiceContext) chatrpc.AIClient {
	switch os.Getenv("AI_PROTOCOL") {
	case "rest", "REST":
		return composeAIRESTClient() // <--- REST adapter mới
	// case "grpc":
	// 	aiGrpc := composeAIRPCClient(serviceCtx)      // <--- gRPC cũ (đã có)
	// 	return NewAIClientAdapter(aiGrpc)
	default:
		return NewAIMockClient() // mock
	}
}

// Dùng cho HTTP
func ComposeChatbotAPIService(serviceCtx sctx.ServiceContext) ChatbotService {
	dbComp := serviceCtx.MustGet(common.KeyCompMySQL).(common.GormComponent)
	gormDB := dbComp.GetDB()

	aiClient := chooseAIClient(serviceCtx) // REST/gRPC

	sqlRepo := chatmysql.NewMySQLRepo(gormDB)
	biz := chatBusiness.NewChatBusiness(sqlRepo, aiClient)
	api := chatAPI.NewAPI(biz)
	return api
}
