package models

type AdCreate struct {
	UserID      int     `json:"user_id" binding:"required"`
	CategoryID  int     `json:"category_id" binding:"required"`
	Title       string  `json:"title" binding:"required,min=1,max=200"`
	Description string  `json:"description" binding:"required,min=1"`
	Price       float64 `json:"price" binding:"required,min=0"`
	Image       string  `json:"image"`
}
