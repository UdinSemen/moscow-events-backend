package storage

import (
	"context"
	"time"

	"github.com/UdinSemen/moscow-events-backend/internal/domain/models"
	storage "github.com/UdinSemen/moscow-events-backend/internal/storage/postgres"
)

type PgStorage interface {
	Ping() error
	InitSession(ctx context.Context,
		userTgID string,
		refreshToken,
		ip,
		fingerprint string,
		expireAt time.Time) error
	GetSession(ctx context.Context, refreshToken string) (models.Session, error)
	GetUserDTO(ctx context.Context, input storage.InputGetUserDTO, typeId string) (models.UserDTO, error)
	RefreshSession(ctx context.Context, refreshTokenOld, refreshTokenNew, ip string, expireAt time.Time) error
	GetEvents(ctx context.Context, userID, category string, date []time.Time) ([]models.Event, error)
}
