package services

import (
	"context"
	"time"

	"github.com/UdinSemen/moscow-events-backend/internal/domain/models"
	jwtmanager "github.com/UdinSemen/moscow-events-backend/internal/jwt-manager"
	"github.com/UdinSemen/moscow-events-backend/internal/storage"
)

type Auth interface {
	CreateRegSession(ctx context.Context, fingerPrint string) (string, error)
	GetRegSession(ctx context.Context, fingerPrint, timeCode string) (string, error)
	GetUserDTOByTg(ctx context.Context, userTgId string) (models.UserDTO, error)
	InitSession(ctx context.Context, userID, refreshToken, ip, fingerprint string) error
	GenerateTokens(ctx context.Context, userID, role string) (string, string, error)
	RefreshToken(ctx context.Context, refreshToken, fingerprint string) (string, string, error)
	RefreshSession(ctx context.Context, refreshTokenOld, refreshTokenNew, ip string) error
}

type Event interface {
	GetEvents(ctx context.Context, userID, category string, date []time.Time) ([]models.Event, error)
}

type Service struct {
	Auth
	Event
}

func NewService(redis storage.Redis,
	postgres storage.PgStorage,
	refreshTTL time.Duration,
	jwtManager jwtmanager.TokenManager) *Service {
	return &Service{
		Auth:  NewAuth(redis, postgres, refreshTTL, jwtManager),
		Event: NewEventService(postgres),
	}
}
