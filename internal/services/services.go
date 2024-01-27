package services

import (
	"context"

	"github.com/UdinSemen/moscow-events-backend/internal/domain/models"
	"github.com/UdinSemen/moscow-events-backend/internal/storage"
)

type Auth interface {
	CreateRegSession(ctx context.Context, fingerPrint string) (string, error)
	GetRegSession(ctx context.Context, fingerPrint, timeCode string) (string, error)
	InitUser(ctx context.Context, userID, refreshToken, ip, fingerprint string) error
	RefreshToken(ctx context.Context, refreshToken, fingerprint string) (models.UserDTO, error)
}

type Service struct {
	Auth
}

func NewService(redis storage.Redis, postgres storage.PgStorage) *Service {
	return &Service{
		Auth: NewAuth(redis, postgres),
	}
}
