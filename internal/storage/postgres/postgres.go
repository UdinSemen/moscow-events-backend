package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/UdinSemen/moscow-events-backend/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type PgStorage struct {
	db *sqlx.DB
}

func InitPgStorage(cfg *config.Config) (*PgStorage, error) {
	const op = "pg_storage.InitPgStorage"

	dbConf := cfg.Postgres

	connect, err := sqlx.Connect("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbConf.Host, dbConf.Port, dbConf.User, dbConf.Password, dbConf.DbName, dbConf.SslMode))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &PgStorage{
		db: connect,
	}, nil
}

func (s *PgStorage) Ping() error {
	return s.db.Ping()
}

func (s *PgStorage) InitUser(
	ctx context.Context,
	userID int64,
	refreshToken,
	ip,
	fingerprint string,
	expireAt time.Time) error {
	const op = "pg_storage.InitUser"

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	var uuid string

	row := tx.QueryRowContext(ctx, "select u.id from users u where u.tg_user_id = $1", userID)
	if err := row.Scan(&uuid); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.Exec("insert into sessions (user_id, refresh_token, ip, finger_print, exp_at) "+
		"values ($1, $2, $3, $4, $5)",
		uuid, refreshToken, ip, fingerprint, expireAt)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}

	return tx.Commit()
}
