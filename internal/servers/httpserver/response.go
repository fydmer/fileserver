package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fydmer/fileserver/internal/domain/repository"
)

func getCodeFromPayloadError(err error) int {
	switch {
	case errors.Is(err, repository.ErrResourceNotFound):
		return http.StatusNotFound
	case errors.Is(err, repository.ErrBadRequest):
		return http.StatusBadRequest
	case errors.Is(err, repository.ErrResourceAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, repository.ErrUnknown):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func httpJson(w http.ResponseWriter, data map[string]any, code int) {
	body, _ := json.Marshal(data)
	w.Header().Del("Content-Length")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	_, _ = w.Write(body)
}

func httpError(w http.ResponseWriter, message string, code int) {
	httpJson(w, map[string]any{
		"status":  http.StatusText(code),
		"message": message,
	}, code)
}
