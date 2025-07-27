package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	MySQL MySQL `yaml:"mysql"`
	Redis Redis `yaml:"redis"`
}

type MySQL struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type Redis struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func LoadConfig() (Config, error) {
	var appConfig Config
	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		log.Println(err)
		return Config{}, err
	}
	err = yaml.Unmarshal(data, &appConfig)
	if err != nil {
		log.Println(err)
	}
	return appConfig, nil
}

func LoadConfigEnv() (Config, error) {
	if _, err := os.Stat(".env"); err == nil {
		// .env 檔案存在，才載入
		err = godotenv.Load()
		if err != nil {
			return Config{}, err
		}
	} else if !os.IsNotExist(err) {
		// 發生其他錯誤（不是檔案不存在）
		return Config{}, err
	}

	var appConfig Config
	appConfig.MySQL.Host = os.Getenv("MYSQL_HOST")
	appConfig.MySQL.Port = os.Getenv("MYSQL_PORT")
	appConfig.MySQL.User = os.Getenv("MYSQL_USER")
	appConfig.MySQL.Password = os.Getenv("MYSQL_PASSWORD")
	appConfig.MySQL.Database = os.Getenv("MYSQL_DATABASE")

	appConfig.Redis.Host = os.Getenv("REDIS_HOST")
	appConfig.Redis.Port, _ = strconv.Atoi(os.Getenv("REDIS_PORT"))
	appConfig.Redis.Password = os.Getenv("REDIS_PASSWORD")
	appConfig.Redis.DB, _ = strconv.Atoi(os.Getenv("REDIS_DB"))

	return appConfig, nil
}
