package handlers

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"golang-test/internal/models"
	"golang-test/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdHandler struct {
	repo *repository.AdRepository
}

func NewAdHandler(repo *repository.AdRepository) *AdHandler {
	return &AdHandler{repo: repo}
}

// GetAdByID получает объявление по ID
// @Summary Получить объявление по ID
// @Description Возвращает информацию об объявлении по его идентификатору
// @Tags ads
// @Accept json
// @Produce json
// @Param id path int true "ID объявления"
// @Security APIKey
// @Success 200 {object} models.Ad
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /ads/{id} [get]
func (h *AdHandler) GetAdByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid ad id",
		})
		return
	}

	ad, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		slog.Error("failed to get ad", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	if ad == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "ad not found",
		})
		return
	}

	c.JSON(http.StatusOK, ad)
}

// GetAllAds получает все объявления
// @Summary Получить все объявления
// @Description Возвращает список всех объявлений
// @Tags ads
// @Accept json
// @Produce json
// @Security APIKey
// @Success 200 {array} models.Ad
// @Failure 500 {object} ErrorResponse
// @Router /ads [get]
func (h *AdHandler) GetAllAds(c *gin.Context) {
	ads, err := h.repo.GetAll(c.Request.Context())
	if err != nil {
		slog.Error("failed to get ads", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, ads)
}

// CreateAd создает новое объявление
// @Summary Создать новое объявление
// @Description Создает новое объявление с изображением
// @Tags ads
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Изображение объявления (png, jpg, jpeg)"
// @Param user_id formData int true "ID пользователя"
// @Param category_id formData int true "ID категории"
// @Param title formData string true "Заголовок объявления"
// @Param description formData string true "Описание объявления"
// @Param price formData number true "Цена"
// @Security APIKey
// @Success 201 {object} models.Ad
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /ads [post]
func (h *AdHandler) CreateAd(c *gin.Context) {
	// 1. Получаем файл из запроса
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "image is required",
		})
		return
	}

	// 2. Проверяем расширение
	ext := filepath.Ext(file.Filename)
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "only png/jpg/jpeg allowed",
		})
		return
	}

	// 3. Генерируем уникальное имя
	filename := uuid.New().String() + ext

	// 4. Путь сохранения
	savePath := "uploads/images/" + filename

	// 5. Сохраняем файл физически
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		slog.Error("failed to save image", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "cannot save image",
		})
		return
	}

	// 6. Парсим остальные данные из формы
	userID, err := strconv.Atoi(c.PostForm("user_id"))
	if err != nil {
		os.Remove(savePath)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user_id",
		})
		return
	}

	categoryID, err := strconv.Atoi(c.PostForm("category_id"))
	if err != nil {
		os.Remove(savePath)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid category_id",
		})
		return
	}

	title := c.PostForm("title")
	if title == "" {
		os.Remove(savePath)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "title is required",
		})
		return
	}

	description := c.PostForm("description")
	if description == "" {
		os.Remove(savePath)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "description is required",
		})
		return
	}

	price, err := strconv.ParseFloat(c.PostForm("price"), 64)
	if err != nil || price < 0 {
		os.Remove(savePath)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid price",
		})
		return
	}

	// 7. Создаем объект для создания объявления
	adCreate := models.AdCreate{
		UserID:      userID,
		CategoryID:  categoryID,
		Title:       title,
		Description: description,
		Price:       price,
		Image:       filename,
	}

	// 8. Создаем объявление в БД
	ad, err := h.repo.Create(c.Request.Context(), &adCreate, filename)
	if err != nil {
		// Удаляем сохраненный файл, если не удалось создать запись в БД
		os.Remove(savePath)
		slog.Error("failed to create ad", "error", err)
		errorMsg := "failed to create ad"
		if err.Error() == fmt.Sprintf("user with id %d does not exist", userID) {
			errorMsg = "user does not exist"
		} else if err.Error() == fmt.Sprintf("category with id %d does not exist", categoryID) {
			errorMsg = "category does not exist"
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errorMsg,
		})
		return
	}

	c.JSON(http.StatusCreated, ad)
}

// UpdateAd обновляет объявление
// @Summary Обновить объявление
// @Description Обновляет информацию об объявлении
// @Tags ads
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "ID объявления"
// @Param image formData file false "Новое изображение (png, jpg, jpeg)"
// @Param title formData string true "Заголовок объявления"
// @Param description formData string true "Описание объявления"
// @Param price formData number true "Цена"
// @Security APIKey
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /ads/{id} [put]
func (h *AdHandler) UpdateAd(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid ad id",
		})
		return
	}

	// Парсим данные из формы
	title := c.PostForm("title")
	description := c.PostForm("description")
	priceStr := c.PostForm("price")

	if title == "" || description == "" || priceStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "title, description and price are required",
		})
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid price",
		})
		return
	}

	// Создаем объект обновления
	adUpdate := models.AdUpdate{
		Title:       title,
		Description: description,
		Price:       price,
	}

	var imageFilename string

	// Проверяем, передается ли новое изображение
	file, err := c.FormFile("image")
	if err == nil {
		// Если файл передан, обрабатываем его
		ext := filepath.Ext(file.Filename)
		if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "only png/jpg/jpeg allowed",
			})
			return
		}

		// Генерируем уникальное имя
		imageFilename = uuid.New().String() + ext
		savePath := "uploads/images/" + imageFilename

		// Сохраняем файл
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			slog.Error("failed to save image", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "cannot save image",
			})
			return
		}
	}

	// Обновляем объявление в БД
	err = h.repo.Update(c.Request.Context(), id, &adUpdate, imageFilename)
	if err != nil {
		// Удаляем сохраненный файл, если не удалось обновить запись
		if imageFilename != "" {
			os.Remove("uploads/images/" + imageFilename)
		}
		slog.Error("failed to update ad", "error", err, "id", id)
		if err.Error() == fmt.Sprintf("ad with id %d does not exist", id) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "ad not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to update ad",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

// ToggleAd выключает/включает объявление
// @Summary Переключить статус объявления
// @Description Включает или выключает объявление
// @Tags ads
// @Accept json
// @Produce json
// @Param id path int true "ID объявления"
// @Security APIKey
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /ads/{id}/toggle [patch]
func (h *AdHandler) ToggleAd(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid ad id",
		})
		return
	}

	// Получаем текущий статус объявления
	var currentStatus bool
	err = h.repo.DB.QueryRowContext(c.Request.Context(),
		"SELECT is_enabled FROM ads WHERE id = $1", id).Scan(&currentStatus)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "ad not found",
			})
			return
		}
		slog.Error("failed to get ad status", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	// Меняем статус на противоположный
	newStatus := !currentStatus
	err = h.repo.Toggle(c.Request.Context(), id, newStatus)
	if err != nil {
		slog.Error("failed to toggle ad", "error", err, "id", id)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to toggle ad",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"enabled": newStatus,
	})
}

// DeleteAd удаляет объявление
// @Summary Удалить объявление
// @Description Удаляет объявление по ID
// @Tags ads
// @Accept json
// @Produce json
// @Param id path int true "ID объявления"
// @Security APIKey
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /ads/{id} [delete]
func (h *AdHandler) DeleteAd(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid ad id",
		})
		return
	}

	err = h.repo.Delete(c.Request.Context(), id)
	if err != nil {
		slog.Error("failed to delete ad", "error", err, "id", id)
		if err.Error() == fmt.Sprintf("ad with id %d does not exist", id) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "ad not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete ad",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}
