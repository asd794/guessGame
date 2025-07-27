package models

import (
	"time"
)

type Users struct {
	ID           string    `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"`
	Username     string    `gorm:"column:username;unique;size:100;not null" json:"username"`
	PasswordHash string    `gorm:"column:password_hash;size:255;not null" json:"password_hash"`
	Email        *string   `gorm:"column:email;unique;size:100" json:"email,omitempty"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`

	// Relations (optional)
	GameResults []GameResults `gorm:"foreignKey:WinnerID" json:"-"`
	GamePlayers []GamePlayers `gorm:"foreignKey:UserID" json:"-"`
}

type GameResults struct {
	ID           string    `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"`
	GameID       string    `gorm:"column:game_id;type:varchar(36);not null;index:uq_game_round,unique" json:"game_id"`
	WinnerID     *string   `gorm:"column:winner_id;type:varchar(36);index" json:"winner_id,omitempty"`
	Round        int       `gorm:"column:round;not null;index:uq_game_round,unique" json:"round"`
	Answer       int       `gorm:"column:answer;not null" json:"answer"`
	TotalTurns   *int      `gorm:"column:total_turns" json:"total_turns,omitempty"`
	TotalPlayers *int      `gorm:"column:total_players" json:"total_players,omitempty"`
	FinishedAt   time.Time `gorm:"column:finished_at;autoCreateTime" json:"finished_at"`

	// Relations
	Winner  *Users        `gorm:"foreignKey:WinnerID;references:ID" json:"winner,omitempty"`
	Players []GamePlayers `gorm:"foreignKey:GameID,GameResultsRound;references:GameID,Round" json:"players,omitempty"`
}

type GamePlayers struct {
	ID               string `gorm:"column:id;primaryKey;type:varchar(36)" json:"id"`
	GameID           string `gorm:"column:game_id;type:varchar(36);not null;index:idx_game_user_round,unique" json:"game_id"`
	UserID           string `gorm:"column:user_id;type:varchar(36);not null;index:idx_game_user_round,unique" json:"user_id"`
	GameResultsRound int    `gorm:"column:game_results_round;not null;index:idx_game_user_round,unique" json:"game_results_round"`
	TurnOrder        int    `gorm:"column:turn_order;not null" json:"turn_order"`
	Score            *int   `gorm:"column:score;default:0" json:"score,omitempty"`
	GuessCount       *int   `gorm:"column:guess_count;default:0" json:"guess_count,omitempty"`

	// Relations
	User       Users       `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
	GameResult GameResults `gorm:"foreignKey:GameID,GameResultsRound;references:GameID,Round" json:"game_result,omitempty"`
}

type Leaderboard struct {
	// UserID   string
	Username string
	WinCount int64
}
