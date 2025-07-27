package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"game/models"

	"github.com/redis/go-redis/v9"
)

type RedisGameService struct {
	redisClient *redis.Client
}

func NewRedisGameService(client *redis.Client) *RedisGameService {
	return &RedisGameService{
		redisClient: client,
	}
}

func (r *RedisGameService) SaveGame(ctx context.Context, gameID string, game *models.Game, ttl time.Duration) error {
	data, err := json.Marshal(game)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("game:%s", gameID)
	return r.redisClient.Set(ctx, key, data, ttl).Err()
}

func (r *RedisGameService) GetGame(ctx context.Context, gameID string) (*models.Game, error) {
	key := fmt.Sprintf("game:%s", gameID)
	val, err := r.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var game models.Game
	if err := json.Unmarshal([]byte(val), &game); err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *RedisGameService) DeleteGame(ctx context.Context, gameID string) error {
	key := fmt.Sprintf("game:%s", gameID)
	return r.redisClient.Del(ctx, key).Err()
}

func (r *RedisGameService) GetAllGames(ctx context.Context) ([]*models.Game, error) {
	keys, err := r.redisClient.Keys(ctx, "game:*").Result()
	if err != nil {
		return nil, err
	}
	var games []*models.Game
	for _, key := range keys {
		val, err := r.redisClient.Get(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		var game models.Game
		if err := json.Unmarshal([]byte(val), &game); err != nil {
			return nil, err
		}
		games = append(games, &game)
	}
	return games, nil
}
