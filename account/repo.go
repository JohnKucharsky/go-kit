package account

import (
	"context"
	"database/sql"
	"errors"
	"github.com/go-kit/kit/log"
)

type repo struct {
	db     *sql.DB
	logger log.Logger
}

func NewRepo(db *sql.DB, logger log.Logger) Repository {
	return &repo{
		db:     db,
		logger: log.With(logger, "repo", "sql"),
	}
}

func (r repo) CreateUser(ctx context.Context, user User) error {
	s := `INSERT INTO users (id,email,password) VALUES ($1, $2, $3)`

	_, err := r.db.ExecContext(ctx, s, user.ID, user.Email, user.Password)
	if err != nil {
		return err
	}

	return nil
}

func (r repo) GetUser(ctx context.Context, id string) (string, error) {
	var email string
	err := r.db.QueryRow("SELECT email FROM users WHERE id = $1", id).Scan(&email)

	if err != nil {
		return "", errors.New("error getting user from db")
	}

	return email, nil
}
