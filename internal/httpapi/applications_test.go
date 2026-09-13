package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/richardnogueira01/feature-flag-mvp/internal/control"
)

// scopeFake is a minimal control.ScopeRepository used only to exercise the HTTP layer end to
// end (routing, status codes, ETag) without a real PostgreSQL instance. The scoping rules
// themselves are covered in depth by internal/control's own scope_test.go.
type scopeFake struct {
	apps map[string]bool
	envs map[string]bool                   // "application/environment"
	cfgs map[string]*control.Configuration // "application/environment/configuration"
	flgs map[string]map[string]control.ScopedFlag
}

func newScopeFake() *scopeFake {
	return &scopeFake{
		apps: map[string]bool{},
		envs: map[string]bool{},
		cfgs: map[string]*control.Configuration{},
		flgs: map[string]map[string]control.ScopedFlag{},
	}
}
func (f *scopeFake) CreateApplication(_ context.Context, name string) (control.Application, error) {
	if f.apps[name] {
		return control.Application{}, control.ErrScopeExists
	}
	f.apps[name] = true
	return control.Application{Name: name}, nil
}
func (f *scopeFake) ListApplications(_ context.Context) ([]control.Application, error) {
	var out []control.Application
	for a := range f.apps {
		out = append(out, control.Application{Name: a})
	}
	return out, nil
}
func (f *scopeFake) CreateEnvironment(_ context.Context, application, name string) (control.Environment, error) {
	if !f.apps[application] {
		return control.Environment{}, control.ErrScopeNotFound
	}
	key := application + "/" + name
	if f.envs[key] {
		return control.Environment{}, control.ErrScopeExists
	}
	f.envs[key] = true
	return control.Environment{Name: name}, nil
}
func (f *scopeFake) ListEnvironments(_ context.Context, application string) ([]control.Environment, error) {
	if !f.apps[application] {
		return nil, control.ErrScopeNotFound
	}
	return nil, nil
}
func (f *scopeFake) CreateConfiguration(_ context.Context, application, environment, name string) (control.Configuration, error) {
	if !f.envs[application+"/"+environment] {
		return control.Configuration{}, control.ErrScopeNotFound
	}
	key := application + "/" + environment + "/" + name
	if _, ok := f.cfgs[key]; ok {
		return control.Configuration{}, control.ErrScopeExists
	}
	cfg := &control.Configuration{Name: name}
	f.cfgs[key] = cfg
	return *cfg, nil
}
func (f *scopeFake) ListConfigurations(_ context.Context, application, environment string) ([]control.Configuration, error) {
	if !f.envs[application+"/"+environment] {
		return nil, control.ErrScopeNotFound
	}
	return nil, nil
}
func (f *scopeFake) resolve(application, environment, configuration string) (*control.Configuration, error) {
	cfg, ok := f.cfgs[application+"/"+environment+"/"+configuration]
	if !ok {
		return nil, control.ErrScopeNotFound
	}
	return cfg, nil
}
func (f *scopeFake) GetConfigurationSnapshot(_ context.Context, application, environment, configuration string) (control.ConfigurationSnapshot, error) {
	cfg, err := f.resolve(application, environment, configuration)
	if err != nil {
		return control.ConfigurationSnapshot{}, err
	}
	var flags []control.ScopedFlag
	for _, fl := range f.flgs[application+"/"+environment+"/"+configuration] {
		flags = append(flags, fl)
	}
	return control.ConfigurationSnapshot{Application: application, Environment: environment, Configuration: configuration, Revision: cfg.Revision, Flags: flags}, nil
}
func (f *scopeFake) GetScopedFlag(_ context.Context, application, environment, configuration, key string) (control.ScopedFlag, error) {
	if _, err := f.resolve(application, environment, configuration); err != nil {
		return control.ScopedFlag{}, err
	}
	flag, ok := f.flgs[application+"/"+environment+"/"+configuration][key]
	if !ok {
		return control.ScopedFlag{}, control.ErrScopeNotFound
	}
	return flag, nil
}
func (f *scopeFake) UpsertScopedFlag(_ context.Context, application, environment, configuration, key string, value json.RawMessage) (control.ScopedFlag, error) {
	cfg, err := f.resolve(application, environment, configuration)
	if err != nil {
		return control.ScopedFlag{}, err
	}
	cfg.Revision++
	enabled := true
	var b bool
	if json.Unmarshal(value, &b) == nil {
		enabled = b
	}
	flag := control.ScopedFlag{Key: key, Enabled: enabled, Value: value, Revision: cfg.Revision}
	k := application + "/" + environment + "/" + configuration
	if f.flgs[k] == nil {
		f.flgs[k] = map[string]control.ScopedFlag{}
	}
	f.flgs[k][key] = flag
	return flag, nil
}
func (f *scopeFake) SetScopedFlagEnabled(_ context.Context, application, environment, configuration, key string, enabled bool) (control.ScopedFlag, error) {
	cfg, err := f.resolve(application, environment, configuration)
	if err != nil {
		return control.ScopedFlag{}, err
	}
	k := application + "/" + environment + "/" + configuration
	flag, ok := f.flgs[k][key]
	if !ok {
		return control.ScopedFlag{}, control.ErrScopeNotFound
	}
	cfg.Revision++
	flag.Enabled = enabled
	flag.Revision = cfg.Revision
	f.flgs[k][key] = flag
	return flag, nil
}
func (f *scopeFake) DeleteScopedFlag(_ context.Context, application, environment, configuration, key string) error {
	cfg, err := f.resolve(application, environment, configuration)
	if err != nil {
		return err
	}
	k := application + "/" + environment + "/" + configuration
	if _, ok := f.flgs[k][key]; !ok {
		return control.ErrScopeNotFound
	}
	cfg.Revision++
	delete(f.flgs[k], key)
	return nil
}

func TestApplicationsHTTPFullHierarchyFlow(t *testing.T) {
	service := control.NewScopeService(newScopeFake())
	handler := NewApplicationsHandler(service)
	server := httptest.NewServer(handler)
	defer server.Close()

	resp := request(t, server, http.MethodPost, "/v1/applications", jsonBody(t, map[string]any{"name": "atlas"}))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create application status=%d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = request(t, server, http.MethodPost, "/v1/applications/atlas/environments", jsonBody(t, map[string]any{"name": "production"}))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create environment status=%d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = request(t, server, http.MethodPost, "/v1/applications/atlas/environments/production/configurations", jsonBody(t, map[string]any{"name": "mobile-menu"}))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create configuration status=%d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = request(t, server, http.MethodPost, "/v1/applications/atlas/environments/production/configurations/mobile-menu/flags",
		jsonBody(t, map[string]any{"key": "new-ui", "enabled": true}))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create flag status=%d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = request(t, server, http.MethodGet, "/v1/applications/atlas/environments/production/configurations/mobile-menu", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get configuration status=%d", resp.StatusCode)
	}
	etag := resp.Header.Get("ETag")
	if etag == "" {
		t.Fatal("expected ETag header on configuration snapshot")
	}
	resp.Body.Close()

	req, err := http.NewRequest(http.MethodGet, server.URL+"/v1/applications/atlas/environments/production/configurations/mobile-menu", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("If-None-Match", etag)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusNotModified {
		t.Fatalf("expected 304 with matching If-None-Match, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = request(t, server, http.MethodPatch, "/v1/applications/atlas/environments/production/configurations/mobile-menu/flags/new-ui",
		jsonBody(t, map[string]any{"enabled": false}))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("patch flag status=%d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = request(t, server, http.MethodDelete, "/v1/applications/atlas/environments/production/configurations/mobile-menu/flags/new-ui", nil)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete flag status=%d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = request(t, server, http.MethodGet, "/v1/applications/atlas/environments/production/configurations/mobile-menu/flags/new-ui", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestApplicationsHTTPMissingAncestorIs404(t *testing.T) {
	service := control.NewScopeService(newScopeFake())
	handler := NewApplicationsHandler(service)
	server := httptest.NewServer(handler)
	defer server.Close()

	resp := request(t, server, http.MethodPost, "/v1/applications/does-not-exist/environments", jsonBody(t, map[string]any{"name": "production"}))
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for missing application, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestApplicationsHTTPWithoutPostgresIs501(t *testing.T) {
	handler := NewApplicationsHandler(nil)
	server := httptest.NewServer(handler)
	defer server.Close()

	resp := request(t, server, http.MethodGet, "/v1/applications", nil)
	if resp.StatusCode != http.StatusNotImplemented {
		t.Fatalf("expected 501 when the hierarchy has no PostgreSQL-backed service, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}
