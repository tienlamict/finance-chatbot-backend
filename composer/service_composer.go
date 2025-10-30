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

	userBusiness "finance-chatbot/microservice/user/business"
	"finance-chatbot/microservice/user/entity"
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

	"gorm.io/gorm"
)

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

	// Enhanced conversation history handlers
	GetConversationHistoryHandler() func(*gin.Context)
	GetConversationSummaryHandler() func(*gin.Context)
	GetUserConversationsHandler() func(*gin.Context)
}

type RBACService interface {
	// Role management
	CreateRoleHdl() func(*gin.Context)
	ListRolesHdl() func(*gin.Context)
	GetRoleWithPermissionsHdl() func(*gin.Context)
	UpdateRoleHdl() func(*gin.Context)
	DeleteRoleHdl() func(*gin.Context)

	// Permission management
	CreatePermissionHdl() func(*gin.Context)
	ListPermissionsHdl() func(*gin.Context)

	// Role-Permission assignment
	AssignPermissionsToRoleHdl() func(*gin.Context)
	RemovePermissionFromRoleHdl() func(*gin.Context)

	// User-Role assignment
	AssignRolesToUserHdl() func(*gin.Context)
	GetUserRolesHdl() func(*gin.Context)
	RemoveRoleFromUserHdl() func(*gin.Context)
}

type AdminUserService interface {
	CreateUserHdl() func(*gin.Context)
	UpdateUserHdl() func(*gin.Context)
	DeleteUserHdl() func(*gin.Context)
	ListUsersHdl() func(*gin.Context)
	GetUserByIDHdl() func(*gin.Context)
}

func ComposeUserAPIService(serviceCtx sctx.ServiceContext) UserService {
	db := serviceCtx.MustGet(common.KeyCompMySQL).(common.GormComponent)

	userRepo := userSQLRepository.NewMySQLRepository(db.GetDB())
	biz := userBusiness.NewBusiness(userRepo)
	userService := userApi.NewAPI(biz)

	return userService
}

func ComposeAuthAPIService(serviceCtx sctx.ServiceContext) AuthService {
	db := serviceCtx.MustGet(common.KeyCompMySQL).(common.GormComponent)
	jwtComp := serviceCtx.MustGet(common.KeyCompJWT).(common.JWTProvider)

	authRepo := authSQLRepository.NewMySQLRepository(db.GetDB())
	hasher := new(common.Hasher)

	// Create RBAC business for role assignment
	rbacStore := userSQLRepository.NewRBACStore(db.GetDB())
	rbacBiz := userBusiness.NewRBACBusiness(rbacStore)

	userClient := authUserRPC.NewClient(composeUserRPCClient(serviceCtx))
	biz := authBusiness.NewBusiness(authRepo, userClient, rbacBiz, jwtComp, hasher)
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

// userStoreAdapter is a simpler adapter that directly uses the repo's methods
type userStoreAdapter struct {
	db *gorm.DB
}

func (w *userStoreAdapter) GetUserByID(ctx context.Context, userID int) (*entity.User, error) {
	repo := userSQLRepository.NewMySQLRepository(w.db)
	return repo.GetUserById(ctx, userID)
}

func ComposeUserStore(serviceCtx sctx.ServiceContext) *userStoreAdapter {
	db := serviceCtx.MustGet(common.KeyCompMySQL).(common.GormComponent)
	return &userStoreAdapter{db: db.GetDB()}
}

func ComposeRBACAPIService(serviceCtx sctx.ServiceContext) RBACService {
	db := serviceCtx.MustGet(common.KeyCompMySQL).(common.GormComponent)

	rbacStore := userSQLRepository.NewRBACStore(db.GetDB())
	biz := userBusiness.NewRBACBusiness(rbacStore)
	serviceAPI := userApi.NewRBACAPI(biz)

	return serviceAPI
}

func ComposeAdminUserAPIService(serviceCtx sctx.ServiceContext) AdminUserService {
	db := serviceCtx.MustGet(common.KeyCompMySQL).(common.GormComponent)
	hasher := new(common.Hasher)

	userRepo := userSQLRepository.NewMySQLRepository(db.GetDB())
	authRepo := userSQLRepository.NewAdminAuthRepository(db.GetDB())
	passwordBiz := userBusiness.NewAdminUserPasswordBusiness(userRepo, authRepo, hasher)
	biz := userBusiness.NewAdminUserBusiness(userRepo, passwordBiz)
	serviceAPI := userApi.NewAdminUserAPI(biz)

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

	// Create RBAC business for role assignment
	rbacStore := userSQLRepository.NewRBACStore(db.GetDB())
	rbacBiz := userBusiness.NewRBACBusiness(rbacStore)

	// In Auth GRPC service, user repository is unnecessary
	biz := authBusiness.NewBusiness(authRepo, nil, rbacBiz, jwtComp, hasher)
	authService := authRPC.NewService(biz)

	return authService
}

// Chọn AI client theo ENV: AI_PROTOCOL=rest|grpc|enhanced (mặc định enhanced)
func chooseAIClient(serviceCtx sctx.ServiceContext) chatrpc.AIClient {
	switch os.Getenv("AI_PROTOCOL") {
	case "mock", "MOCK":
		return NewAIMockClient() // mock for testing
	case "rest":
		return composeAIRESTClient() // <--- Basic REST adapter (legacy)
	// case "grpc":
	// 	aiGrpc := composeAIRPCClient(serviceCtx)      // <--- gRPC cũ (đã có)
	// 	return NewAIClientAdapter(aiGrpc)
	default:
		return NewEnhancedAIClient() // <--- Enhanced REST adapter with full context support (default)
	}
}

// Dùng cho HTTP
func ComposeChatbotAPIService(serviceCtx sctx.ServiceContext) ChatbotService {
	dbComp := serviceCtx.MustGet(common.KeyCompMySQL).(common.GormComponent)
	gormDB := dbComp.GetDB()

	aiClient := chooseAIClient(serviceCtx) // REST/gRPC

	// Get storage component (MinIO/S3) if available
	var storage common.StorageProvider
	if storageComp, ok := serviceCtx.Get(common.KeyCompStorage); ok {
		storage = storageComp.(common.StorageProvider)
	}

	sqlRepo := chatmysql.NewMySQLRepo(gormDB)
	biz := chatBusiness.NewChatBusiness(sqlRepo, aiClient, storage)
	api := chatAPI.NewAPI(biz)
	return api
}
