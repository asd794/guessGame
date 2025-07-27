package ws

import (
	"encoding/json"
	"fmt"
	"game/models"
	"log"
	"sync"
	"time"
)

// 定義接口
type GameManager interface {
	CreateGame(gameID string, numPlayers int) error
	AddPlayer(gameID string, uuid string, name string) error
	GetAGameStatus(gameID string) (*models.Game, error)
	PlayerReady(gameID string, uuid string) (*models.Game, error)
	StartGame(gameID string) (*models.Game, error)
	PlayerLeave(gameID string, uuid string) (*models.Game, error)
	PlayerForceLeave(gameID string, uuid string) (*models.Game, error)
	GuessNumber(gameID string, uuid string, guess int) (bool, string, error)
	ResetGame(gameID string) (*models.Game, error)
	ForceGameReset(gameID string) (*models.Game, error)
}

type MySQLGameService interface {
	GetUsers() ([]models.Users, error)
	GameResult(gameID string, userID *string, resultRound int, answer int, totalPlayers *int) error
	GamePlayer(gameID string, userID string, gameResultRound int, turnOrder int) error
}

type ChatHub struct {
	Rooms        map[string]*Room
	Join         chan *Client
	Leave        chan *Client
	Broadcast    chan []byte
	mu           sync.RWMutex
	GameManager  GameManager
	MySQLService MySQLGameService
}

func NewChatHub(gameManager GameManager, mySQLService MySQLGameService) *ChatHub {
	return &ChatHub{
		Rooms:        make(map[string]*Room),
		Join:         make(chan *Client, 256),
		Leave:        make(chan *Client, 256),
		Broadcast:    make(chan []byte, 256),
		GameManager:  gameManager,
		MySQLService: mySQLService,
	}
}

func (h *ChatHub) Run() {
	log.Printf("ChatHub 正在運行...")

	for {
		select {
		case client := <-h.Join:
			log.Printf("收到加入請求: %s 要加入房間 %s", client.PlayerName, client.RoomID)

			if _, ok := h.Rooms[client.RoomID]; !ok {
				h.Rooms[client.RoomID] = NewRoom(client.RoomID)
				log.Printf("創建新房間: %s", client.RoomID)
			}
			h.Rooms[client.RoomID].Clients[client] = true

			log.Printf("玩家 %s 加入聊天室 %s，目前聊天室人數：%d",
				client.PlayerName, client.RoomID, len(h.Rooms[client.RoomID].Clients))

			gameMsg := models.GameMessage{
				Type:        models.EventPlayerJoined,
				GameId:      client.RoomID,
				Message:     fmt.Sprintf("玩家 %s 加入了聊天室", client.PlayerName),
				From:        "系統",
				PlayerName:  client.PlayerName,
				PlayerCount: len(h.Rooms[client.RoomID].Clients),
				Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
			}

			h.BroadcastGameMessage(client.RoomID, &gameMsg)

		case client := <-h.Leave:
			log.Printf("收到離開請求: %s 要離開聊天室 %s", client.PlayerName, client.RoomID)

			h.mu.Lock()
			if room, ok := h.Rooms[client.RoomID]; ok {
				if _, exists := room.Clients[client]; exists {
					delete(room.Clients, client)
					close(client.Send)

					go func(roomID, playerName string) {
						time.Sleep(100 * time.Millisecond)
						h.broadcastRoomStatusAfterLeave(roomID)
					}(client.RoomID, client.PlayerName)

					log.Printf("玩家 %s 已從房間 %s 移除", client.PlayerName, client.RoomID)
				}
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			log.Printf("收到廣播訊息: %+v", message)
			var msg models.Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("解析廣播訊息失敗: %v", err)
				continue
			}
			h.BroadcastToRoom(msg.GameId, &msg)
		}
	}
}

func (h *ChatHub) BroadcastToRoom(roomID string, message *models.Message) {
	if room, ok := h.Rooms[roomID]; ok {
		jsonMessage, _ := json.Marshal(message)
		for client := range room.Clients {
			select {
			case client.Send <- jsonMessage:
			default:
				close(client.Send)
				delete(room.Clients, client)
			}
		}
	}
}

// 廣播遊戲訊息到指定房間
func (h *ChatHub) BroadcastGameMessage(roomID string, gameMsg *models.GameMessage) {
	if room, ok := h.Rooms[roomID]; ok {
		jsonMessage, err := json.Marshal(gameMsg)
		if err != nil {
			log.Printf("序列化訊息失敗: %v", err)
			return
		}

		log.Printf("廣播到房間 %s: %s", roomID, string(jsonMessage))
		for client := range room.Clients {
			select {
			case client.Send <- jsonMessage:
			default:
				close(client.Send)
				delete(room.Clients, client)
			}
		}
	}
}

func (h *ChatHub) broadcastRoomStatusAfterLeave(roomID string) {
	gameState, err := h.GameManager.GetAGameStatus(roomID)
	if err != nil {
		log.Printf("獲取遊戲狀態失敗: %v", err)
		return
	}

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

	roomStatusMsg := models.GameMessage{
		Type:    "room_status_update",
		GameId:  roomID,
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

	h.BroadcastGameMessage(roomID, &roomStatusMsg)
	log.Printf("✅ 成功廣播玩家離開後的房間狀態")
}
