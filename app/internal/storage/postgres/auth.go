package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/UdinSemen/moscow-events-backend/internal/domain/models"
	"golang.org/x/net/context"
)

const opPrefixPgStorageAuth = "pg_storage.auth."

func (s *PgStorage) InitSession(
	ctx context.Context,
	userTgID string,
	refreshToken,
	ip,
	fingerprint string,
	expireAt time.Time) error {
	const op = opPrefixPgStorageAuth + "InitUser"

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	var uuid string

	row := tx.QueryRowContext(ctx, "select u.id from users u where u.tg_user_id = $1", userTgID)
	if err := row.Scan(&uuid); err != nil {
		_ = tx.Rollback()
		outErr := fmt.Errorf("%s: %w", op, err)
		return outErr
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
	const op = opPrefixPgStorageAuth + "GetSession"

	var session models.Session
	query := "select s.user_id, s.finger_print, s.exp_at from sessions s where s.refresh_token=$1"
	if err := s.db.Get(&session, query, refreshToken); err != nil {
		outErr := fmt.Errorf("%s:%w", op, err)
		if errors.Is(err, sql.ErrNoRows) {
			outErr = fmt.Errorf("%s:%w", op, ErrNoRows)
		}
		return models.Session{}, outErr
	}

	return session, nil
}

type InputGetUserDTO struct {
	TgID int64
	UUID string
}

func (s *PgStorage) GetUserDTO(ctx context.Context, input InputGetUserDTO, typeId string) (models.UserDTO, error) {
	const op = opPrefixPgStorageAuth + "GetUserDTO"

	if !slices.Contains([]string{TypeTgID, TypeUUID}, typeId) {
		return models.UserDTO{}, ErrInvalidIDType
	}
	var model models.UserDTO
	var query string
	args := map[string]interface{}{
		"id": input.TgID,
	}
	switch typeId {
	case TypeTgID:
		query = "select u.id, Coalesce(u.role, '') as role from users u where u.tg_user_id=:id"
	case TypeUUID:
		query = "select u.id, Coalesce(u.role, '') as role from users u where u.id=:id"
		args["id"] = input.UUID
	}

	rows, err := s.db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return models.UserDTO{}, fmt.Errorf("%s:%w", op, err)
	}

	for rows.Next() {
		if err := rows.StructScan(&model); err != nil {
			return models.UserDTO{}, fmt.Errorf("%s:%w", op, err)
		}
	}

	return model, nil
}

func (s *PgStorage) RefreshSession(ctx context.Context, refreshTokenOld, refreshTokenNew, ip string, expireAt time.Time) error {
	const op = opPrefixPgStorageAuth + "RefreshSession"

	query := "update sessions set refresh_token=:refresh_token_new, " +
		"ip=:ip, exp_at=:exp_at where refresh_token=:refresh_token_old "
	_, err := s.db.NamedExecContext(ctx, query, map[string]interface{}{
		"refresh_token_new": refreshTokenNew,
		"ip":                ip,
		"exp_at":            expireAt,
		"refresh_token_old": refreshTokenOld,
	})

	return err
}
