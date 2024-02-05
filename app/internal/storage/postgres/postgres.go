package storage

import (
	"errors"
	"fmt"

	"github.com/UdinSemen/moscow-events-backend/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	opPrefixPgStorage = "pg_storage."
	TypeTgID          = "tg_user_id"
	TypeUUID          = "uuid_id"
)

var (
	ErrInvalidIDType = errors.New("invalid id type")
	ErrNoRows        = errors.New("no rows")
)

type PgStorage struct {
	db *sqlx.DB
}

func InitPgStorage(cfg *config.Config) (*PgStorage, error) {
	const op = opPrefixPgStorage + "InitPgStorage"

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
