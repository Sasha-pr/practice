package repository

import (
	"context"
	"database/sql"
	"fmt"
	"golang-test/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.UserCreate) (*models.User, error) {
	query := `
		INSERT INTO users (name, email)
		VALUES ($1, $2)
		RETURNING id, created_at
	`

	var newUser models.User
	newUser.Name = user.Name
	newUser.Email = user.Email

	err := r.db.QueryRowContext(ctx, query, user.Name, user.Email).Scan(&newUser.ID, &newUser.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &newUser, nil
}

func (r *UserRepository) Delete(ctx context.Context, id int) error {
	// Проверяем существование пользователя
	var exists bool
	err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("user with id %d does not exist", id)
	}

	// Удаляем пользователя (объявления удалятся каскадно из-за FOREIGN KEY)
	_, err = r.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", id)
	return err
}
