package repository

import (
	"game/models"
	"log"

	"gorm.io/gorm"
)

type MySQLGameService struct {
	db *gorm.DB
}

func NewMySQLGameRepository(db *gorm.DB) *MySQLGameService {
	return &MySQLGameService{db: db}
}

func (r *MySQLGameService) GetUsers() ([]models.Users, error) {
	var users []models.Users
	err := r.db.Find(&users).Error
	return users, err
}

func (r *MySQLGameService) CreateGame(game models.Game) error {
	return r.db.Create(&game).Error
}

func (r *MySQLGameService) GetGame(gameID string) (models.Game, error) {
	var game models.Game
	err := r.db.First(&game, "id = ?", gameID).Error
	return game, err
}

func (r *MySQLGameService) UpdateGame(game models.Game) error {
	return r.db.Save(&game).Error
}

func (r *MySQLGameService) DeleteGame(gameID string) error {
	return r.db.Delete(&models.Game{}, gameID).Error
}

func (r *MySQLGameService) AddGameResult(gameResult models.GameResults) error {
	return r.db.Save(&gameResult).Error
}

func (r *MySQLGameService) AddGamePlayer(gamePlayer models.GamePlayers) error {
	return r.db.Create(&gamePlayer).Error
}

func (r *MySQLGameService) GetUser(email string) (models.Users, error) {
	var user models.Users
	err := r.db.Select("id", "email", "password_hash", "username").First(&user, "email = ?", email).Error
	return user, err
}

func (r *MySQLGameService) CreateUser(user models.Users) error {
	return r.db.Create(&user).Error
}

func (r *MySQLGameService) GetTopPlayers(limit int) ([]models.Leaderboard, error) {
	var leaderboard []models.Leaderboard

	err := r.db.Table("game_results AS gr").
		Select("u.id AS user_id, u.username, COUNT(gr.id) AS win_count").
		Joins("JOIN users u ON gr.winner_id = u.id").
		Where("gr.winner_id IS NOT NULL").
		Group("u.id, u.username").
		Order("win_count DESC").
		Limit(10).
		Scan(&leaderboard).Error

	if err != nil {
		log.Println("查詢排行榜失敗:", err)
	}
	return leaderboard, err
}
