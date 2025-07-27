package models

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 生產環境請限制來源
	},
}

type GameMessage struct {
	Type        string                   `json:"type"`
	GameId      string                   `json:"gameId"`
	Message     string                   `json:"message"`
	From        string                   `json:"from,omitempty"`
	PlayerName  string                   `json:"playerName,omitempty"`
	PlayerCount int                      `json:"playerCount,omitempty"`
	Timestamp   string                   `json:"timestamp"`
	Players     []map[string]interface{} `json:"players,omitempty"`
	GameInfo    map[string]interface{}   `json:"gameInfo,omitempty"`
}

// 遊戲事件類型常數
const (
	EventAuthenticate = "auth"
	EventPlayerJoined = "player_joined"
	EventPlayerLeft   = "player_left"
	EventGameStarted  = "game_started"
	EventPlayerGuess  = "player_guess"
	EventGameOver     = "game_over"
	EventGameReset    = "game_reset"
	EventGameUpdate   = "game_update"
	EventChat         = "chat"
	EventStartGame    = "start_game"
	EventJoinGame     = "join_game"
	EventLeftGame     = "left_game"
	EventPlayerReady  = "player_ready"
)
