package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/UdinSemen/moscow-events-backend/internal/domain/models"
	"github.com/UdinSemen/moscow-events-backend/internal/storage"
	"github.com/UdinSemen/moscow-events-backend/pkg/utils"
)

const (
	timeCodeLen = 6
)

var (
	ErrNoRegSession        = errors.New("session doesn't found with same fingerprint and time code")
	ErrInvalidFingerPrint  = errors.New("invalid fingerprint")
	ErrSessionNotConfirmed = errors.New("session not confirmed")
	ErrInvalidUserID       = errors.New("invalid userID")
)

type AuthService struct {
	redis    storage.Redis
	postgres storage.PgStorage
}

func NewAuth(redis storage.Redis, postgres storage.PgStorage) *AuthService {
	return &AuthService{
		redis:    redis,
		postgres: postgres,
	}
}

func (s *AuthService) CreateRegSession(ctx context.Context, fingerPrint string) (string, error) {
	timeCode := utils.GenTimeCode(timeCodeLen)
	return timeCode, s.redis.CreateRegSession(ctx, fingerPrint, timeCode)
}

func (s *AuthService) GetRegSession(ctx context.Context, fingerPrint, timeCode string) (string, error) {
	const op = "services.GetRegSession"

	val, err := s.redis.GetRegSession(ctx, timeCode)
	if err != nil {
		return "", fmt.Errorf("%s:%w", op, err)
	}

	if val == "" {
		return "", ErrNoRegSession
	}

	var session models.RegSession
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return "", fmt.Errorf("%s:%w", op, err)
	}

	if session.FingerPrint != fingerPrint {
		return "", ErrInvalidFingerPrint
	}

	if !session.IsConfirmed {
		return "", ErrSessionNotConfirmed
	}

	userID := session.UserID
	if userID == "" {
		return "", ErrInvalidUserID
	}
	return userID, nil
}

func (s *AuthService) InitUser(ctx context.Context, userID, refreshToken string) error {
	const op = "services.InitUser"

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return s.postgres.InitUser(ctx, userIDInt, refreshToken)
}
