package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type requestFixture struct {
	Email string `json:"email" validate:"required,email"`
}

func TestDecodeAndValidateJSON(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		wantOK      bool
		wantStatus  int
	}{
		{name: "valid", contentType: "application/json; charset=utf-8", body: `{"email":"a@example.com"}`, wantOK: true},
		{name: "content type", contentType: "text/plain", body: `{}`, wantStatus: http.StatusUnsupportedMediaType},
		{name: "unknown field", contentType: "application/json", body: `{"email":"a@example.com","admin":true}`, wantStatus: http.StatusBadRequest},
		{name: "second JSON value", contentType: "application/json", body: `{} {}`, wantStatus: http.StatusBadRequest},
		{name: "validation", contentType: "application/json", body: `{"email":"invalid"}`, wantStatus: http.StatusUnprocessableEntity},
		{
			name:        "body limit",
			contentType: "application/json",
			body:        `{"email":"` + strings.Repeat("a", int(maxRequestBodyBytes)) + `"}`,
			wantStatus:  http.StatusRequestEntityTooLarge,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			request.Header.Set("Content-Type", test.contentType)
			var destination requestFixture

			got := DecodeAndValidateJSON(recorder, request, &destination)
			if got != test.wantOK {
				t.Fatalf("DecodeAndValidateJSON() = %v, want %v", got, test.wantOK)
			}
			if !got && recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
		})
	}
}
