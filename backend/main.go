package main

import (
	"fmt"
	"game/config"
	"game/database"
	"game/routes"
)

func main() {

	// Load yaml configuration
	// appConfig, _ := config.LoadConfig()

	appConfig, err := config.LoadConfigEnv()
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}

	db, err := database.InitMySQL(&appConfig.MySQL)
	if err != nil {
		panic(err)
	}
	defer database.CloseMysql(db)
	rds, err := database.InitRedis(&appConfig.Redis)
	if err != nil {
		panic(err)
	}
	defer database.CloseRedis(rds)

	r := routes.SetupRoutes(db, rds)
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}

}
