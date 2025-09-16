package composer

import (
	"finance-chatbot/addon/common"
	authBusiness "finance-chatbot/microservice/auth/business"
	authSQLRepository "finance-chatbot/microservice/auth/repository/mysql"
	authUserRPC "finance-chatbot/microservice/auth/repository/rpc"
	authAPI "finance-chatbot/microservice/auth/transport/api"
	authRPC "finance-chatbot/microservice/auth/transport/rpc"

	taskBusiness "finance-chatbot/microservice/task/business"
	taskSQLRepository "finance-chatbot/microservice/task/repository/mysql"
	taskUserRPC "finance-chatbot/microservice/task/repository/rpc"
	taskAPI "finance-chatbot/microservice/task/transport/api"

	userBusiness "finance-chatbot/microservice/user/business"
	userSQLRepository "finance-chatbot/microservice/user/repository/mysql"
	userApi "finance-chatbot/microservice/user/transport/api"
	userRPC "finance-chatbot/microservice/user/transport/rpc"

	chatBiz "finance-chatbot/microservice/chatbot/business"
	"finance-chatbot/microservice/chatbot/entity"
	chatRepo "finance-chatbot/microservice/chatbot/repository/mysql"
	chatAPI "finance-chatbot/microservice/chatbot/transport/api"
	httpclient "finance-chatbot/microservice/chatbot/transport/http"

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
	SendMessage() func(*gin.Context)
	GetHistory() func(*gin.Context)
}

func ComposeUserAPIService(serviceCtx sctx.ServiceContext) UserService {
	db := serviceCtx.MustGet(common.KeyCompMySQL).(common.GormComponent)

	userRepo := userSQLRepository.NewMySQLRepository(db.GetDB())
	biz := userBusiness.NewBusiness(userRepo)
	userService := userApi.NewAPI(biz)

	return userService
}

func ComposeTaskAPIService(serviceCtx sctx.ServiceContext) TaskService {
	db := serviceCtx.MustGet(common.KeyCompMySQL).(common.GormComponent)

	userClient := taskUserRPC.NewClient(composeUserRPCClient(serviceCtx))
	taskRepo := taskSQLRepository.NewMySQLRepository(db.GetDB())
	biz := taskBusiness.NewBusiness(taskRepo, userClient)
	serviceAPI := taskAPI.NewAPI(serviceCtx, biz)

	return serviceAPI
}

func ComposeChatAPIService(sc sctx.ServiceContext) *chatAPI.ChatAPI {
	db := sc.MustGet(common.KeyCompMySQL).(common.GormComponent)

	// Lấy config từ KeyCompConf (repo của bạn dùng interface common.Config)
	cfg := sc.MustGet(common.KeyCompConf).(common.Config)

	// giả sử bạn bổ sung method này, hoặc đọc ENV: AI_HTTP_BASE_URL, AI_HTTP_API_KEY
	baseURL := cfg.GetAIAddress() // ví dụ: http://ai-service:8080
	//apiKey := cfg.GetAIAuthorization() // nếu có

	var ai entity.AIClient
	if v, ok := sc.Get(common.KeyCompAIClient); ok {
		ai = v.(entity.AIClient)
	} else {
		ai = httpclient.NewAIHTTPClient(baseURL, apiKey)
		sc.Add(common.KeyCompAIClient, ai)
	}

	repo := chatRepo.NewMySQLRepository(db.GetDB())
	uc := chatBiz.NewChatBusiness(repo, ai)

	return chatAPI.NewAPI(uc)
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
