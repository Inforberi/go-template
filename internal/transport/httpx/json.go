package httpx

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

type FieldError struct {
	Field string `json:"field"`
	Rule  string `json:"rule"`
	Param string `json:"param,omitempty"`
}

func JSONResponse(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

func ErrorJson(w http.ResponseWriter, status int, code int, message string, details ...FieldError) {
	JSONResponse(w, status, ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	})
}
