package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"lamarankerja/internal/model"
)

var ErrLoginSalah = errors.New("username atau password salah")

func (s *Store) UserByUsername(ctx context.Context, username string) (model.User, error) {
	var u model.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, password_hash FROM users WHERE username = $1`, username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, ErrLoginSalah
	}
	return u, err
}

func (s *Store) CekPassword(ctx context.Context, username, password string) (model.User, error) {
	u, err := s.UserByUsername(ctx, username)
	if err != nil {
		return u, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return model.User{}, ErrLoginSalah
	}
	return u, nil
}

// SimpanAdmin membuat akun baru atau menimpa password akun yang sudah ada.
func (s *Store) SimpanAdmin(ctx context.Context, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO users (username, password_hash) VALUES ($1, $2)
		ON CONFLICT (username) DO UPDATE SET password_hash = EXCLUDED.password_hash`,
		username, string(hash))
	return err
}

func (s *Store) JumlahUser(ctx context.Context) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}
