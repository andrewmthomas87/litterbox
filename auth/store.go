package auth

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/mediocregopher/radix/v3"
)

type Store struct {
	db   *sqlx.DB
	pool *radix.Pool
}

func NewStore(db *sqlx.DB, pool *radix.Pool) *Store {
	return &Store{db: db, pool: pool}
}

func (s *Store) SetUser(ctx context.Context, id, email, name, accessToken, refreshToken string) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	exists, err := s.UserExists(id)
	if err != nil {
		_ = tx.Rollback()

		return err
	}

	if exists {
		_, err := tx.ExecContext(ctx, "UPDATE users SET email=?, name=? WHERE id=?", email, name, id)
		if err != nil {
			_ = tx.Rollback()

			return err
		}
	} else {
		_, err := tx.ExecContext(ctx, "INSERT INTO users (id, email, name) VALUES (?, ?, ?)", id, email, name)
		if err != nil {
			_ = tx.Rollback()

			return err
		}
	}

	cmd := radix.Cmd(nil, "HMSET", fmt.Sprintf("user:%s", id), "access_token", accessToken, "refresh_token", refreshToken)
	if err := s.pool.Do(cmd); err != nil {
		_ = tx.Rollback()

		return err
	}

	if err := tx.Commit(); err != nil {
		cmd := radix.Cmd(nil, "DEL", fmt.Sprintf("user:%s", id))
		_ = s.pool.Do(cmd)

		return err
	}

	return nil
}

func (s *Store) UserExists(id string) (bool, error) {
	var exists bool
	return exists, s.pool.Do(radix.Cmd(&exists, "EXISTS", fmt.Sprintf("user:%s", id)))
}
