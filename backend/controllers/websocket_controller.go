package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"game/models"
	"game/services"
	"game/ws"

	"github.com/gin-gonic/gin"
)

type WebSocketController struct {
	wsService *services.NewStruWebSocketService
}

func NewWebSocketController(wsService *services.NewStruWebSocketService) *WebSocketController {
	return &WebSocketController{
		wsService: wsService,
	}
}

func (wsc *WebSocketController) HandleWebSocket2(c *gin.Context) {
	log.Println("HandleWebSocket2 called, token:", c.Query("token"))
	// 從查詢參數獲取遊戲房間資訊
	gameID := c.Query("game_id")
	username := c.GetString("username")
	log.Printf("HandleWebSocket2: gameID=%s, username=%s", gameID, username)

	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少遊戲ID"})
		return
	}

	// 如果沒有提供玩家名稱，生成匿名玩家名稱
	if username == "" {
		username = "anonymous"
	}

	playerUuid := c.GetString("uuid")

	log.Printf("玩家 %s 嘗試連接到遊戲 %s", username, gameID)

	// 升級 HTTP 連接為 WebSocket
	conn, err := models.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket 升級失敗: %v", err)
		return
	}

	// 檢查 GameManager 是否存在
	if wsc.wsService == nil {
		log.Printf("錯誤: wsService 為 nil")
		conn.Close()
		return
	}

	// 獲取 Redis GameManager (services.GameManager)
	gameManager := wsc.wsService.GetRedisGameManager()
	if gameManager == nil {
		log.Printf("錯誤: Redis GameManager 為 nil")
		conn.Close()
		return
	}

	// 創建客戶端
	client := &ws.Client{
		ChatHub:    wsc.wsService.GetChatHub(),
		Conn:       conn,
		Send:       make(chan []byte, 256),
		RoomID:     gameID,
		PlayerUuid: playerUuid,
		PlayerName: username,
	}

	log.Printf("客戶端創建成功，準備加入聊天室房間...")

	// 玩家加入遊戲
	client.ChatHub.Join <- client

	// WebSocket 連接成功後發送房間狀態
	go func() {
		time.Sleep(200 * time.Millisecond)
		err := wsc.broadcastRoomStatus(gameID, fmt.Sprintf("玩家 %s 加入了房間", username))
		if err != nil {
			log.Printf("廣播房間狀態失敗: %v", err)
		}
	}()

	// 啟動讀寫協程
	go client.WritePump()
	go client.ReadPump()

	log.Printf("WebSocket 連接建立成功: %s 加入聊天室 %s", username, gameID)
}

// 廣播房間狀態給所有玩家
func (wsc *WebSocketController) broadcastRoomStatus(gameID string, eventMessage string) error {
	// 使用 Redis GameManager (services.GameManager)
	gameManager := wsc.wsService.GetRedisGameManager()
	if gameManager == nil {
		return fmt.Errorf("Redis遊戲管理器未初始化")
	}

	// 獲取遊戲狀態 - 使用 services.GameManager 的方法
	gameState, err := gameManager.GetAGameStatus(gameID)
	if err != nil {
		log.Printf("獲取遊戲狀態失敗: %v", err)
		return err
	}

	// 準備玩家列表數據
	players := make([]map[string]interface{}, 0)
	readyCount := 0
	totalPlayers := len(gameState.Players)

	for _, player := range gameState.Players {
		players = append(players, map[string]interface{}{
			"uuid":    player.Uuid,
			"name":    player.Name,
			"isReady": player.Ready,
		})

		if player.Ready {
			readyCount++
		}
	}

	// 發送系統訊息
	if eventMessage != "" {
		eventMsg := models.GameMessage{
			Type:      "system",
			GameId:    gameID,
			Message:   eventMessage,
			From:      "系統",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		}
		wsc.wsService.GetChatHub().BroadcastGameMessage(gameID, &eventMsg)
		time.Sleep(50 * time.Millisecond)
	}

	// 廣播完整房間狀態
	roomStatusMsg := models.GameMessage{
		Type:    "room_status_update",
		GameId:  gameID,
		Message: fmt.Sprintf("房間狀態更新: %d/%d 玩家，%d/%d 已準備", totalPlayers, gameState.NumOfPeople, readyCount, totalPlayers),
		From:    "系統",
		Players: players,
		GameInfo: map[string]interface{}{
			"maxPlayers":     gameState.NumOfPeople,
			"currentPlayers": totalPlayers,
			"readyCount":     readyCount,
			"gameStatus":     gameState.Status,
			"minRange":       gameState.MinRange,
			"maxRange":       gameState.MaxRange,
		},
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}

	wsc.wsService.GetChatHub().BroadcastGameMessage(gameID, &roomStatusMsg)
	log.Printf("✅ 成功廣播房間狀態到房間 %s", gameID)

	return nil
}

// 保留舊的 WebSocketHandler 作為備用
func WebSocketHandler(c *gin.Context) {
	log.Printf("使用舊的 WebSocketHandler")
	c.JSON(http.StatusOK, gin.H{
		"message": "請使用正確的 WebSocket 端點",
	})
}
