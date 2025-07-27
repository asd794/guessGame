package models

type Player struct {
	Uuid      string
	Name      string
	GuessNum  int
	Score     int
	TurnOrder int
	Guessed   bool
	Ready     bool
}

type Game struct {
	NumOfPeople    int
	Answer         int
	Round          int
	MinRange       int
	MaxRange       int
	Status         string
	Players        []Player
	CurrentTurn    int
	PlayersGuessed map[string]bool
}

type Message struct {
	Type    string      `json:"type"`
	GameId  string      `json:"gameId"`
	Message interface{} `json:"message"`
	From    string      `json:"from,omitempty"`
}

// 添加方法到 Game 結構體
func (g *Game) GetCurrentPlayer() *Player {
	if len(g.Players) == 0 || g.CurrentTurn >= len(g.Players) {
		return nil
	}
	return &g.Players[g.CurrentTurn]
}
