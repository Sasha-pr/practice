package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"golang-test/internal/models"
	"golang-test/internal/repository"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	repo *repository.UserRepository
}

func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

// CreateUser создает нового пользователя
// @Summary Создать пользователя
// @Description Создает нового пользователя в системе
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.UserCreate true "Данные пользователя"
// @Security APIKey
// @Success 201 {object} models.User
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var userCreate models.UserCreate
	if err := c.ShouldBindJSON(&userCreate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.repo.Create(c.Request.Context(), &userCreate)
	if err != nil {
		slog.Error("failed to create user", "error", err)
		// Проверяем на дубликат email
		if err.Error() == "pq: duplicate key value violates unique constraint \"users_email_key\"" {
			c.JSON(http.StatusConflict, gin.H{
				"error": "user with this email already exists",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create user",
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// DeleteUser удаляет пользователя
// @Summary Удалить пользователя
// @Description Удаляет пользователя по ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Security APIKey
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user id",
		})
		return
	}

	err = h.repo.Delete(c.Request.Context(), id)
	if err != nil {
		slog.Error("failed to delete user", "error", err, "id", id)
		if err.Error() == fmt.Sprintf("user with id %d does not exist", id) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "user not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to delete user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}
