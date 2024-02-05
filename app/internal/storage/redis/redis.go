package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/UdinSemen/moscow-events-backend/internal/config"
	"github.com/UdinSemen/moscow-events-backend/internal/domain/models"
	"github.com/redis/go-redis/v9"
)

const (
	timeCodeTTL   = "90s" //seconds
	timeCodeTable = "time_codes."
)

type Redis struct {
	rdb *redis.Client
}

func NewRedisClient(cfg *config.Config) *Redis {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DbName,
	})

	return &Redis{rdb: rdb}
}

func (s *Redis) Ping(ctx context.Context) error {
	const op = "storage.redis.Ping"
	if err := s.rdb.Ping(ctx); err.Err() != nil {
		return fmt.Errorf("%s:%w", op, err.Err())
	}
	return nil
}

func (s *Redis) CreateRegSession(ctx context.Context, fingerPrint, timeCode string) error {
	const op = "storage.redis.CreateRegSession"

	dur, err := time.ParseDuration(timeCodeTTL)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	jsonRow, err := json.Marshal(models.RegSession{
		FingerPrint: fingerPrint,
		IsConfirmed: false,
	})

	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}
	if err := s.rdb.SetEx(ctx, timeCodeTable+timeCode, jsonRow, dur).Err(); err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}

func (s *Redis) GetRegSession(ctx context.Context, timeCode string) (string, error) {
	const op = "storage.redis.GetRegSession"

	val, err := s.rdb.Get(ctx, timeCodeTable+timeCode).Result()
	if err != nil {
		return "", fmt.Errorf("%s:%w", op, err)
	}
	return val, nil
}

func (s *Redis) Close() error {
	return s.rdb.Close()
}
