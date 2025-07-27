package game

import "github.com/teris-io/shortid"

func GenerateGameID() string {
	sid, err := shortid.New(1, shortid.DefaultABC, 2342)
	if err != nil {
		panic(err)
	}
	roomID, err := sid.Generate()
	if err != nil {
		panic(err)
	}
	return roomID
}
