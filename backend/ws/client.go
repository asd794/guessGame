package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"game/models"

	"github.com/gorilla/websocket"
)

type Client struct {
	ChatHub    *ChatHub
	Send       chan []byte
	RoomID     string
	PlayerUuid string
	PlayerName string
	Conn       *websocket.Conn
}

// ReadPump 處理從客戶端接收的訊息
func (c *Client) ReadPump() {
	var leftGame bool // 是否離開遊戲
	defer func() {
		// 強制關閉瀏覽器斷線websocket連接
		c.ChatHub.Leave <- c
		if !leftGame {
			c.handleForceLeftGame()
			c.handleForceGameReset()
		}
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		log.Printf("收到客戶端訊息: %s", string(message))

		var msg models.Message
		if err := json.Unmarshal(message, &msg); err == nil {
			msg.GameId = c.RoomID
			msg.From = c.PlayerName

			switch msg.Type {
			case models.EventChat:
				c.handleChat(msg)
			case models.EventAuthenticate:
				c.handleAuthenticate(msg)
			case models.EventJoinGame:
				c.handleJoinGame()
			case models.EventLeftGame:
				leftGame = true
				c.handleLeftGame()
			case models.EventStartGame:
				c.handleStartGame()
			case models.EventPlayerGuess:
				c.handlePlayerGuess(msg)
			case models.EventPlayerReady:
				c.handleGameReady(msg)
			case models.EventGameReset:
				c.handleGameReset()
			default:
				c.ChatHub.BroadcastToRoom(c.RoomID, &msg)
			}
		} else {
			log.Printf("解析訊息失敗: %v", err)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("發送訊息失敗給 %s: %v", c.PlayerName, err)
				return
			}

			log.Printf("✅ 發送訊息給 %s: %s", c.PlayerName, string(message))

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 處理方法
func (c *Client) handleChat(msg models.Message) {
	var messageContent string
	switch v := msg.Message.(type) {
	case string:
		messageContent = v
	case map[string]interface{}:
		if msgText, ok := v["text"].(string); ok {
			messageContent = msgText
		} else {
			messageContent = fmt.Sprintf("%v", v)
		}
	default:
		messageContent = fmt.Sprintf("%v", v)
	}

	c.ChatHub.BroadcastGameMessage(c.RoomID, &models.GameMessage{
		Type:       models.EventChat,
		GameId:     c.RoomID,
		Message:    messageContent,
		From:       c.PlayerName,
		PlayerName: c.PlayerName,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (c *Client) handleAuthenticate(msg models.Message) {
	log.Printf("玩家 %s 認證成功，加入遊戲 %s", msg.From, msg.GameId)

	authMsg := models.GameMessage{
		Type:       models.EventAuthenticate,
		GameId:     msg.GameId,
		Message:    "認證成功",
		From:       "系統",
		PlayerName: msg.From,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}
	c.ChatHub.BroadcastGameMessage(msg.GameId, &authMsg)
}

func (c *Client) handleJoinGame() {
	err := c.ChatHub.GameManager.AddPlayer(c.RoomID, c.PlayerUuid, c.PlayerName)
	if err != nil {
		errorMsg := models.GameMessage{
			Type:      "error",
			GameId:    c.RoomID,
			Message:   err.Error(),
			From:      "系統",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		}
		c.ChatHub.BroadcastGameMessage(c.RoomID, &errorMsg)
		return
	}

	gameState, _ := c.ChatHub.GameManager.GetAGameStatus(c.RoomID)
	joinMsg := models.GameMessage{
		Type:        models.EventPlayerJoined,
		GameId:      c.RoomID,
		Message:     fmt.Sprintf("玩家 %s 加入了遊戲", c.PlayerName),
		From:        "系統",
		PlayerName:  c.PlayerName,
		PlayerCount: len(gameState.Players),
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
	}
	c.ChatHub.BroadcastGameMessage(c.RoomID, &joinMsg)
}

func (c *Client) handleLeftGame() {
	_, err := c.ChatHub.GameManager.PlayerLeave(c.RoomID, c.PlayerUuid)
	if err != nil {
		errorMsg := models.GameMessage{
			Type:      "error",
			GameId:    c.RoomID,
			Message:   fmt.Sprintf("離開遊戲失敗: %s", err.Error()),
			From:      "系統",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		}
		c.ChatHub.BroadcastGameMessage(c.RoomID, &errorMsg)
		return
	}

	gameState, err := c.ChatHub.GameManager.GetAGameStatus(c.RoomID)
	if err != nil {
		return
	}

	leftMsg := models.GameMessage{
		Type:        models.EventPlayerLeft,
		GameId:      c.RoomID,
		Message:     fmt.Sprintf("玩家 %s 離開了遊戲", c.PlayerName),
		From:        "系統",
		PlayerName:  c.PlayerName,
		PlayerCount: len(gameState.Players),
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
	}
	c.ChatHub.BroadcastGameMessage(c.RoomID, &leftMsg)
}

func (c *Client) handleStartGame() {
	game, err := c.ChatHub.GameManager.StartGame(c.RoomID)
	if err != nil {
		errorMsg := models.GameMessage{
			Type:      "error",
			GameId:    c.RoomID,
			Message:   err.Error(),
			From:      "系統",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		}
		c.ChatHub.BroadcastGameMessage(c.RoomID, &errorMsg)
		return
	}

	startMsg := models.GameMessage{
		Type:      models.EventGameStarted,
		GameId:    c.RoomID,
		Message:   "遊戲開始了！所有玩家可以開始猜數字",
		From:      "系統",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		GameInfo: map[string]interface{}{
			"Players": game.Players,
		},
	}
	c.ChatHub.BroadcastGameMessage(c.RoomID, &startMsg)
}

func (c *Client) handlePlayerGuess(msg models.Message) {
	var guessNum int
	var err error

	switch v := msg.Message.(type) {
	case string:
		guessNum, err = strconv.Atoi(v)
	case float64:
		guessNum = int(v)
	case int:
		guessNum = v
	default:
		errorMsg := models.GameMessage{
			Type:      "error",
			GameId:    c.RoomID,
			Message:   "猜測格式錯誤",
			From:      "系統",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		}
		c.ChatHub.BroadcastGameMessage(c.RoomID, &errorMsg)
		return
	}

	if err != nil {
		errorMsg := models.GameMessage{
			Type:      "error",
			GameId:    c.RoomID,
			Message:   "請輸入有效的數字",
			From:      "系統",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		}
		c.ChatHub.BroadcastGameMessage(c.RoomID, &errorMsg)
		return
	}

	isCorrect, result, err := c.ChatHub.GameManager.GuessNumber(c.RoomID, c.PlayerUuid, guessNum)
	if err != nil {
		errorMsg := models.GameMessage{
			Type:      "error",
			GameId:    c.RoomID,
			Message:   err.Error(),
			From:      "系統",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		}
		c.ChatHub.BroadcastGameMessage(c.RoomID, &errorMsg)
		return
	}

	game, _ := c.ChatHub.GameManager.GetAGameStatus(c.RoomID)

	var eventType string
	if isCorrect {
		eventType = models.EventGameOver
		// 儲存遊戲結果到 MySQL
		go func() {

			totalPlayers := len(game.Players)
			err = c.ChatHub.MySQLService.GameResult(c.RoomID, &c.PlayerUuid, game.Round, guessNum, &totalPlayers)
			if err != nil {
				log.Printf("儲存遊戲結果到 MySQL 失敗: %v", err)
			}
			for _, player := range game.Players {
				err = c.ChatHub.MySQLService.GamePlayer(c.RoomID, player.Uuid, game.Round, player.TurnOrder)
				if err != nil {
					log.Printf("儲存玩家參與結果到 MySQL 失敗: %v", err)
				}
			}

		}()

	} else {
		eventType = models.EventPlayerGuess
	}

	guessMsg := models.GameMessage{
		Type:       eventType,
		GameId:     c.RoomID,
		Message:    fmt.Sprintf("玩家 %s 猜測 %d，結果：%s", c.PlayerName, guessNum, result),
		From:       "系統",
		PlayerName: c.PlayerName,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}
	c.ChatHub.BroadcastGameMessage(c.RoomID, &guessMsg)

	var turnIndex int
	for i, player := range game.Players {
		if player.Uuid == c.PlayerUuid {
			turnIndex = (i + 1) % len(game.Players) // 下一個玩家的索引
			break
		}
	}
	if game.Status == "finished" {
		return
	}
	turnMsg := models.GameMessage{
		Type:      "player_turn",
		GameId:    c.RoomID,
		Message:   fmt.Sprintf("輪到 %s 猜測", game.Players[turnIndex].Name),
		From:      "系統",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		GameInfo: map[string]interface{}{
			"CurrentTurn": game.CurrentTurn,
		},
	}
	c.ChatHub.BroadcastGameMessage(c.RoomID, &turnMsg)

}

func (c *Client) handleGameReady(msg models.Message) {
	game, err := c.ChatHub.GameManager.PlayerReady(c.RoomID, c.PlayerUuid)
	if err != nil {
		errorMsg := models.GameMessage{
			Type:      "error",
			GameId:    c.RoomID,
			Message:   err.Error(),
			From:      "系統",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		}
		c.ChatHub.BroadcastGameMessage(c.RoomID, &errorMsg)
		return
	}

	var readyPlayer *models.Player
	for _, player := range game.Players {
		if player.Uuid == c.PlayerUuid {
			readyPlayer = &player
			break
		}
	}

	var statusMessage string
	if readyPlayer != nil && readyPlayer.Ready {
		statusMessage = fmt.Sprintf("玩家 %s 已準備就緒", c.PlayerName)
	} else {
		statusMessage = fmt.Sprintf("玩家 %s 取消準備", c.PlayerName)
	}

	readyMsg := models.GameMessage{
		Type:       models.EventPlayerReady,
		GameId:     c.RoomID,
		Message:    statusMessage,
		From:       "系統",
		PlayerName: c.PlayerName,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}
	c.ChatHub.BroadcastGameMessage(c.RoomID, &readyMsg)
}

func (c *Client) handleGameReset() {
	gameManager := c.ChatHub.GameManager
	if gameManager == nil {
		log.Printf("錯誤: GameManager 為 nil，無法重置遊戲")
		return
	}

	game, err := gameManager.ResetGame(c.RoomID)
	if err != nil {
		errorMsg := models.GameMessage{
			Type:      "error",
			GameId:    c.RoomID,
			Message:   fmt.Sprintf("重置遊戲失敗: %s", err.Error()),
			From:      "系統",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		}
		c.ChatHub.BroadcastGameMessage(c.RoomID, &errorMsg)
		return
	}

	resetMsg := models.GameMessage{
		Type:      models.EventGameReset,
		GameId:    c.RoomID,
		Message:   "遊戲已重置，請重新開始",
		From:      "系統",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
	c.ChatHub.BroadcastGameMessage(c.RoomID, &resetMsg)

	players := make([]map[string]interface{}, 0)
	readyCount := 0
	for _, player := range game.Players {
		players = append(players, map[string]interface{}{
			"uuid":    player.Uuid,
			"name":    player.Name,
			"isReady": player.Ready,
		})

		if player.Ready {
			readyCount++
		}
	}
	// game, _ := gameManager.GetAGameStatus(c.RoomID)
	roomStatusMsg := models.GameMessage{
		Type:    "room_status_update",
		GameId:  c.RoomID,
		Message: fmt.Sprintf("房間狀態更新: %d/%d 玩家，%d/%d 已準備", len(game.Players), game.NumOfPeople, readyCount, len(game.Players)),
		From:    "系統",
		Players: players,
		GameInfo: map[string]interface{}{
			"maxPlayers":     game.NumOfPeople,
			"currentPlayers": len(game.Players),
			"gameStatus":     game.Status,
			"minRange":       game.MinRange,
			"maxRange":       game.MaxRange,
		},
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
	c.ChatHub.BroadcastGameMessage(c.RoomID, &roomStatusMsg)
}

func (c *Client) handleForceLeftGame() {
	_, err := c.ChatHub.GameManager.PlayerForceLeave(c.RoomID, c.PlayerUuid)
	if err != nil {
		errorMsg := models.GameMessage{
			Type:      "error",
			GameId:    c.RoomID,
			Message:   fmt.Sprintf("離開遊戲失敗: %s", err.Error()),
			From:      "系統",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		}
		c.ChatHub.BroadcastGameMessage(c.RoomID, &errorMsg)
		return
	}

	gameState, err := c.ChatHub.GameManager.GetAGameStatus(c.RoomID)
	if err != nil {
		return
	}

	leftMsg := models.GameMessage{
		Type:        models.EventPlayerLeft,
		GameId:      c.RoomID,
		Message:     fmt.Sprintf("玩家 %s 離開了遊戲", c.PlayerName),
		From:        "系統",
		PlayerName:  c.PlayerName,
		PlayerCount: len(gameState.Players),
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
	}
	c.ChatHub.BroadcastGameMessage(c.RoomID, &leftMsg)
}

func (c *Client) handleForceGameReset() {
	gameManager := c.ChatHub.GameManager
	if gameManager == nil {
		log.Printf("錯誤: GameManager 為 nil，無法重置遊戲")
		return
	}

	game, err := gameManager.ForceGameReset(c.RoomID)
	if err != nil {
		errorMsg := models.GameMessage{
			Type:      "error",
			GameId:    c.RoomID,
			Message:   fmt.Sprintf("重置遊戲失敗: %s", err.Error()),
			From:      "系統",
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		}
		c.ChatHub.BroadcastGameMessage(c.RoomID, &errorMsg)
		return
	}

	resetMsg := models.GameMessage{
		Type:      models.EventGameReset,
		GameId:    c.RoomID,
		Message:   "遊戲已重置，請重新開始",
		From:      "系統",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
	c.ChatHub.BroadcastGameMessage(c.RoomID, &resetMsg)

	players := make([]map[string]interface{}, 0)
	readyCount := 0
	for _, player := range game.Players {
		players = append(players, map[string]interface{}{
			"uuid":    player.Uuid,
			"name":    player.Name,
			"isReady": player.Ready,
		})

		if player.Ready {
			readyCount++
		}
	}
	// game, _ := gameManager.GetAGameStatus(c.RoomID)
	roomStatusMsg := models.GameMessage{
		Type:    "room_status_update",
		GameId:  c.RoomID,
		Message: fmt.Sprintf("房間狀態更新: %d/%d 玩家，%d/%d 已準備", len(game.Players), game.NumOfPeople, readyCount, len(game.Players)),
		From:    "系統",
		Players: players,
		GameInfo: map[string]interface{}{
			"maxPlayers":     game.NumOfPeople,
			"currentPlayers": len(game.Players),
			"gameStatus":     game.Status,
			"minRange":       game.MinRange,
			"maxRange":       game.MaxRange,
		},
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
	c.ChatHub.BroadcastGameMessage(c.RoomID, &roomStatusMsg)
}
