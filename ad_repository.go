package repository

import (
	"context"
	"database/sql"
	"fmt"
	"golang-test/internal/models"
	"os"
	"path/filepath"
)

type AdRepository struct {
	DB *sql.DB
}

func NewAdRepository(db *sql.DB) *AdRepository {
	return &AdRepository{DB: db}
}

func (r *AdRepository) GetByID(ctx context.Context, id int) (*models.Ad, error) {
	query := `
		SELECT 
			a.id, a.title, a.description, a.price, a.image_filename, 
			a.is_enabled, a.created_at,
			u.id, u.name, u.email, u.created_at,
			c.id, c.name, c.extra_property
		FROM ads a
		JOIN users u ON a.user_id = u.id
		JOIN categories c ON a.category_id = c.id
		WHERE a.id = $1
	`

	var ad models.Ad
	var user models.User
	var category models.Category

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&ad.ID, &ad.Title, &ad.Description, &ad.Price, &ad.Image,
		&ad.IsEnabled, &ad.CreatedAt,
		&user.ID, &user.Name, &user.Email, &user.CreatedAt,
		&category.ID, &category.Name, &category.ExtraProperty,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	ad.User = user
	ad.Category = category

	return &ad, nil
}

func (r *AdRepository) GetAll(ctx context.Context) ([]models.Ad, error) {
	query := `
		SELECT 
			a.id, a.title, a.description, a.price, a.image_filename, 
			a.is_enabled, a.created_at,
			u.id, u.name, u.email, u.created_at,
			c.id, c.name, c.extra_property
		FROM ads a
		JOIN users u ON a.user_id = u.id
		JOIN categories c ON a.category_id = c.id
		ORDER BY a.created_at DESC
	`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ads []models.Ad

	for rows.Next() {
		var ad models.Ad
		var user models.User
		var category models.Category

		err := rows.Scan(
			&ad.ID, &ad.Title, &ad.Description, &ad.Price, &ad.Image,
			&ad.IsEnabled, &ad.CreatedAt,
			&user.ID, &user.Name, &user.Email, &user.CreatedAt,
			&category.ID, &category.Name, &category.ExtraProperty,
		)

		if err != nil {
			return nil, err
		}

		ad.User = user
		ad.Category = category
		ads = append(ads, ad)
	}

	return ads, nil
}

func (r *AdRepository) Create(ctx context.Context, ad *models.AdCreate, imageFilename string) (*models.Ad, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Проверяем существование пользователя
	var userExists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", ad.UserID).Scan(&userExists)
	if err != nil {
		return nil, err
	}
	if !userExists {
		return nil, fmt.Errorf("user with id %d does not exist", ad.UserID)
	}

	// Проверяем существование категории
	var categoryExists bool
	err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)", ad.CategoryID).Scan(&categoryExists)
	if err != nil {
		return nil, err
	}
	if !categoryExists {
		return nil, fmt.Errorf("category with id %d does not exist", ad.CategoryID)
	}

	query := `
		INSERT INTO ads (user_id, category_id, title, description, price, image_filename, is_enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	var createdAd models.Ad
	err = tx.QueryRowContext(ctx, query,
		ad.UserID, ad.CategoryID, ad.Title, ad.Description, ad.Price,
		imageFilename, true, // По умолчанию включено
	).Scan(&createdAd.ID, &createdAd.CreatedAt)

	if err != nil {
		return nil, err
	}

	// Получаем полные данные объявления
	fullQuery := `
		SELECT 
			a.title, a.description, a.price, a.image_filename, a.is_enabled,
			u.id, u.name, u.email, u.created_at,
			c.id, c.name, c.extra_property
		FROM ads a
		JOIN users u ON a.user_id = u.id
		JOIN categories c ON a.category_id = c.id
		WHERE a.id = $1
	`

	var user models.User
	var category models.Category

	err = tx.QueryRowContext(ctx, fullQuery, createdAd.ID).Scan(
		&createdAd.Title, &createdAd.Description, &createdAd.Price, &createdAd.Image,
		&createdAd.IsEnabled,
		&user.ID, &user.Name, &user.Email, &user.CreatedAt,
		&category.ID, &category.Name, &category.ExtraProperty,
	)

	if err != nil {
		return nil, err
	}

	createdAd.User = user
	createdAd.Category = category

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &createdAd, nil
}

func (r *AdRepository) Update(ctx context.Context, id int, update *models.AdUpdate, imageFilename string) error {
	// Сначала проверим существование объявления
	var exists bool
	err := r.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM ads WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("ad with id %d does not exist", id)
	}

	query := `
		UPDATE ads 
		SET title = $1, description = $2, price = $3
		WHERE id = $4
	`

	args := []interface{}{update.Title, update.Description, update.Price, id}

	// Если передано новое изображение, обновляем его
	if imageFilename != "" {
		// Получаем старое имя файла для удаления
		var oldImage string
		err := r.DB.QueryRowContext(ctx, "SELECT image_filename FROM ads WHERE id = $1", id).Scan(&oldImage)
		if err != nil {
			return err
		}

		// Удаляем старый файл
		if oldImage != "" {
			oldPath := filepath.Join("uploads", "images", oldImage)
			os.Remove(oldPath)
		}

		query = `
			UPDATE ads 
			SET title = $1, description = $2, price = $3, image_filename = $4
			WHERE id = $5
		`
		args = []interface{}{update.Title, update.Description, update.Price, imageFilename, id}
	}

	_, err = r.DB.ExecContext(ctx, query, args...)
	return err
}

func (r *AdRepository) Toggle(ctx context.Context, id int, enabled bool) error {
	query := "UPDATE ads SET is_enabled = $1 WHERE id = $2"
	_, err := r.DB.ExecContext(ctx, query, enabled, id)
	return err
}

func (r *AdRepository) Delete(ctx context.Context, id int) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Получаем имя файла изображения
	var imageFilename string
	err = tx.QueryRowContext(ctx, "SELECT image_filename FROM ads WHERE id = $1", id).Scan(&imageFilename)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("ad with id %d does not exist", id)
		}
		return err
	}

	// Удаляем объявление
	_, err = tx.ExecContext(ctx, "DELETE FROM ads WHERE id = $1", id)
	if err != nil {
		return err
	}

	// Удаляем файл изображения
	if imageFilename != "" {
		imagePath := filepath.Join("uploads", "images", imageFilename)
		os.Remove(imagePath)
	}

	return tx.Commit()
}
