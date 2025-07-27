package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"game/models"
	"game/repository"
)

// RedisGameManager 使用 Redis 儲存
type RedisGameManager struct {
	redisRepo *repository.RedisGameService
}

func NewRedisGameManager(redisRepo *repository.RedisGameService) *RedisGameManager {
	return &RedisGameManager{
		redisRepo: redisRepo,
	}
}

// 創建遊戲
func (g *RedisGameManager) CreateGame(gameID string, numPlayers int) error {
	game := &models.Game{
		NumOfPeople:    numPlayers,
		Answer:         rand.Intn(100) + 1,
		Round:          0,
		MinRange:       1,
		MaxRange:       100,
		Status:         "waiting",
		Players:        []models.Player{},
		CurrentTurn:    0,
		PlayersGuessed: make(map[string]bool),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return g.redisRepo.SaveGame(ctx, gameID, game, 1*time.Hour)
}

// 玩家加入遊戲
func (g *RedisGameManager) AddPlayer(gameID string, uuid string, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	game, err := g.redisRepo.GetGame(ctx, gameID)
	if err != nil {
		return err
	}
	// 檢查遊戲狀態
	if game.Status == "playing" {
		return fmt.Errorf("遊戲狀態不正確: %s", game.Status)

	}
	// 檢查人數限制
	if len(game.Players) >= game.NumOfPeople {
		fmt.Println("遊戲人數已滿")
		return fmt.Errorf("遊戲人數已滿: %d/%d", len(game.Players), game.NumOfPeople)
	}
	// 檢查玩家是否已存在
	for _, p := range game.Players {
		if uuid == p.Uuid {
			fmt.Println("玩家已存在", uuid)
			return fmt.Errorf("玩家已存在: %s", uuid)
		}
	}

	// 玩家順序
	playerTurnOrder := len(game.Players)
	player := models.Player{
		Uuid:      uuid,
		Name:      name,
		GuessNum:  0,
		Score:     0,
		TurnOrder: playerTurnOrder,
		Guessed:   false,
		Ready:     false,
	}
	game.Players = append(game.Players, player)

	return g.redisRepo.SaveGame(ctx, gameID, game, 1*time.Hour)
}

// 玩家猜數字
func (g *RedisGameManager) GuessNumber(gameID string, uuid string, guess int) (bool, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	game, err := g.redisRepo.GetGame(ctx, gameID)
	if err != nil {
		return false, "", err
	}

	if game.Status != "playing" {
		return false, "遊戲尚未開始", nil
	}

	// var playerIndex int
	var result string

	for i, player := range game.Players {
		if player.Uuid == uuid {
			// playerIndex = i
			if game.PlayersGuessed[uuid] {
				return false, "您已經猜過數字了", nil
			}
			game.PlayersGuessed[uuid] = true
			game.Players[i].Guessed = true
			game.Players[i].GuessNum = guess // 記錄玩家猜測的數字
			break
		} else if i == len(game.Players)-1 {
			return false, "您不在遊戲中", nil
		}
	}

	if guess < game.MinRange || guess > game.MaxRange {
		return false, fmt.Sprintf("猜測數字必須在 %d 到 %d 之間", game.MinRange, game.MaxRange), nil
	}

	if game.Answer == guess {
		game.Status = "finished"
		return true, "恭喜你猜對了！", g.redisRepo.SaveGame(ctx, gameID, game, 1*time.Hour)
	} else if game.Answer < guess {
		result = fmt.Sprintf("猜的數字 %d 太大了", guess)
		// return fmt.Sprintf("猜的數字 %d 太大了，請再試一次", guess), g.redisRepo.SaveGame(gameID, game, 24*time.Hour)
	} else if game.Answer > guess {
		result = fmt.Sprintf("猜的數字 %d 太小了", guess)
		// return fmt.Sprintf("猜的數字 %d 太小了，請再試一次", guess), g.redisRepo.SaveGame(gameID, game, 24*time.Hour)
	}

	game.CurrentTurn = (game.CurrentTurn + 1) % len(game.Players)
	if game.CurrentTurn == 0 {
		game.PlayersGuessed = make(map[string]bool) // 重置玩家猜測狀態
		for i := range game.Players {
			game.Players[i].Guessed = false // 重置玩家猜測狀態
			// game.Players[i].GuessNum = 0     // 重置玩家猜測數字
		}
	}
	return false, result, g.redisRepo.SaveGame(ctx, gameID, game, 1*time.Hour)
}

// 玩家準備或取消準備
func (g *RedisGameManager) PlayerReady(gameID string, uuid string) (*models.Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	game, err := g.redisRepo.GetGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if game.Status == "playing" {
		return nil, fmt.Errorf("遊戲正在進行中，無法準備或取消準備")
	}
	for i, player := range game.Players {
		if player.Uuid == uuid {
			game.Players[i].Ready = !player.Ready
			return game, g.redisRepo.SaveGame(ctx, gameID, game, 1*time.Hour)
		}

	}
	return nil, fmt.Errorf("您不在房間內")

}

// 開始遊戲
func (g *RedisGameManager) StartGame(gameID string) (*models.Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	game, err := g.redisRepo.GetGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if game.Status != "waiting" {
		return nil, fmt.Errorf("遊戲狀態不正確: %s", game.Status)
	}
	for _, player := range game.Players {
		if !player.Ready {
			return nil, fmt.Errorf("有玩家尚未準備好")
		}
	}
	game.Status = "playing"
	game.CurrentTurn = 0
	game.PlayersGuessed = make(map[string]bool)

	return game, g.redisRepo.SaveGame(ctx, gameID, game, 1*time.Hour)
}

// 玩家離開遊戲
func (g *RedisGameManager) PlayerLeave(gameID string, uuid string) (*models.Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	game, err := g.redisRepo.GetGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if game.Status == "playing" {
		return nil, fmt.Errorf("遊戲正在進行中，無法離開")
	}
	for i, player := range game.Players {
		if player.Uuid == uuid {
			if len(game.Players) <= 1 {
				if err := g.redisRepo.DeleteGame(ctx, gameID); err != nil {
					return nil, fmt.Errorf("無法刪除遊戲: %v", err)
				}
				return nil, fmt.Errorf("遊戲已被刪除，因為沒有玩家剩下")
			} else {
				game.Players = append(game.Players[:i], game.Players[i+1:]...)
				break
			}
		}
	}

	for i := range game.Players {
		game.Players[i].TurnOrder = i
	}

	return game, g.redisRepo.SaveGame(ctx, gameID, game, 1*time.Hour)
}

// 玩家強制離開遊戲
func (g *RedisGameManager) PlayerForceLeave(gameID string, uuid string) (*models.Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	game, err := g.redisRepo.GetGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	// if game.Status == "playing" {
	// 	return nil, fmt.Errorf("遊戲正在進行中，無法離開")
	// }
	for i, player := range game.Players {
		if player.Uuid == uuid {
			if len(game.Players) <= 1 {
				if err := g.redisRepo.DeleteGame(ctx, gameID); err != nil {
					return nil, fmt.Errorf("無法刪除遊戲: %v", err)
				}
				return nil, fmt.Errorf("遊戲已被刪除，因為沒有玩家剩下")
			} else {
				game.Players = append(game.Players[:i], game.Players[i+1:]...)
				break
			}
		}
	}

	for i := range game.Players {
		game.Players[i].TurnOrder = i
	}

	return game, g.redisRepo.SaveGame(ctx, gameID, game, 1*time.Hour)
}

// 重置遊戲
func (g *RedisGameManager) ResetGame(gameID string) (*models.Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	game, err := g.redisRepo.GetGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	if game.Status == "playing" {
		return nil, fmt.Errorf("遊戲正在進行中，無法重置")
	}
	game.Round++
	game.Status = "waiting"
	game.CurrentTurn = 0
	game.PlayersGuessed = make(map[string]bool)
	game.Answer = rand.Intn(100) + 1

	for i := range game.Players {
		game.Players[i].TurnOrder = i
		game.Players[i].Guessed = false
		game.Players[i].GuessNum = 0
		game.Players[i].Ready = false
	}

	return game, g.redisRepo.SaveGame(ctx, gameID, game, 1*time.Hour)
}

// 強制重置遊戲
func (g *RedisGameManager) ForceGameReset(gameID string) (*models.Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	game, err := g.redisRepo.GetGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	// if game.Status == "playing" {
	// 	return nil, fmt.Errorf("遊戲正在進行中，無法重置")
	// }
	game.Round++
	game.Status = "waiting"
	game.CurrentTurn = 0
	game.PlayersGuessed = make(map[string]bool)
	game.Answer = rand.Intn(100) + 1

	for i := range game.Players {
		game.Players[i].TurnOrder = i
		game.Players[i].Guessed = false
		game.Players[i].GuessNum = 0
		game.Players[i].Ready = false
	}

	return game, g.redisRepo.SaveGame(ctx, gameID, game, 1*time.Hour)
}

// 獲取特定遊戲狀態
func (g *RedisGameManager) GetAGameStatus(gameID string) (*models.Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	game, err := g.redisRepo.GetGame(ctx, gameID)
	if err != nil {
		return nil, err
	}
	return game, nil
}

// 獲取所有遊戲狀態
func (g *RedisGameManager) GetAllGames() map[string]*models.Game {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	games, err := g.redisRepo.GetAllGames(ctx)
	if err != nil {
		return nil
	}

	// Convert slice to map
	gameMap := make(map[string]*models.Game)
	for i := range games {
		gameMap[fmt.Sprintf("game_%d", i)] = games[i]
	}
	return gameMap
}
