package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/UdinSemen/moscow-events-backend/internal/domain/models"
	"github.com/UdinSemen/moscow-events-backend/internal/storage"
	"github.com/UdinSemen/moscow-events-backend/pkg/utils"
	"github.com/redis/go-redis/v9"
)

const (
	timeCodeLen = 6
)

var (
	ErrNoRegSession         = errors.New("session doesn't found with same fingerprint and time code")
	ErrInvalidFingerPrint   = errors.New("invalid fingerprint")
	ErrSessionNotConfirmed  = errors.New("session not confirmed")
	ErrInvalidUserID        = errors.New("invalid userID")
	ErrRefreshTokenExp      = errors.New("refresh token expired")
	ErrDifferentFingerPrint = errors.New("different fingerprint")
)

type AuthService struct {
	refreshTokenTTL time.Duration
	redis           storage.Redis
	postgres        storage.PgStorage
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
		if errors.Is(err, redis.Nil) {
			return "", ErrNoRegSession
		} else {
			return "", fmt.Errorf("%s:%w", op, err)
		}
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

	userTgId := session.UserID
	if userTgId == "" {
		return "", ErrInvalidUserID
	}
	userID, err := strconv.ParseInt(userTgId, 10, 64)
	if err != nil {
		return "", fmt.Errorf("%s:%w", op, err)
	}
	userDTO, err := s.postgres.GetUserDTO(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("%s:%w", op, err)
	}
	return userDTO.Uuid, nil
}

func (s *AuthService) InitUser(ctx context.Context, userID, refreshToken, ip, fingerprint string) error {
	const op = "services.InitUser"

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	refreshTokenExp := time.Now().Add(s.refreshTokenTTL)
	return s.postgres.InitUser(ctx,
		userIDInt,
		refreshToken,
		ip,
		fingerprint,
		refreshTokenExp)
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken, fingerprint string) (models.UserDTO, error) {
	const op = "services.RefreshToken"

	session, err := s.postgres.GetSession(ctx, refreshToken)
	if err != nil {
		return models.UserDTO{}, fmt.Errorf("%s:%w", op, err)
	}

	if session.ExpiredAt.Before(time.Now()) {
		return models.UserDTO{}, fmt.Errorf("%s:%w", op, ErrRefreshTokenExp)
	}

	if session.FingerPrint != fingerprint {
		// todo may be delete session ?
		return models.UserDTO{}, fmt.Errorf("%s:%w", op, ErrDifferentFingerPrint)
	}

	role, err := s.postgres.RoleUser(ctx, session.UserID)
	if err != nil {
		return models.UserDTO{}, fmt.Errorf("%s:%w", op, err)
	}
	// todo refresh token

	return models.UserDTO{
		Uuid: session.UserID,
		Role: role,
	}, nil
}
