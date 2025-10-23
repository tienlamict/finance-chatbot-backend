package cmd

import (
	"finance-chatbot/addon/common"
	"finance-chatbot/addon/component/ginc"
	smdlw "finance-chatbot/addon/component/ginc/middleware"
	"finance-chatbot/addon/component/gormc"
	"finance-chatbot/addon/component/jwtc"
	sctx "finance-chatbot/addon/sctx"
	"finance-chatbot/composer"
	"finance-chatbot/middleware"
	"finance-chatbot/proto/pb"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func newServiceCtx() sctx.ServiceContext {
	opts := []sctx.Option{
		sctx.WithName("Demo Microservices"),
		sctx.WithComponent(ginc.NewGin(common.KeyCompGIN)),
		sctx.WithComponent(gormc.NewGormDB(common.KeyCompMySQL, "")),
		sctx.WithComponent(jwtc.NewJWT(common.KeyCompJWT)),
		sctx.WithComponent(NewConfig()),
	}

	// Add MinIO component if configured
	if minioComps := getMinioComponentIfConfigured(); minioComps != nil {
		for _, comp := range minioComps {
			opts = append(opts, sctx.WithComponent(comp))
		}
	}

	return sctx.NewServiceContext(opts...)
}

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "Start service",
	Run: func(cmd *cobra.Command, args []string) {
		serviceCtx := newServiceCtx()

		logger := sctx.GlobalLogger().GetLogger("service")

		// Make some delay for DB ready (migration)
		// remove it if you already had your own DB
		time.Sleep(time.Second * 5)

		if err := serviceCtx.Load(); err != nil {
			logger.Fatal(err)
		}

		ginComp := serviceCtx.MustGet(common.KeyCompGIN).(common.GINComponent)

		router := ginComp.GetRouter()
		router.Use(gin.Recovery(), gin.Logger(), smdlw.Recovery(serviceCtx))

		// Add CORS middleware
		router.Use(middleware.CORSMiddleware())

		router.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"data": "pong"})
		})

		go StartGRPCServices(serviceCtx)

		v1 := router.Group("/v1")

		SetupRoutes(v1, serviceCtx)

		if err := router.Run(fmt.Sprintf(":%d", ginComp.GetPort())); err != nil {
			logger.Fatal(err)
		}
	},
}

func SetupRoutes(router *gin.RouterGroup, serviceCtx sctx.ServiceContext) {

	userAPIService := composer.ComposeUserAPIService(serviceCtx)
	authAPIService := composer.ComposeAuthAPIService(serviceCtx)
	chatbotAPIService := composer.ComposeChatbotAPIService(serviceCtx)
	rbacAPIService := composer.ComposeRBACAPIService(serviceCtx)
	adminUserAPIService := composer.ComposeAdminUserAPIService(serviceCtx)

	requireAuthMdw := middleware.RequireAuth(composer.ComposeAuthRPCClient(serviceCtx))
	rbacClient := composer.ComposeRBACClient(serviceCtx)
	userStore := composer.ComposeUserStore(serviceCtx)
	requireSuperAdminMdw := middleware.RequireSuperAdmin(userStore)

	router.POST("/authenticate", authAPIService.LoginHdl())
	router.POST("/register", authAPIService.RegisterHdl())
	router.GET("/profile", requireAuthMdw, userAPIService.GetUserProfileHdl())

	// Chat permissions (example codes: chat.send, chat.read)
	chat := router.Group("/chatbot", requireAuthMdw)
	{
		chat.POST("/promt", middleware.RequirePermissions(rbacClient, "chat.send"), chatbotAPIService.SendMessageHandler())
		chat.GET("/messages", middleware.RequirePermissions(rbacClient, "chat.read"), chatbotAPIService.ListMessagesHandler())

		// Enhanced conversation history endpoints
		chat.GET("/conversations/:conversationId/messages", middleware.RequirePermissions(rbacClient, "chat.read"), chatbotAPIService.GetConversationHistoryHandler())
		chat.GET("/conversations/:conversationId", middleware.RequirePermissions(rbacClient, "chat.read"), chatbotAPIService.GetConversationSummaryHandler())
		chat.GET("/users/:userId/conversations", middleware.RequirePermissions(rbacClient, "chat.read"), chatbotAPIService.GetUserConversationsHandler())
	}

	// Admin User Management (admin/superadmin only)
	admin := router.Group("/admin", requireAuthMdw, requireSuperAdminMdw)
	{
		// User CRUD operations
		admin.POST("/users", adminUserAPIService.CreateUserHdl())
		admin.GET("/users", adminUserAPIService.ListUsersHdl())
		admin.GET("/users/:id", adminUserAPIService.GetUserByIDHdl())
		admin.PUT("/users/:id", adminUserAPIService.UpdateUserHdl())
		admin.PATCH("/users/:id", adminUserAPIService.UpdateUserHdl())
		admin.DELETE("/users/:id", adminUserAPIService.DeleteUserHdl())
	}

	// RBAC Management (superadmin only)
	// IMPORTANT: Routes are structured to avoid Gin wildcard conflicts
	rbac := router.Group("/rbac", requireAuthMdw, requireSuperAdminMdw)
	{
		// Role management
		rbac.POST("/roles", rbacAPIService.CreateRoleHdl())
		rbac.GET("/roles", rbacAPIService.ListRolesHdl())
		rbac.GET("/roles/:id", rbacAPIService.GetRoleWithPermissionsHdl())
		rbac.PATCH("/roles/:id", rbacAPIService.UpdateRoleHdl())
		rbac.DELETE("/roles/:id", rbacAPIService.DeleteRoleHdl())

		// Permission management
		rbac.POST("/permissions", rbacAPIService.CreatePermissionHdl())
		rbac.GET("/permissions", rbacAPIService.ListPermissionsHdl())

		// Role-Permission assignment (using 'role-permissions' to avoid conflicts)
		rbac.POST("/role-permissions/:id", rbacAPIService.AssignPermissionsToRoleHdl())
		rbac.DELETE("/role-permissions/:id/:permId", rbacAPIService.RemovePermissionFromRoleHdl())

		// User-Role assignment (using 'user-roles' to avoid conflicts)
		rbac.POST("/user-roles/:userId", rbacAPIService.AssignRolesToUserHdl())
		rbac.GET("/user-roles/:userId", rbacAPIService.GetUserRolesHdl())
		rbac.DELETE("/user-roles/:userId/:roleId", rbacAPIService.RemoveRoleFromUserHdl())
	}
}

func StartGRPCServices(serviceCtx sctx.ServiceContext) {
	configComp := serviceCtx.MustGet(common.KeyCompConf).(common.Config)
	logger := serviceCtx.Logger("grpc")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", configComp.GetGRPCPort()))

	if err != nil {
		log.Fatal(err)
	}

	logger.Infof("GRPC Server is listening on %d ...\n", configComp.GetGRPCPort())

	s := grpc.NewServer()

	pb.RegisterUserServiceServer(s, composer.ComposeUserGRPCService(serviceCtx))
	pb.RegisterAuthServiceServer(s, composer.ComposeAuthGRPCService(serviceCtx))

	if err := s.Serve(lis); err != nil {
		log.Fatalln(err)
	}
}

func Execute() {
	rootCmd.AddCommand(outEnvCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
