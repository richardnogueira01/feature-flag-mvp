package control

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

// fakeScopeRepo is a minimal in-memory ScopeRepository used to unit test ScopeService without a
// real PostgreSQL instance. It intentionally mirrors persistence.Store's behavior/error
// semantics (ErrScopeNotFound / ErrScopeExists, per-configuration monotonic revision).
type fakeScopeRepo struct {
	apps map[string]Application
	envs map[string]map[string]Environment
	cfgs map[string]map[string]map[string]*Configuration
	flgs map[string]map[string]ScopedFlag // "application/environment/configuration" -> key -> flag
}

func newFakeScopeRepo() *fakeScopeRepo {
	return &fakeScopeRepo{
		apps: map[string]Application{},
		envs: map[string]map[string]Environment{},
		cfgs: map[string]map[string]map[string]*Configuration{},
		flgs: map[string]map[string]ScopedFlag{},
	}
}

func scopeKey(a, e, c string) string { return a + "/" + e + "/" + c }

func (f *fakeScopeRepo) CreateApplication(_ context.Context, name string) (Application, error) {
	if _, ok := f.apps[name]; ok {
		return Application{}, ErrScopeExists
	}
	app := Application{ID: int64(len(f.apps) + 1), Name: name}
	f.apps[name] = app
	return app, nil
}
func (f *fakeScopeRepo) ListApplications(_ context.Context) ([]Application, error) {
	out := make([]Application, 0, len(f.apps))
	for _, a := range f.apps {
		out = append(out, a)
	}
	return out, nil
}
func (f *fakeScopeRepo) CreateEnvironment(_ context.Context, application, name string) (Environment, error) {
	if _, ok := f.apps[application]; !ok {
		return Environment{}, ErrScopeNotFound
	}
	if f.envs[application] == nil {
		f.envs[application] = map[string]Environment{}
	}
	if _, ok := f.envs[application][name]; ok {
		return Environment{}, ErrScopeExists
	}
	env := Environment{ID: int64(len(f.envs[application]) + 1), Name: name}
	f.envs[application][name] = env
	return env, nil
}
func (f *fakeScopeRepo) ListEnvironments(_ context.Context, application string) ([]Environment, error) {
	if _, ok := f.apps[application]; !ok {
		return nil, ErrScopeNotFound
	}
	out := make([]Environment, 0, len(f.envs[application]))
	for _, e := range f.envs[application] {
		out = append(out, e)
	}
	return out, nil
}
func (f *fakeScopeRepo) CreateConfiguration(_ context.Context, application, environment, name string) (Configuration, error) {
	if _, ok := f.envs[application][environment]; !ok {
		return Configuration{}, ErrScopeNotFound
	}
	if f.cfgs[application] == nil {
		f.cfgs[application] = map[string]map[string]*Configuration{}
	}
	if f.cfgs[application][environment] == nil {
		f.cfgs[application][environment] = map[string]*Configuration{}
	}
	if _, ok := f.cfgs[application][environment][name]; ok {
		return Configuration{}, ErrScopeExists
	}
	cfg := &Configuration{ID: int64(len(f.cfgs[application][environment]) + 1), Name: name}
	f.cfgs[application][environment][name] = cfg
	return *cfg, nil
}
func (f *fakeScopeRepo) ListConfigurations(_ context.Context, application, environment string) ([]Configuration, error) {
	if _, ok := f.envs[application][environment]; !ok {
		return nil, ErrScopeNotFound
	}
	out := make([]Configuration, 0, len(f.cfgs[application][environment]))
	for _, c := range f.cfgs[application][environment] {
		out = append(out, *c)
	}
	return out, nil
}
func (f *fakeScopeRepo) resolveConfiguration(application, environment, configuration string) (*Configuration, error) {
	byEnv, ok := f.cfgs[application]
	if !ok {
		return nil, ErrScopeNotFound
	}
	byCfg, ok := byEnv[environment]
	if !ok {
		return nil, ErrScopeNotFound
	}
	cfg, ok := byCfg[configuration]
	if !ok {
		return nil, ErrScopeNotFound
	}
	return cfg, nil
}
func (f *fakeScopeRepo) GetConfigurationSnapshot(_ context.Context, application, environment, configuration string) (ConfigurationSnapshot, error) {
	cfg, err := f.resolveConfiguration(application, environment, configuration)
	if err != nil {
		return ConfigurationSnapshot{}, err
	}
	var flags []ScopedFlag
	for _, fl := range f.flgs[scopeKey(application, environment, configuration)] {
		flags = append(flags, fl)
	}
	return ConfigurationSnapshot{Application: application, Environment: environment, Configuration: configuration, Revision: cfg.Revision, Flags: flags}, nil
}
func (f *fakeScopeRepo) GetScopedFlag(_ context.Context, application, environment, configuration, key string) (ScopedFlag, error) {
	if _, err := f.resolveConfiguration(application, environment, configuration); err != nil {
		return ScopedFlag{}, err
	}
	flag, ok := f.flgs[scopeKey(application, environment, configuration)][key]
	if !ok {
		return ScopedFlag{}, ErrScopeNotFound
	}
	return flag, nil
}
func (f *fakeScopeRepo) UpsertScopedFlag(_ context.Context, application, environment, configuration, key string, value json.RawMessage) (ScopedFlag, error) {
	cfg, err := f.resolveConfiguration(application, environment, configuration)
	if err != nil {
		return ScopedFlag{}, err
	}
	cfg.Revision++
	enabled := true
	var b bool
	if json.Unmarshal(value, &b) == nil {
		enabled = b
	}
	flag := ScopedFlag{Key: key, Enabled: enabled, Value: value, Revision: cfg.Revision}
	k := scopeKey(application, environment, configuration)
	if f.flgs[k] == nil {
		f.flgs[k] = map[string]ScopedFlag{}
	}
	f.flgs[k][key] = flag
	return flag, nil
}
func (f *fakeScopeRepo) SetScopedFlagEnabled(_ context.Context, application, environment, configuration, key string, enabled bool) (ScopedFlag, error) {
	cfg, err := f.resolveConfiguration(application, environment, configuration)
	if err != nil {
		return ScopedFlag{}, err
	}
	k := scopeKey(application, environment, configuration)
	flag, ok := f.flgs[k][key]
	if !ok {
		return ScopedFlag{}, ErrScopeNotFound
	}
	cfg.Revision++
	flag.Enabled = enabled
	flag.Revision = cfg.Revision
	f.flgs[k][key] = flag
	return flag, nil
}
func (f *fakeScopeRepo) DeleteScopedFlag(_ context.Context, application, environment, configuration, key string) error {
	cfg, err := f.resolveConfiguration(application, environment, configuration)
	if err != nil {
		return err
	}
	k := scopeKey(application, environment, configuration)
	if _, ok := f.flgs[k][key]; !ok {
		return ErrScopeNotFound
	}
	cfg.Revision++
	delete(f.flgs[k], key)
	return nil
}

func setupScope(t *testing.T, service *ScopeService) {
	t.Helper()
	ctx := context.Background()
	if _, err := service.CreateApplication(ctx, "atlas"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateEnvironment(ctx, "atlas", "production"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateConfiguration(ctx, "atlas", "production", "mobile-menu"); err != nil {
		t.Fatal(err)
	}
}

func TestScopeServiceSameKeyAcrossConfigurationsDoesNotCollide(t *testing.T) {
	ctx := context.Background()
	service := NewScopeService(newFakeScopeRepo())
	setupScope(t, service)
	if _, err := service.CreateConfiguration(ctx, "atlas", "production", "checkout"); err != nil {
		t.Fatal(err)
	}

	value, _ := json.Marshal(true)
	if _, err := service.CreateFlag(ctx, "atlas", "production", "mobile-menu", "new-ui", value); err != nil {
		t.Fatalf("create in mobile-menu: %v", err)
	}
	valueFalse, _ := json.Marshal(false)
	if _, err := service.CreateFlag(ctx, "atlas", "production", "checkout", "new-ui", valueFalse); err != nil {
		t.Fatalf("create in checkout: %v", err)
	}

	a, err := service.GetFlag(ctx, "atlas", "production", "mobile-menu", "new-ui")
	if err != nil || !a.Enabled {
		t.Fatalf("mobile-menu/new-ui = %+v, err=%v", a, err)
	}
	b, err := service.GetFlag(ctx, "atlas", "production", "checkout", "new-ui")
	if err != nil || b.Enabled {
		t.Fatalf("checkout/new-ui = %+v, err=%v", b, err)
	}
}

func TestScopeServiceRevisionIsMonotonicPerConfiguration(t *testing.T) {
	ctx := context.Background()
	service := NewScopeService(newFakeScopeRepo())
	setupScope(t, service)

	value, _ := json.Marshal(true)
	created, err := service.CreateFlag(ctx, "atlas", "production", "mobile-menu", "a", value)
	if err != nil || created.Revision != 1 {
		t.Fatalf("create a: flag=%+v err=%v", created, err)
	}
	created2, err := service.CreateFlag(ctx, "atlas", "production", "mobile-menu", "b", value)
	if err != nil || created2.Revision != 2 {
		t.Fatalf("create b: flag=%+v err=%v", created2, err)
	}
	snapshot, err := service.GetConfiguration(ctx, "atlas", "production", "mobile-menu")
	if err != nil || snapshot.Revision != 2 || len(snapshot.Flags) != 2 {
		t.Fatalf("snapshot=%+v err=%v", snapshot, err)
	}
}

func TestScopeServiceCreateThenCreateAgainConflicts(t *testing.T) {
	ctx := context.Background()
	service := NewScopeService(newFakeScopeRepo())
	setupScope(t, service)
	value, _ := json.Marshal(true)
	if _, err := service.CreateFlag(ctx, "atlas", "production", "mobile-menu", "a", value); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateFlag(ctx, "atlas", "production", "mobile-menu", "a", value); !errors.Is(err, ErrScopeExists) {
		t.Fatalf("expected ErrScopeExists, got %v", err)
	}
}

func TestScopeServiceUpdateMissingFlagNotFound(t *testing.T) {
	ctx := context.Background()
	service := NewScopeService(newFakeScopeRepo())
	setupScope(t, service)
	value, _ := json.Marshal(true)
	if _, err := service.UpdateFlag(ctx, "atlas", "production", "mobile-menu", "missing", value); !errors.Is(err, ErrScopeNotFound) {
		t.Fatalf("expected ErrScopeNotFound, got %v", err)
	}
}

func TestScopeServiceMissingAncestorIsNotFound(t *testing.T) {
	ctx := context.Background()
	service := NewScopeService(newFakeScopeRepo())
	if _, err := service.CreateEnvironment(ctx, "does-not-exist", "production"); !errors.Is(err, ErrScopeNotFound) {
		t.Fatalf("expected ErrScopeNotFound for missing application, got %v", err)
	}
	setupScope(t, service)
	if _, err := service.GetConfiguration(ctx, "atlas", "production", "does-not-exist"); !errors.Is(err, ErrScopeNotFound) {
		t.Fatalf("expected ErrScopeNotFound for missing configuration, got %v", err)
	}
}

func TestScopeServiceInvalidNameRejected(t *testing.T) {
	ctx := context.Background()
	service := NewScopeService(newFakeScopeRepo())
	if _, err := service.CreateApplication(ctx, " padded "); !errors.Is(err, ErrScopeInvalidName) {
		t.Fatalf("expected ErrScopeInvalidName, got %v", err)
	}
	if _, err := service.CreateApplication(ctx, ""); !errors.Is(err, ErrScopeInvalidName) {
		t.Fatalf("expected ErrScopeInvalidName for empty name, got %v", err)
	}
}

func TestScopeServiceDeleteAndSetEnabled(t *testing.T) {
	ctx := context.Background()
	service := NewScopeService(newFakeScopeRepo())
	setupScope(t, service)
	value, _ := json.Marshal(true)
	if _, err := service.CreateFlag(ctx, "atlas", "production", "mobile-menu", "a", value); err != nil {
		t.Fatal(err)
	}
	toggled, err := service.SetFlagEnabled(ctx, "atlas", "production", "mobile-menu", "a", false)
	if err != nil || toggled.Enabled {
		t.Fatalf("toggle: flag=%+v err=%v", toggled, err)
	}
	if err := service.DeleteFlag(ctx, "atlas", "production", "mobile-menu", "a"); err != nil {
		t.Fatal(err)
	}
	if _, err := service.GetFlag(ctx, "atlas", "production", "mobile-menu", "a"); !errors.Is(err, ErrScopeNotFound) {
		t.Fatalf("expected ErrScopeNotFound after delete, got %v", err)
	}
}
