package storage

import (
	"context"
	"time"

	"github.com/UdinSemen/moscow-events-backend/internal/domain/models"
)

type PgStorage interface {
	Ping() error
	InitUser(ctx context.Context,
		userID int64,
		refreshToken,
		ip,
		fingerprint string,
		expireAt time.Time) error
	GetSession(ctx context.Context, refreshToken string) (models.Session, error)
	RoleUser(ctx context.Context, uuid string) (string, error)
	GetUserDTO(ctx context.Context, tgId int64) (models.UserDTO, error)
}
