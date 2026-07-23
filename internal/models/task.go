package models

type Task struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Done        bool    `json:"done"`
}
