package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/yourname/stockbit-appsec/internal/domain"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3) RETURNING id`
	err := r.DB.QueryRowContext(ctx, query, user.Email, user.PasswordHash, user.Role).Scan(&user.ID)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, role FROM users WHERE email = $1`
	row := r.DB.QueryRowContext(ctx, query, email)

	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // User not found
		}
		return nil, err
	}
	return &user, nil
}
