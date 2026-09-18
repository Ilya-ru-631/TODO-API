package handler

import "todo_api/internal/models"

type CreateTaskRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
}

type UpdateTaskRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Done        bool    `json:"done"`
}

func (r CreateTaskRequest) ToTask() models.Task {
	return models.Task{
		Title:       r.Title,
		Description: r.Description,
	}
}

func (r UpdateTaskRequest) ToTask() models.Task {
	return models.Task{
		Title:       r.Title,
		Description: r.Description,
		Done:        r.Done,
	}
}
