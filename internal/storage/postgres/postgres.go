package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/UdinSemen/moscow-events-backend/internal/config"
	"github.com/UdinSemen/moscow-events-backend/internal/domain/models"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const opPrefixPgStorage = "pg_storage."

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

func (s *PgStorage) InitUser(
	ctx context.Context,
	userID int64,
	refreshToken,
	ip,
	fingerprint string,
	expireAt time.Time) error {
	const op = opPrefixPgStorage + "InitUser"

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

func (s *PgStorage) GetSession(ctx context.Context, refreshToken string) (models.Session, error) {
	const op = opPrefixPgStorage + "GetSession"

	var session models.Session
	query := "select s.user_id, s.finger_print, s.exp_at from sessions s where s.refresh_token=$1"
	if err := s.db.Get(&session, query, refreshToken); err != nil {
		return models.Session{}, fmt.Errorf("%s:%w", op, err)
	}

	return session, nil
}

func (s *PgStorage) RoleUser(ctx context.Context, uuid string) (string, error) {
	const op = opPrefixPgStorage + "RoleUser"

	var role string
	query := "select u.role from users u where u.id=$1"
	if err := s.db.Get(&role, query, uuid); err != nil {
		return "", fmt.Errorf("%s:%w", op, err)
	}

	return role, nil
}

func (s *PgStorage) GetUserDTO(ctx context.Context, tgId int64) (models.UserDTO, error) {
	const op = opPrefixPgStorage + "GetUserDTO"

	var model models.UserDTO
	query := "select u.id, u.role from users u where u.tg_user_id=$1"
	if err := s.db.Get(&model, query, tgId); err != nil {
		return models.UserDTO{}, fmt.Errorf("%s:%w", op, err)
	}

	return model, nil
}
