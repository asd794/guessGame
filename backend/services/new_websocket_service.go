package services

import (
	"game/repository"
	"game/ws"
	"log"
)

// NewStruWebSocketService 使用新的 WebSocket 服務
type NewStruWebSocketService struct {
	chatHub          *ws.ChatHub
	redisGameService *repository.RedisGameService // Redis 服務
	redisGameManager *RedisGameManager            // Redis GameManager
}

func NewWebSocketService(redisGameService *repository.RedisGameService, mySQLService *repository.MySQLGameService) *NewStruWebSocketService {
	redisGameManager := NewRedisGameManager(redisGameService) // 使用 RedisGameService 初始化 RedisGameManager
	mysqlGameManager := NewGameManagerMysql(mySQLService)     // 使用 MySQLGameService 初始化 GameManager
	// 將 RedisGameManager 和 MySQLGameManager 作為接口傳入
	chatHub := ws.NewChatHub(redisGameManager, mysqlGameManager)

	return &NewStruWebSocketService{
		chatHub:          chatHub,
		redisGameService: redisGameService,
		redisGameManager: redisGameManager,
	}
}

func (s *NewStruWebSocketService) StartChatHub() {
	log.Printf("Starting ChatHub...")
	go s.chatHub.Run()
	log.Printf("ChatHub started")
}

func (s *NewStruWebSocketService) GetChatHub() *ws.ChatHub {
	return s.chatHub
}

// 獲取 Redis GameManager (services.GameManager)
func (s *NewStruWebSocketService) GetRedisGameManager() *RedisGameManager {
	return s.redisGameManager
}

// // 保持向後相容性，逐步淘汰
// func (s *NewStruWebSocketService) GetGameManager() *repository.RedisGameService {
// 	return s.redisGameService
// }
