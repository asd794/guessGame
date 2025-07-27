package database

import (
	"fmt"
	"game/config"
	"game/models"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitMySQL(cfg *config.MySQL) (*gorm.DB, error) {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database)

	// 連接資料庫
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("MySQL連線成功！")

	// 自動建立資料表（已存在則略過）
	if err := db.AutoMigrate(
		&models.Users{},
		&models.GameResults{},
		&models.GamePlayers{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate tables: %w", err)
	}
	log.Println("資料表檢查/建立完成！")

	return db, nil
}

func CloseMysql(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Println("Failed to get generic database object:", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Println("Failed to close database connection:", err)
	}
}
