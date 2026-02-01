package models

type Category struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	ExtraProperty string `json:"extra_property"`
}
