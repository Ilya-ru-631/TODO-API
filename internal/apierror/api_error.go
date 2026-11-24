package apierror

import (
	"encoding/json"
	"net/http"
)

const (
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeInvalidToken       = "INVALID_TOKEN"
	CodeInvalidID          = "INVALID_ID"
	CodeTaskNotFound       = "TASK_NOT_FOUND"
	CodeDuplicateTitle     = "DUPLICATE_TITLE"
	CodeInvalidJSON        = "INVALID_JSON"
	CodeValidation         = "VALIDATION_ERROR"
	CodeInternal           = "INTERNAL_ERROR"
	CodeUsernameTaken      = "USERNAME_TAKEN"
	CodeInvalidCredentials = "INVALID_CREDENTIALS"
)

type Response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func Write(w http.ResponseWriter, status int, code, message string) {
	write(w, status, Response{Code: code, Message: message})
}

func WriteValidation(w http.ResponseWriter, code, message, field string) {
	write(w, http.StatusBadRequest, Response{Code: code, Message: message, Field: field})
}

func write(w http.ResponseWriter, status int, body Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
