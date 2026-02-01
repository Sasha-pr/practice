package models

import "time"

type Ad struct {
	ID          int       `json:"id"`
	User        User      `json:"user"`
	Category    Category  `json:"category"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Image       string    `json:"image"`
	IsEnabled   bool      `json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at"`
}
