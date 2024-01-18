package storage

import (
	"context"
)

type PgStorage interface {
	Ping() error
	InitUser(ctx context.Context, userID int64, refreshToken string) error
}
