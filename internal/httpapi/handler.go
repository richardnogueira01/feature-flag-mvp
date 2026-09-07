package httpapi

import (
	"encoding/json"
	"errors"
	"github.com/richardnogueira01/feature-flag-mvp/internal/control"
	"net/http"
	"strings"
)

type Service interface {
	Create(string, bool) (control.Flag, error)
	Update(string, bool) (control.Flag, error)
	Delete(string) error
	Get(string) (control.Flag, error)
	List() []control.Flag
}
type FlexibleService interface {
	CreateValue(string, json.RawMessage) (control.Flag, error)
	UpdateValue(string, json.RawMessage) (control.Flag, error)
}
type Handler struct{ service Service }

func NewHandler(service Service) http.Handler { return &Handler{service: service} }
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
		var in struct {
			Key     string          `json:"key"`
			Enabled *bool           `json:"enabled"`
			Value   json.RawMessage `json:"value"`
		}
		if !decodeJSON(w, r, &in) {
			return
		}
		var f control.Flag
		var err error
		if len(in.Value) > 0 && in.Enabled != nil {
			writeError(w, 400, "provide exactly one of value or enabled")
			return
		}
		if fs, ok := h.service.(FlexibleService); ok && len(in.Value) > 0 {
			f, err = fs.CreateValue(in.Key, in.Value)
		} else {
			enabled := false
			if in.Enabled == nil {
				writeError(w, 400, "value or enabled is required")
				return
			}
			enabled = *in.Enabled
			f, err = h.service.Create(in.Key, enabled)
		}
		if err != nil {
			h.writeServiceError(w, err)
			return
		}
		writeJSON(w, 201, f)
	case http.MethodGet:
		writeJSON(w, 200, map[string]any{"flags": h.service.List()})
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, 405, "method not allowed")
	}
}
func (h *Handler) patch(w http.ResponseWriter, r *http.Request, key string) {
	var input struct {
		Enabled *bool `json:"enabled"`
	}
	if !decodeJSON(w, r, &input) || input.Enabled == nil {
		writeError(w, http.StatusBadRequest, "enabled is required")
		return
	}
	toggler, ok := h.service.(interface {
		SetEnabled(string, bool) (control.Flag, error)
	})
	if !ok {
		writeError(w, http.StatusNotImplemented, "toggle is not supported")
		return
	}
	flag, err := toggler.SetEnabled(key, *input.Enabled)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, flag)
}
func (h *Handler) item(w http.ResponseWriter, r *http.Request, key string) {
	switch r.Method {
	case http.MethodPatch:
		h.patch(w, r, key)
		return
	case http.MethodGet:
		f, e := h.service.Get(key)
		if e != nil {
			h.writeServiceError(w, e)
			return
		}
		writeJSON(w, 200, f)
	case http.MethodPut:
		var in struct {
			Enabled *bool           `json:"enabled"`
			Value   json.RawMessage `json:"value"`
		}
		if !decodeJSON(w, r, &in) {
			return
		}
		if len(in.Value) > 0 && in.Enabled != nil {
			writeError(w, 400, "provide exactly one of value or enabled")
			return
		}
		var f control.Flag
		var e error
		if fs, ok := h.service.(FlexibleService); ok && len(in.Value) > 0 {
			f, e = fs.UpdateValue(key, in.Value)
		} else {
			if in.Enabled == nil {
				writeError(w, 400, "value or enabled is required")
				return
			}
			f, e = h.service.Update(key, *in.Enabled)
		}
		if e != nil {
			h.writeServiceError(w, e)
			return
		}
		writeJSON(w, 200, f)
	case http.MethodDelete:
		if e := h.service.Delete(key); e != nil {
			h.writeServiceError(w, e)
			return
		}
		w.WriteHeader(204)
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		writeError(w, 405, "method not allowed")
	}
}
func decodeJSON(w http.ResponseWriter, r *http.Request, t any) bool {
	if json.NewDecoder(r.Body).Decode(t) != nil {
		writeError(w, 400, "invalid JSON")
		return false
	}
	return true
}
func (h *Handler) writeServiceError(w http.ResponseWriter, e error) {
	switch {
	case errors.Is(e, control.ErrInvalidKey), errors.Is(e, control.ErrInvalidValue):
		writeError(w, 400, e.Error())
	case errors.Is(e, control.ErrExists):
		writeError(w, 409, e.Error())
	case errors.Is(e, control.ErrNotFound):
		writeError(w, 404, e.Error())
	default:
		writeError(w, 500, "internal server error")
	}
}
func writeJSON(w http.ResponseWriter, s int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(s)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, s int, m string) {
	writeJSON(w, s, map[string]string{"error": m})
}
