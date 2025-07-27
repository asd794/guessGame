package services

import (
	"fmt"
	"game/models"
	"game/repository"
	"game/utils"
)

type GameManagerMysql struct {
	mysqlRepo *repository.MySQLGameService
}

func NewGameManagerMysql(repo *repository.MySQLGameService) *GameManagerMysql {
	return &GameManagerMysql{mysqlRepo: repo}
}

func (g *GameManagerMysql) GetUsers() ([]models.Users, error) {
	// 這裡可以實現從 MySQL 獲取用戶的邏輯

	users, err := g.mysqlRepo.GetUsers()
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (g *GameManagerMysql) GameResult(gameID string, userID *string, resultRound int, answer int, totalPlayers *int) error {
	resultUUID := utils.GenerateUUID()
	gameResult := models.GameResults{
		ID:           resultUUID,
		GameID:       gameID,
		WinnerID:     userID,
		Round:        resultRound,
		Answer:       answer,
		TotalPlayers: totalPlayers,
	}

	return g.mysqlRepo.AddGameResult(gameResult)
}

func (g *GameManagerMysql) GamePlayer(gameID string, userID string, gameResultRound int, turnOrder int) error {
	resultUUID := utils.GenerateUUID()
	gamePlayer := models.GamePlayers{
		ID:               resultUUID,
		GameID:           gameID,
		UserID:           userID,
		GameResultsRound: gameResultRound,
		TurnOrder:        turnOrder,
	}
	return g.mysqlRepo.AddGamePlayer(gamePlayer)
}

func (g *GameManagerMysql) Login(email string, password string) (string, string, error) {
	user, err := g.mysqlRepo.GetUser(email)
	if err != nil {
		return "", "", fmt.Errorf("email or password is incorrect")
	}

	err = utils.CheckPasswordHash(password, user.PasswordHash)
	if err != nil {
		return "", "", fmt.Errorf("email or password is incorrect")
	}

	return user.ID, user.Username, nil
}

func (g *GameManagerMysql) Register(username string, email string, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("username and password cannot be empty")
	}

	_, err := g.mysqlRepo.GetUser(email)
	if err == nil {
		return fmt.Errorf("email already exists")
	}

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %v", err)
	}

	user := models.Users{
		ID:           utils.GenerateUUID(),
		Username:     username,
		Email:        &email,
		PasswordHash: passwordHash,
	}
	return g.mysqlRepo.CreateUser(user)
}

func (g *GameManagerMysql) GetTopPlayers(limit int) ([]models.Leaderboard, error) {
	return g.mysqlRepo.GetTopPlayers(limit)
}
