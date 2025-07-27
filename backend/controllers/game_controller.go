package controllers

import (
	"game/game"
	"game/services"
	"log"

	"github.com/gin-gonic/gin"
)

type ReqCreate struct {
	NumOfPeople int `json:"num_of_people"`
	MinRange    int `json:"min_range"`
	MaxRange    int `json:"max_range"`
}

type ReqJoin struct {
	GameId     string `json:"game_id"`
	PlayerUuid string `json:"player_uuid"`
	PlayerName string `json:"player_name"`
}

type ReqJoin2 struct {
	GameId string `json:"game_id"`
}

type ReqGuess struct {
	GameId     string `json:"game_id"`
	PlayerUuid string `json:"player_uuid"`
	GuessNum   int    `json:"guess_num"`
}

type ReqStart struct {
	GameId string `json:"game_id"`
}

type ReqGameStatus struct {
	GameId string `json:"game_id"`
}

type GameHandler struct {
	redisGameManager *services.RedisGameManager
	mysqlGameManager *services.GameManagerMysql
}

func NewGameHandlerWithManager(redisGameManager *services.RedisGameManager, mysqlGameManager *services.GameManagerMysql) *GameHandler {
	return &GameHandler{
		redisGameManager: redisGameManager,
		mysqlGameManager: mysqlGameManager,
	}
}

// 創建遊戲控制器
func (g *GameHandler) CreateGameController(c *gin.Context) {

	// 生成唯一遊戲ID
	gameID := game.GenerateGameID()

	err := g.redisGameManager.CreateGame(gameID, 5)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	username := c.GetString("username")
	uuid := c.GetString("uuid")
	err = g.redisGameManager.AddPlayer(gameID, uuid, username)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"game_id": gameID,
		// "game":    game,
		"message": "Game created successfully",
	})

}

// 加入遊戲控制器
func (g *GameHandler) JoinGameController(c *gin.Context) {
	var reqJoin ReqJoin2
	if err := c.ShouldBindJSON(&reqJoin); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	username := c.GetString("username")
	log.Println("Username from context:", username)
	err := g.redisGameManager.AddPlayer(reqJoin.GameId, c.GetString("uuid"), username)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	game, err := g.redisGameManager.GetAGameStatus(reqJoin.GameId)
	if err != nil {
		c.JSON(404, gin.H{"error": "Game not found"})
		return
	}

	c.JSON(200, gin.H{
		"game_id": reqJoin.GameId,
		"game":    game,
		"message": "Player joined successfully",
	})

}

// 獲取所有遊戲控制器
func (g *GameHandler) AllGamesController(c *gin.Context) {
	games := g.redisGameManager.GetAllGames()
	if len(games) == 0 {
		c.JSON(404, gin.H{"error": "No games found"})
		return
	}

	c.JSON(200, gin.H{
		"games": games,
	})
}

// 使用 MySQL 獲取排行榜
func (g *GameHandler) LeaderboardController(c *gin.Context) {
	limit := 10
	topPlayers, err := g.mysqlGameManager.GetTopPlayers(limit)
	if err != nil {
		c.JSON(500, gin.H{"error": "獲取排行榜失敗"})
		return
	}

	c.JSON(200, topPlayers)
}

// 專門的除錯控制器
type DebugController struct {
	wsService *services.NewStruWebSocketService
}

func NewDebugController(wsService *services.NewStruWebSocketService) *DebugController {
	return &DebugController{
		wsService: wsService,
	}
}

// func (d *DebugController) AllController(c *gin.Context) {
// 	fmt.Printf("DebugController: %+v\n", d)

// 	// 檢查 wsService 是否為 nil
// 	if d.wsService == nil {
// 		c.JSON(500, gin.H{
// 			"error": "wsService is nil",
// 		})
// 		return
// 	}

// 	// 安全地獲取 hub
// 	hub := d.wsService.GetChatHub()
// 	if hub == nil {
// 		c.JSON(500, gin.H{
// 			"error": "hub is nil",
// 		})
// 		return
// 	}

// 	// 檢查遊戲管理器
// 	gameManager := d.wsService.GetGameManager()
// 	var gamesInfo map[string]interface{}
// 	if gameManager != nil {
// 		games := gameManager.GetAllGames()
// 		gamesInfo = map[string]interface{}{
// 			"count": len(games),
// 			"games": games,
// 		}
// 	} else {
// 		gamesInfo = map[string]interface{}{
// 			"count": 0,
// 			"error": "gameManager is nil",
// 		}
// 	}

// 	// 安全地處理 hub.Rooms
// 	roomsInfo := map[string]interface{}{
// 		"count": 0,
// 		"rooms": map[string]interface{}{},
// 	}

// 	if hub.Rooms != nil {
// 		roomsInfo["count"] = len(hub.Rooms)
// 		roomsDetail := make(map[string]interface{})

// 		for roomID, room := range hub.Rooms {
// 			if room != nil {
// 				// 安全地獲取客戶端資訊
// 				clientList := make([]map[string]string, 0)
// 				if room.Clients != nil {
// 					for client := range room.Clients {
// 						if client != nil {
// 							clientList = append(clientList, map[string]string{
// 								"player_id":   client.PlayerUuid,
// 								"player_name": client.PlayerName,
// 								"room_id":     client.RoomID,
// 							})
// 						}
// 					}
// 				}

// 				roomsDetail[roomID] = map[string]interface{}{
// 					"room_id":       roomID,
// 					"clients_count": len(clientList),
// 					"clients":       clientList,
// 					"created_at":    room.CreatedAt.Format("2006-01-02 15:04:05"),
// 				}
// 			}
// 		}
// 		roomsInfo["rooms"] = roomsDetail
// 	}

// 	// 返回結構化的資料
// 	c.JSON(200, gin.H{
// 		"status":    "success",
// 		"timestamp": fmt.Sprintf("%v", time.Now().Format("2006-01-02 15:04:05")),
// 		"games":     gamesInfo,
// 		"websocket": roomsInfo,
// 		"debug": map[string]interface{}{
// 			"hub_address":         fmt.Sprintf("%p", hub),
// 			"gameManager_address": fmt.Sprintf("%p", gameManager),
// 			"wsService_address":   fmt.Sprintf("%p", d.wsService),
// 		},
// 	})
// }
