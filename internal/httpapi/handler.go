package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/richardnogueira01/feature-flag-mvp/internal/control"
)

type Service interface {
	Create(string, bool) (control.Flag, error)
	Update(string, bool) (control.Flag, error)
	Delete(string) error
	Get(string) (control.Flag, error)
	List() []control.Flag
}

type Handler struct {
	service Service
}

func NewHandler(service Service) http.Handler {
	return &Handler{service: service}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/v1/flags") {
		http.NotFound(w, r)
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, "/v1/flags")
	if rest == "" || rest == "/" {
		h.collection(w, r)
		return
	}
	key := strings.TrimPrefix(rest, "/")
	if strings.Contains(key, "/") || key == "" {
		http.NotFound(w, r)
		return
	}
	h.item(w, r, key)
}

func (h *Handler) collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var input struct {
			Key     string
			Enabled bool
		}
		if !decodeJSON(w, r, &input) {
			return
		}
		flag, err := h.service.Create(input.Key, input.Enabled)
		if err != nil {
			h.writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, flag)
	case http.MethodGet:
		writeJSON(w, http.StatusOK, map[string]any{"flags": h.service.List()})
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) item(w http.ResponseWriter, r *http.Request, key string) {
	switch r.Method {
	case http.MethodGet:
		flag, err := h.service.Get(key)
		if err != nil {
			h.writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, flag)
	case http.MethodPut:
		var input struct{ Enabled bool }
		if !decodeJSON(w, r, &input) {
			return
		}
		flag, err := h.service.Update(key, input.Enabled)
		if err != nil {
			h.writeServiceError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, flag)
	case http.MethodDelete:
		if err := h.service.Delete(key); err != nil {
			h.writeServiceError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return false
	}
	return true
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, control.ErrInvalidKey):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, control.ErrExists):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, control.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
