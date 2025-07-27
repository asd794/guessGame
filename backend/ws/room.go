package ws

import "time"

type Room struct {
	ID        string
	Clients   map[*Client]bool
	CreatedAt time.Time
}

// NewRoom 創建一個新的房間
func NewRoom(roomID string) *Room {
	return &Room{
		ID:        roomID,
		Clients:   make(map[*Client]bool),
		CreatedAt: time.Now(),
	}
}
