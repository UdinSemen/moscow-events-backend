package storage

import (
	"context"
	"time"
)

type PgStorage interface {
	Ping() error
	InitUser(ctx context.Context,
		userID int64,
		refreshToken,
		ip,
		fingerprint string,
		expireAt time.Time) error
}
