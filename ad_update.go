package models

type AdUpdate struct {
	Title       string  `json:"title" binding:"required,min=1,max=200"`
	Description string  `json:"description" binding:"required,min=1"`
	Price       float64 `json:"price" binding:"required,min=0"`
	Image       string  `json:"image"`
}
