package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"github.com/UdinSemen/moscow-events-backend/internal/domain/models"
	jwtmanager "github.com/UdinSemen/moscow-events-backend/internal/jwt-manager"
	"github.com/UdinSemen/moscow-events-backend/internal/storage"
	storagePg "github.com/UdinSemen/moscow-events-backend/internal/storage/postgres"
	"github.com/UdinSemen/moscow-events-backend/pkg/utils"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	timeCodeLen      = 6
	opAuthServPrefix = "service."
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
	jwtManager      jwtmanager.TokenManager
}

func NewAuth(redis storage.Redis, postgres storage.PgStorage, refreshTTL time.Duration, jwtManager jwtmanager.TokenManager) *AuthService {
	zap.S().Infow("tokenTTL",
		"refresh", refreshTTL)
	return &AuthService{
		redis:           redis,
		postgres:        postgres,
		refreshTokenTTL: refreshTTL,
		jwtManager:      jwtManager,
	}
}

func (s *AuthService) CreateRegSession(ctx context.Context, fingerPrint string) (string, error) {
	timeCode := utils.GenTimeCode(timeCodeLen)
	return timeCode, s.redis.CreateRegSession(ctx, fingerPrint, timeCode)
}

func (s *AuthService) GetRegSession(ctx context.Context, fingerPrint, timeCode string) (string, error) {
	const op = opAuthServPrefix + "GetRegSession"

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

	return userTgId, nil
}

func (s *AuthService) InitSession(ctx context.Context, userID, refreshToken, ip, fingerprint string) error {
	const op = opAuthServPrefix + "InitUser"

	refreshTokenExp := time.Now().Add(s.refreshTokenTTL)
	return s.postgres.InitSession(ctx,
		userID,
		refreshToken,
		ip,
		fingerprint,
		refreshTokenExp)
}

func (s *AuthService) GetUserDTOByTg(ctx context.Context, userTgId string) (models.UserDTO, error) {
	const op = opAuthServPrefix + "GetUserDTO"

	id, err := strconv.ParseInt(userTgId, 10, 64)
	if err != nil {
		return models.UserDTO{}, fmt.Errorf("%s:%w", op, err)
	}
	return s.postgres.GetUserDTO(ctx, storagePg.InputGetUserDTO{
		TgID: id,
	}, storagePg.TypeTgID)
}

func (s *AuthService) GenerateTokens(_ context.Context, userID, role string) (string, string, error) {
	const op = opAuthServPrefix + "GenerateTokens"

	accessToken, err := s.jwtManager.GenerateToken(userID, role)
	if err != nil {
		return "", "", fmt.Errorf("%s:%w", op, err)
	}

	refreshToken, err := s.jwtManager.NewRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("%s:%w", op, err)
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken, fingerprint string) (string, string, error) {
	const op = opAuthServPrefix + "RefreshToken"

	session, err := s.postgres.GetSession(ctx, refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("%s:%w", op, err)
	}

	zap.S().Debug(session)
	zap.S().Debug(session.ExpiredAt)
	zap.S().Debug(session.ExpiredAt.Before(time.Now()))
	zap.S().Debug(time.Now())
	if session.ExpiredAt.Before(time.Now()) {
		return "", "", fmt.Errorf("%s:%w", op, ErrRefreshTokenExp)
	}

	if session.FingerPrint != fingerprint {
		// todo may be delete session ?
		return "", "", fmt.Errorf("%s:%w", op, ErrDifferentFingerPrint)
	}

	userDTO, err := s.postgres.GetUserDTO(ctx, storagePg.InputGetUserDTO{
		UUID: session.UserID,
	}, storagePg.TypeUUID)

	if err != nil {
		return "", "", fmt.Errorf("%s:%w", op, err)
	}

	return s.GenerateTokens(ctx, userDTO.Uuid, userDTO.Role)
}

func (s *AuthService) RefreshSession(ctx context.Context, refreshTokenOld, refreshTokenNew, ip string) error {
	const op = opAuthServPrefix + "RefreshSession"

	expireAt := time.Now().Add(s.refreshTokenTTL)
	return s.postgres.RefreshSession(ctx, refreshTokenOld, refreshTokenNew, ip, expireAt)
}
