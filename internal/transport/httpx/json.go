package httpx

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func JSONResponse(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

func ErrorJson(w http.ResponseWriter, status int, code int, message string) {
	JSONResponse(w, status, ErrorResponse{
		Code:    code,
		Message: message,
	})
}
