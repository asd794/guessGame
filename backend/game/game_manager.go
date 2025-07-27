package game

import (
	"game/models"
	"sync"
)

// 重新命名為 MemoryGameManager，明確表示這是記憶體版本
type MemoryGameManager struct {
	Games map[string]models.Game
	Mutex sync.RWMutex
}

func NewMemoryGameManager() *MemoryGameManager {
	return &MemoryGameManager{
		Games: make(map[string]models.Game),
	}
}
