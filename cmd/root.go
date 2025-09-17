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
	return sctx.NewServiceContext(
		sctx.WithName("Demo Microservices"),
		sctx.WithComponent(ginc.NewGin(common.KeyCompGIN)),
		sctx.WithComponent(gormc.NewGormDB(common.KeyCompMySQL, "")),
		sctx.WithComponent(jwtc.NewJWT(common.KeyCompJWT)),
		sctx.WithComponent(NewConfig()),
	)
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
	taskAPIService := composer.ComposeTaskAPIService(serviceCtx)
	authAPIService := composer.ComposeAuthAPIService(serviceCtx)
	chatAPIService := composer.ComposeChatbotAPIService(serviceCtx)

	requireAuthMdw := middleware.RequireAuth(composer.ComposeAuthRPCClient(serviceCtx))

	router.POST("/authenticate", authAPIService.LoginHdl())
	router.POST("/register", authAPIService.RegisterHdl())
	router.GET("/profile", requireAuthMdw, userAPIService.GetUserProfileHdl())

	tasks := router.Group("/tasks", requireAuthMdw)
	{
		tasks.GET("", taskAPIService.ListTaskHdl())
		tasks.POST("", taskAPIService.CreateTaskHdl())
		tasks.GET("/:task-id", taskAPIService.GetTaskHdl())
		tasks.PATCH("/:task-id", taskAPIService.UpdateTaskHdl())
		tasks.DELETE("/:task-id", taskAPIService.DeleteTaskHdl())
	}

	chatbot := router.Group("/chat", requireAuthMdw)
	{
		//chatbot.GET("/history", chatAPIService.GetHistory())
		chatbot.POST("/send-message", chatAPIService.SendMessageHandler())
		//chatbot.POST("/create", /*requireAuthMdw,*/ /*chatAPIService.*/chatAPIService.CreateChatHandler())
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
