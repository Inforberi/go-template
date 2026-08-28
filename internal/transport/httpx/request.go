package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

const maxRequestBodyBytes int64 = 1 << 20

var requestValidator = newRequestValidator()

func DecodeAndValidateJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		ErrorJson(w, http.StatusUnsupportedMediaType, http.StatusUnsupportedMediaType, "content type must be application/json")
		return false
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		writeDecodeError(w, err)
		return false
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		writeDecodeError(w, err)
		return false
	}

	if err := requestValidator.Struct(dst); err != nil {
		var validationErrors validator.ValidationErrors
		if !errors.As(err, &validationErrors) {
			ErrorJson(w, http.StatusInternalServerError, http.StatusInternalServerError, "request validation failed")
			return false
		}

		details := make([]FieldError, 0, len(validationErrors))
		for _, validationError := range validationErrors {
			field := validationError.Namespace()
			if _, suffix, ok := strings.Cut(field, "."); ok {
				field = suffix
			}

			details = append(details, FieldError{
				Field: field,
				Rule:  validationError.Tag(),
				Param: validationError.Param(),
			})
		}

		ErrorJson(w, http.StatusUnprocessableEntity, http.StatusUnprocessableEntity, "request validation failed", details...)
		return false
	}

	return true
}

func newRequestValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return validate
}

func writeDecodeError(w http.ResponseWriter, err error) {
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		ErrorJson(w, http.StatusRequestEntityTooLarge, http.StatusRequestEntityTooLarge, "request body is too large")
		return
	}

	ErrorJson(w, http.StatusBadRequest, http.StatusBadRequest, "invalid request body")
}
