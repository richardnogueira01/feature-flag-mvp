package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/richardnogueira01/feature-flag-mvp/internal/control"
)

// ApplicationsHandler serves the /v1/applications/** hierarchy (task 013): applications,
// environments, configurations and the flags that live inside a configuration.
type ApplicationsHandler struct{ service *control.ScopeService }

func NewApplicationsHandler(service *control.ScopeService) http.Handler {
	return &ApplicationsHandler{service: service}
}

func (h *ApplicationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeError(w, http.StatusNotImplemented, "the application/environment/configuration hierarchy requires PostgreSQL (DATABASE_URL)")
		return
	}
	if !strings.HasPrefix(r.URL.Path, "/v1/applications") {
		http.NotFound(w, r)
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, "/v1/applications")
	rest = strings.TrimPrefix(rest, "/")
	var segments []string
	if rest != "" {
		segments = strings.Split(rest, "/")
	}
	switch len(segments) {
	case 0:
		h.applications(w, r)
	case 2:
		if segments[1] != "environments" {
			http.NotFound(w, r)
			return
		}
		h.environments(w, r, segments[0])
	case 4:
		if segments[1] != "environments" || segments[3] != "configurations" {
			http.NotFound(w, r)
			return
		}
		h.configurations(w, r, segments[0], segments[2])
	case 5:
		if segments[1] != "environments" || segments[3] != "configurations" {
			http.NotFound(w, r)
			return
		}
		h.configuration(w, r, segments[0], segments[2], segments[4])
	case 6:
		if segments[1] != "environments" || segments[3] != "configurations" || segments[5] != "flags" {
			http.NotFound(w, r)
			return
		}
		h.flags(w, r, segments[0], segments[2], segments[4])
	case 7:
		if segments[1] != "environments" || segments[3] != "configurations" || segments[5] != "flags" {
			http.NotFound(w, r)
			return
		}
		h.flag(w, r, segments[0], segments[2], segments[4], segments[6])
	default:
		http.NotFound(w, r)
	}
}

func (h *ApplicationsHandler) applications(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var in struct {
			Name string `json:"name"`
		}
		if !decodeJSON(w, r, &in) {
			return
		}
		app, err := h.service.CreateApplication(r.Context(), in.Name)
		if err != nil {
			h.writeScopeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, app)
	case http.MethodGet:
		apps, err := h.service.ListApplications(r.Context())
		if err != nil {
			h.writeScopeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"applications": apps})
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *ApplicationsHandler) environments(w http.ResponseWriter, r *http.Request, application string) {
	switch r.Method {
	case http.MethodPost:
		var in struct {
			Name string `json:"name"`
		}
		if !decodeJSON(w, r, &in) {
			return
		}
		env, err := h.service.CreateEnvironment(r.Context(), application, in.Name)
		if err != nil {
			h.writeScopeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, env)
	case http.MethodGet:
		envs, err := h.service.ListEnvironments(r.Context(), application)
		if err != nil {
			h.writeScopeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"environments": envs})
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *ApplicationsHandler) configurations(w http.ResponseWriter, r *http.Request, application, environment string) {
	switch r.Method {
	case http.MethodPost:
		var in struct {
			Name string `json:"name"`
		}
		if !decodeJSON(w, r, &in) {
			return
		}
		cfg, err := h.service.CreateConfiguration(r.Context(), application, environment, in.Name)
		if err != nil {
			h.writeScopeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, cfg)
	case http.MethodGet:
		cfgs, err := h.service.ListConfigurations(r.Context(), application, environment)
		if err != nil {
			h.writeScopeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"configurations": cfgs})
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *ApplicationsHandler) configuration(w http.ResponseWriter, r *http.Request, application, environment, configuration string) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	snapshot, err := h.service.GetConfiguration(r.Context(), application, environment, configuration)
	if err != nil {
		h.writeScopeError(w, err)
		return
	}
	etag := fmt.Sprintf("\"%d\"", snapshot.Revision)
	w.Header().Set("ETag", etag)
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (h *ApplicationsHandler) flags(w http.ResponseWriter, r *http.Request, application, environment, configuration string) {
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
		value, ok := h.resolveValue(w, in.Enabled, in.Value)
		if !ok {
			return
		}
		flag, err := h.service.CreateFlag(r.Context(), application, environment, configuration, in.Key, value)
		if err != nil {
			h.writeScopeError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, flag)
	case http.MethodGet:
		snapshot, err := h.service.GetConfiguration(r.Context(), application, environment, configuration)
		if err != nil {
			h.writeScopeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"flags": snapshot.Flags})
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *ApplicationsHandler) flag(w http.ResponseWriter, r *http.Request, application, environment, configuration, key string) {
	switch r.Method {
	case http.MethodGet:
		flag, err := h.service.GetFlag(r.Context(), application, environment, configuration, key)
		if err != nil {
			h.writeScopeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, flag)
	case http.MethodPut:
		var in struct {
			Enabled *bool           `json:"enabled"`
			Value   json.RawMessage `json:"value"`
		}
		if !decodeJSON(w, r, &in) {
			return
		}
		value, ok := h.resolveValue(w, in.Enabled, in.Value)
		if !ok {
			return
		}
		flag, err := h.service.UpdateFlag(r.Context(), application, environment, configuration, key, value)
		if err != nil {
			h.writeScopeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, flag)
	case http.MethodPatch:
		var in struct {
			Enabled *bool `json:"enabled"`
		}
		if !decodeJSON(w, r, &in) || in.Enabled == nil {
			writeError(w, http.StatusBadRequest, "enabled is required")
			return
		}
		flag, err := h.service.SetFlagEnabled(r.Context(), application, environment, configuration, key, *in.Enabled)
		if err != nil {
			h.writeScopeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, flag)
	case http.MethodDelete:
		if err := h.service.DeleteFlag(r.Context(), application, environment, configuration, key); err != nil {
			h.writeScopeError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.Header().Set("Allow", "GET, PUT, PATCH, DELETE")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// resolveValue mirrors the legacy handler's "exactly one of value or enabled" contract.
func (h *ApplicationsHandler) resolveValue(w http.ResponseWriter, enabled *bool, value json.RawMessage) (json.RawMessage, bool) {
	if len(value) > 0 && enabled != nil {
		writeError(w, http.StatusBadRequest, "provide exactly one of value or enabled")
		return nil, false
	}
	if len(value) > 0 {
		return value, true
	}
	if enabled == nil {
		writeError(w, http.StatusBadRequest, "value or enabled is required")
		return nil, false
	}
	v, _ := json.Marshal(*enabled)
	return v, true
}

func (h *ApplicationsHandler) writeScopeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, control.ErrScopeInvalidName), errors.Is(err, control.ErrInvalidKey), errors.Is(err, control.ErrInvalidValue):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, control.ErrScopeExists):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, control.ErrScopeNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
