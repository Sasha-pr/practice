package models

type UserCreate struct {
	Name  string `json:"name" binding:"required,min=1,max=100"`
	Email string `json:"email" binding:"required,email,max=150"`
}
