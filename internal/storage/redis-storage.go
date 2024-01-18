package storage

import "context"

type Redis interface {
	Ping(ctx context.Context) error
	CreateRegSession(ctx context.Context, fingerPrint, timeCode string) error
	GetRegSession(ctx context.Context, timeCode string) (string, error)
}
