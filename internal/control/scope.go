package control

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrScopeInvalidName = errors.New("name must contain between 1 and 128 characters, with no leading or trailing spaces")
	ErrScopeExists      = errors.New("already exists")
	ErrScopeNotFound    = errors.New("not found")
)

type Application struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Environment struct {
	ID            int64     `json:"id"`
	ApplicationID int64     `json:"application_id"`
	Name          string    `json:"name"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Configuration struct {
	ID            int64     `json:"id"`
	EnvironmentID int64     `json:"environment_id"`
	Name          string    `json:"name"`
	Revision      uint64    `json:"revision"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ScopedFlag is a Flag that lives inside an application/environment/configuration, as opposed to
// the legacy global Flag. The same key can exist as both a legacy flag and any number of
// ScopedFlags, or in several configurations at once, without colliding.
type ScopedFlag struct {
	Key      string          `json:"key"`
	Enabled  bool            `json:"enabled"`
	Value    json.RawMessage `json:"value,omitempty"`
	Revision uint64          `json:"revision"`
}

type ConfigurationSnapshot struct {
	Application   string       `json:"application"`
	Environment   string       `json:"environment"`
	Configuration string       `json:"configuration"`
	Revision      uint64       `json:"revision"`
	UpdatedAt     time.Time    `json:"updated_at"`
	Flags         []ScopedFlag `json:"flags"`
}

// ValidateScopeName applies the same 1-128 character, no-surrounding-whitespace rule used for
// flag keys (see validateKey) to application/environment/configuration names.
func ValidateScopeName(name string) error {
	if err := validateKey(name); err != nil {
		return ErrScopeInvalidName
	}
	return nil
}

// ScopeRepository persists the application/environment/configuration hierarchy and the flags
// that live inside a configuration. It is implemented by *persistence.Store.
type ScopeRepository interface {
	CreateApplication(ctx context.Context, name string) (Application, error)
	ListApplications(ctx context.Context) ([]Application, error)

	CreateEnvironment(ctx context.Context, application, name string) (Environment, error)
	ListEnvironments(ctx context.Context, application string) ([]Environment, error)

	CreateConfiguration(ctx context.Context, application, environment, name string) (Configuration, error)
	ListConfigurations(ctx context.Context, application, environment string) ([]Configuration, error)
	GetConfigurationSnapshot(ctx context.Context, application, environment, configuration string) (ConfigurationSnapshot, error)

	GetScopedFlag(ctx context.Context, application, environment, configuration, key string) (ScopedFlag, error)
	UpsertScopedFlag(ctx context.Context, application, environment, configuration, key string, value json.RawMessage) (ScopedFlag, error)
	SetScopedFlagEnabled(ctx context.Context, application, environment, configuration, key string, enabled bool) (ScopedFlag, error)
	DeleteScopedFlag(ctx context.Context, application, environment, configuration, key string) error
}

// ScopeService validates input and enforces create-vs-update semantics on top of a
// ScopeRepository, mirroring how PersistentService sits on top of Repository for legacy flags.
type ScopeService struct{ repo ScopeRepository }

func NewScopeService(repo ScopeRepository) *ScopeService { return &ScopeService{repo: repo} }

func (s *ScopeService) CreateApplication(ctx context.Context, name string) (Application, error) {
	if err := ValidateScopeName(name); err != nil {
		return Application{}, err
	}
	return s.repo.CreateApplication(ctx, name)
}

func (s *ScopeService) ListApplications(ctx context.Context) ([]Application, error) {
	return s.repo.ListApplications(ctx)
}

func (s *ScopeService) CreateEnvironment(ctx context.Context, application, name string) (Environment, error) {
	if err := ValidateScopeName(application); err != nil {
		return Environment{}, err
	}
	if err := ValidateScopeName(name); err != nil {
		return Environment{}, err
	}
	return s.repo.CreateEnvironment(ctx, application, name)
}

func (s *ScopeService) ListEnvironments(ctx context.Context, application string) ([]Environment, error) {
	if err := ValidateScopeName(application); err != nil {
		return nil, err
	}
	return s.repo.ListEnvironments(ctx, application)
}

func (s *ScopeService) CreateConfiguration(ctx context.Context, application, environment, name string) (Configuration, error) {
	if err := ValidateScopeName(application); err != nil {
		return Configuration{}, err
	}
	if err := ValidateScopeName(environment); err != nil {
		return Configuration{}, err
	}
	if err := ValidateScopeName(name); err != nil {
		return Configuration{}, err
	}
	return s.repo.CreateConfiguration(ctx, application, environment, name)
}

func (s *ScopeService) ListConfigurations(ctx context.Context, application, environment string) ([]Configuration, error) {
	if err := ValidateScopeName(application); err != nil {
		return nil, err
	}
	if err := ValidateScopeName(environment); err != nil {
		return nil, err
	}
	return s.repo.ListConfigurations(ctx, application, environment)
}

func (s *ScopeService) GetConfiguration(ctx context.Context, application, environment, configuration string) (ConfigurationSnapshot, error) {
	if err := validateScopePath(application, environment, configuration); err != nil {
		return ConfigurationSnapshot{}, err
	}
	return s.repo.GetConfigurationSnapshot(ctx, application, environment, configuration)
}

func (s *ScopeService) GetFlag(ctx context.Context, application, environment, configuration, key string) (ScopedFlag, error) {
	if err := validateScopePath(application, environment, configuration); err != nil {
		return ScopedFlag{}, err
	}
	if err := validateKey(key); err != nil {
		return ScopedFlag{}, err
	}
	return s.repo.GetScopedFlag(ctx, application, environment, configuration, key)
}

// CreateFlag fails with ErrScopeExists if the key already exists in this configuration,
// mirroring the legacy Service.Create / PersistentService.CreateValue semantics.
func (s *ScopeService) CreateFlag(ctx context.Context, application, environment, configuration, key string, value json.RawMessage) (ScopedFlag, error) {
	if err := validateScopePath(application, environment, configuration); err != nil {
		return ScopedFlag{}, err
	}
	if err := validateKey(key); err != nil {
		return ScopedFlag{}, err
	}
	if err := validateValue(value); err != nil {
		return ScopedFlag{}, err
	}
	if _, err := s.repo.GetScopedFlag(ctx, application, environment, configuration, key); err == nil {
		return ScopedFlag{}, ErrScopeExists
	} else if !errors.Is(err, ErrScopeNotFound) {
		return ScopedFlag{}, err
	}
	return s.repo.UpsertScopedFlag(ctx, application, environment, configuration, key, value)
}

// UpdateFlag fails with ErrScopeNotFound if the key does not already exist in this
// configuration, mirroring the legacy PersistentService.UpdateValue semantics.
func (s *ScopeService) UpdateFlag(ctx context.Context, application, environment, configuration, key string, value json.RawMessage) (ScopedFlag, error) {
	if err := validateScopePath(application, environment, configuration); err != nil {
		return ScopedFlag{}, err
	}
	if err := validateKey(key); err != nil {
		return ScopedFlag{}, err
	}
	if err := validateValue(value); err != nil {
		return ScopedFlag{}, err
	}
	if _, err := s.repo.GetScopedFlag(ctx, application, environment, configuration, key); err != nil {
		return ScopedFlag{}, err
	}
	return s.repo.UpsertScopedFlag(ctx, application, environment, configuration, key, value)
}

func (s *ScopeService) SetFlagEnabled(ctx context.Context, application, environment, configuration, key string, enabled bool) (ScopedFlag, error) {
	if err := validateScopePath(application, environment, configuration); err != nil {
		return ScopedFlag{}, err
	}
	if err := validateKey(key); err != nil {
		return ScopedFlag{}, err
	}
	return s.repo.SetScopedFlagEnabled(ctx, application, environment, configuration, key, enabled)
}

func (s *ScopeService) DeleteFlag(ctx context.Context, application, environment, configuration, key string) error {
	if err := validateScopePath(application, environment, configuration); err != nil {
		return err
	}
	if err := validateKey(key); err != nil {
		return err
	}
	return s.repo.DeleteScopedFlag(ctx, application, environment, configuration, key)
}

func validateScopePath(application, environment, configuration string) error {
	if err := ValidateScopeName(application); err != nil {
		return err
	}
	if err := ValidateScopeName(environment); err != nil {
		return err
	}
	return ValidateScopeName(configuration)
}
