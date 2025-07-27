package database

import (
	"context"
	"fmt"
	"game/config"
	"log"

	"github.com/redis/go-redis/v9"
)

func InitRedis(cfg *config.Redis) (*redis.Client, error) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 測試 Redis 連線
	if err := rdb.Ping(context.Background()).Err(); err != nil {

		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	log.Println("Redis連線成功！")

	// 清除所有 game: 開頭的 key
	keys, err := rdb.Keys(context.Background(), "game:*").Result()
	if err != nil {
		log.Printf("取得 game:* keys 失敗: %v\n", err)
	} else if len(keys) > 0 {
		if err := rdb.Del(context.Background(), keys...).Err(); err != nil {
			log.Printf("刪除 game:* keys 失敗: %v\n", err)
		} else {
			log.Printf("已清除 %d 個 game:* keys\n", len(keys))
		}
	}

	return rdb, nil
}

func CloseRedis(rdb *redis.Client) {
	// 關閉 Redis 連線
	if err := rdb.Close(); err != nil {
		fmt.Printf("Failed to close Redis connection: %v\n", err)
	}
}
