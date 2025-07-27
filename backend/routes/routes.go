package routes

import (
	"game/controllers"
	"game/middleware"
	"game/repository"
	"game/services"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB, rds *redis.Client) *gin.Engine {
	route := gin.Default()

	// 初始化 RedisGameService
	redisGameService := repository.NewRedisGameService(rds)
	// 初始化 MySQLGameService
	mysqlGameService := repository.NewMySQLGameRepository(db)
	// 使用 RedisGameService 初始化 RedisGameManager
	redisGameManager := services.NewRedisGameManager(redisGameService)
	// 初始化新的 WebSocket 服務
	websocketService := services.NewWebSocketService(redisGameService, mysqlGameService)
	// 啟動
	websocketService.StartChatHub()

	// 創建控制器，使用相同的遊戲管理器
	gameHandler := controllers.NewGameHandlerWithManager(redisGameManager, services.NewGameManagerMysql(mysqlGameService))
	wsController := controllers.NewWebSocketController(websocketService)

	// debugController := controllers.NewDebugController(wsService)

	// 初始化 authController
	authController := controllers.NewAuthController(services.NewGameManagerMysql(mysqlGameService))

	// CORS 中間件
	route.Use(middleware.CORS())

	api := route.Group("/api")
	{
		v1 := api.Group("/v1")
		{

			// 認證相關路由
			auth := v1.Group("/auth")
			{
				auth.POST("/login", authController.LoginController)
				auth.POST("/register", authController.RegisterController)
				auth.POST("/captcha", authController.GetCaptchaController)
			}
			auth.Use(middleware.JWTAuthGame()) // 使用 JWT 認證中間件
			{
				auth.POST("/allGames", gameHandler.AllGamesController)
				auth.GET("/leaderboard", gameHandler.LeaderboardController)
				auth.POST("/createGame", gameHandler.CreateGameController)
				auth.POST("/joinGame", gameHandler.JoinGameController)
				auth.GET("/wsGame", wsController.HandleWebSocket2)
			}

		}
	}

	return route
}
