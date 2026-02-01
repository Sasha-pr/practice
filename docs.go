// Package main Golang Test API
//
// # Документация для API сервиса объявлений
//
// BasePath: /
// Version: 1.0.0
//
// Consumes:
// - application/json
// - multipart/form-data
//
// Produces:
// - application/json
//
// SecurityDefinitions:
//
//	APIKey:
//	 type: apiKey
//	 name: Key
//	 in: header
//
// swagger:meta
package docs

import (
	_ "golang-test/internal/models"
)

// Общие структуры ответов
// swagger:response ErrorResponse
type ErrorResponse struct {
	// in: body
	Body struct {
		// Ошибка
		// Example: invalid request
		Error string `json:"error"`
	}
}

// swagger:response SuccessResponse
type SuccessResponse struct {
	// in: body
	Body struct {
		// Статус
		// Example: success
		Status string `json:"status"`
	}
}
