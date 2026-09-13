package persistence

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/richardnogueira01/feature-flag-mvp/internal/control"
)

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (s *Store) CreateApplication(ctx context.Context, name string) (control.Application, error) {
	var a control.Application
	err := s.pool.QueryRow(ctx,
		`INSERT INTO applications(name) VALUES ($1) RETURNING id, name, created_at, updated_at`,
		name,
	).Scan(&a.ID, &a.Name, &a.CreatedAt, &a.UpdatedAt)
	if isUniqueViolation(err) {
		return control.Application{}, control.ErrScopeExists
	}
	return a, err
}

func (s *Store) ListApplications(ctx context.Context) ([]control.Application, error) {
	rows, err := s.pool.Query(ctx, `SELECT id, name, created_at, updated_at FROM applications ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []control.Application
	for rows.Next() {
		var a control.Application
		if err := rows.Scan(&a.ID, &a.Name, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) applicationID(ctx context.Context, name string) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `SELECT id FROM applications WHERE name=$1`, name).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, control.ErrScopeNotFound
	}
	return id, err
}

func (s *Store) CreateEnvironment(ctx context.Context, application, name string) (control.Environment, error) {
	appID, err := s.applicationID(ctx, application)
	if err != nil {
		return control.Environment{}, err
	}
	var e control.Environment
	err = s.pool.QueryRow(ctx,
		`INSERT INTO environments(application_id, name) VALUES ($1, $2)
		 RETURNING id, application_id, name, created_at, updated_at`,
		appID, name,
	).Scan(&e.ID, &e.ApplicationID, &e.Name, &e.CreatedAt, &e.UpdatedAt)
	if isUniqueViolation(err) {
		return control.Environment{}, control.ErrScopeExists
	}
	return e, err
}

func (s *Store) ListEnvironments(ctx context.Context, application string) ([]control.Environment, error) {
	appID, err := s.applicationID(ctx, application)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx,
		`SELECT id, application_id, name, created_at, updated_at FROM environments WHERE application_id=$1 ORDER BY name`,
		appID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []control.Environment
	for rows.Next() {
		var e control.Environment
		if err := rows.Scan(&e.ID, &e.ApplicationID, &e.Name, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (s *Store) environmentID(ctx context.Context, application, environment string) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		SELECT e.id FROM environments e
		JOIN applications a ON a.id = e.application_id
		WHERE a.name = $1 AND e.name = $2`, application, environment).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, control.ErrScopeNotFound
	}
	return id, err
}

func (s *Store) CreateConfiguration(ctx context.Context, application, environment, name string) (control.Configuration, error) {
	envID, err := s.environmentID(ctx, application, environment)
	if err != nil {
		return control.Configuration{}, err
	}
	var c control.Configuration
	err = s.pool.QueryRow(ctx,
		`INSERT INTO configurations(environment_id, name) VALUES ($1, $2)
		 RETURNING id, environment_id, name, revision, created_at, updated_at`,
		envID, name,
	).Scan(&c.ID, &c.EnvironmentID, &c.Name, &c.Revision, &c.CreatedAt, &c.UpdatedAt)
	if isUniqueViolation(err) {
		return control.Configuration{}, control.ErrScopeExists
	}
	return c, err
}

func (s *Store) ListConfigurations(ctx context.Context, application, environment string) ([]control.Configuration, error) {
	envID, err := s.environmentID(ctx, application, environment)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx,
		`SELECT id, environment_id, name, revision, created_at, updated_at FROM configurations WHERE environment_id=$1 ORDER BY name`,
		envID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []control.Configuration
	for rows.Next() {
		var c control.Configuration
		if err := rows.Scan(&c.ID, &c.EnvironmentID, &c.Name, &c.Revision, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// configurationID resolves an (application, environment, configuration) path to the
// configuration's id, or control.ErrScopeNotFound if any level of the path is missing.
func (s *Store) configurationID(ctx context.Context, application, environment, configuration string) (int64, error) {
	var id int64
	err := s.pool.QueryRow(ctx, `
		SELECT c.id FROM configurations c
		JOIN environments e ON e.id = c.environment_id
		JOIN applications a ON a.id = e.application_id
		WHERE a.name = $1 AND e.name = $2 AND c.name = $3`,
		application, environment, configuration).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, control.ErrScopeNotFound
	}
	return id, err
}

func (s *Store) GetConfigurationSnapshot(ctx context.Context, application, environment, configuration string) (control.ConfigurationSnapshot, error) {
	cfgID, err := s.configurationID(ctx, application, environment, configuration)
	if err != nil {
		return control.ConfigurationSnapshot{}, err
	}
	var revision uint64
	if err := s.pool.QueryRow(ctx, `SELECT revision FROM configurations WHERE id=$1`, cfgID).Scan(&revision); err != nil {
		return control.ConfigurationSnapshot{}, err
	}
	rows, err := s.pool.Query(ctx,
		`SELECT key, enabled, value, revision FROM feature_flags WHERE configuration_id=$1 ORDER BY key`, cfgID)
	if err != nil {
		return control.ConfigurationSnapshot{}, err
	}
	defer rows.Close()
	flags := []control.ScopedFlag{}
	for rows.Next() {
		var f control.ScopedFlag
		var v []byte
		if err := rows.Scan(&f.Key, &f.Enabled, &v, &f.Revision); err != nil {
			return control.ConfigurationSnapshot{}, err
		}
		f.Value = v
		flags = append(flags, f)
	}
	if err := rows.Err(); err != nil {
		return control.ConfigurationSnapshot{}, err
	}
	return control.ConfigurationSnapshot{
		Application:   application,
		Environment:   environment,
		Configuration: configuration,
		Revision:      revision,
		Flags:         flags,
	}, nil
}

func (s *Store) GetScopedFlag(ctx context.Context, application, environment, configuration, key string) (control.ScopedFlag, error) {
	cfgID, err := s.configurationID(ctx, application, environment, configuration)
	if err != nil {
		return control.ScopedFlag{}, err
	}
	var f control.ScopedFlag
	var v []byte
	err = s.pool.QueryRow(ctx,
		`SELECT key, enabled, value, revision FROM feature_flags WHERE configuration_id=$1 AND key=$2`,
		cfgID, key,
	).Scan(&f.Key, &f.Enabled, &v, &f.Revision)
	if errors.Is(err, pgx.ErrNoRows) {
		return control.ScopedFlag{}, control.ErrScopeNotFound
	}
	if err != nil {
		return control.ScopedFlag{}, err
	}
	f.Value = v
	return f, nil
}

// UpsertScopedFlag creates or replaces a flag inside a configuration, bumping the
// configuration's revision, recording history and an outbox event, all in one transaction.
func (s *Store) UpsertScopedFlag(ctx context.Context, application, environment, configuration, key string, value json.RawMessage) (control.ScopedFlag, error) {
	cfgID, err := s.configurationID(ctx, application, environment, configuration)
	if err != nil {
		return control.ScopedFlag{}, err
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return control.ScopedFlag{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var revision uint64
	if err := tx.QueryRow(ctx,
		`UPDATE configurations SET revision=revision+1, updated_at=now() WHERE id=$1 RETURNING revision`,
		cfgID,
	).Scan(&revision); err != nil {
		return control.ScopedFlag{}, err
	}

	enabled := true
	var boolValue bool
	if json.Unmarshal(value, &boolValue) == nil {
		enabled = boolValue
	}

	var f control.ScopedFlag
	var v []byte
	err = tx.QueryRow(ctx, `
		INSERT INTO feature_flags(key, enabled, value, revision, configuration_id)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (configuration_id, key) DO UPDATE
			SET enabled=EXCLUDED.enabled, value=EXCLUDED.value, revision=EXCLUDED.revision, updated_at=now()
		RETURNING key, enabled, value, revision`,
		key, enabled, value, revision, cfgID,
	).Scan(&f.Key, &f.Enabled, &v, &f.Revision)
	if err != nil {
		return control.ScopedFlag{}, err
	}
	f.Value = v

	history, _ := json.Marshal(f)
	if _, err := tx.Exec(ctx,
		`INSERT INTO flag_history(revision, key, operation, payload, configuration_id) VALUES ($1, $2, 'upsert', $3, $4)`,
		revision, key, history, cfgID,
	); err != nil {
		return control.ScopedFlag{}, err
	}

	payload, err := scopedSnapshotPayload(ctx, tx, application, environment, configuration, cfgID, revision)
	if err != nil {
		return control.ScopedFlag{}, err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO outbox_events(revision, event_type, payload, configuration_id) VALUES ($1, 'configuration.updated', $2, $3)`,
		revision, payload, cfgID,
	); err != nil {
		return control.ScopedFlag{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return control.ScopedFlag{}, err
	}
	return f, nil
}

func (s *Store) SetScopedFlagEnabled(ctx context.Context, application, environment, configuration, key string, enabled bool) (control.ScopedFlag, error) {
	cfgID, err := s.configurationID(ctx, application, environment, configuration)
	if err != nil {
		return control.ScopedFlag{}, err
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return control.ScopedFlag{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var revision uint64
	if err := tx.QueryRow(ctx,
		`UPDATE configurations SET revision=revision+1, updated_at=now() WHERE id=$1 RETURNING revision`,
		cfgID,
	).Scan(&revision); err != nil {
		return control.ScopedFlag{}, err
	}

	var f control.ScopedFlag
	var v []byte
	err = tx.QueryRow(ctx,
		`UPDATE feature_flags SET enabled=$3, revision=$4, updated_at=now() WHERE configuration_id=$1 AND key=$2
		 RETURNING key, enabled, value, revision`,
		cfgID, key, enabled, revision,
	).Scan(&f.Key, &f.Enabled, &v, &f.Revision)
	if errors.Is(err, pgx.ErrNoRows) {
		return control.ScopedFlag{}, control.ErrScopeNotFound
	}
	if err != nil {
		return control.ScopedFlag{}, err
	}
	f.Value = v

	history, _ := json.Marshal(f)
	if _, err := tx.Exec(ctx,
		`INSERT INTO flag_history(revision, key, operation, payload, configuration_id) VALUES ($1, $2, 'toggle', $3, $4)`,
		revision, key, history, cfgID,
	); err != nil {
		return control.ScopedFlag{}, err
	}

	payload, err := scopedSnapshotPayload(ctx, tx, application, environment, configuration, cfgID, revision)
	if err != nil {
		return control.ScopedFlag{}, err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO outbox_events(revision, event_type, payload, configuration_id) VALUES ($1, 'configuration.updated', $2, $3)`,
		revision, payload, cfgID,
	); err != nil {
		return control.ScopedFlag{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return control.ScopedFlag{}, err
	}
	return f, nil
}

func (s *Store) DeleteScopedFlag(ctx context.Context, application, environment, configuration, key string) error {
	cfgID, err := s.configurationID(ctx, application, environment, configuration)
	if err != nil {
		return err
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var revision uint64
	if err := tx.QueryRow(ctx,
		`UPDATE configurations SET revision=revision+1, updated_at=now() WHERE id=$1 RETURNING revision`,
		cfgID,
	).Scan(&revision); err != nil {
		return err
	}

	tag, err := tx.Exec(ctx, `DELETE FROM feature_flags WHERE configuration_id=$1 AND key=$2`, cfgID, key)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return control.ErrScopeNotFound
	}

	p, _ := json.Marshal(map[string]any{"key": key, "revision": revision})
	if _, err := tx.Exec(ctx,
		`INSERT INTO flag_history(revision, key, operation, payload, configuration_id) VALUES ($1, $2, 'delete', $3, $4)`,
		revision, key, p, cfgID,
	); err != nil {
		return err
	}

	payload, err := scopedSnapshotPayload(ctx, tx, application, environment, configuration, cfgID, revision)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO outbox_events(revision, event_type, payload, configuration_id) VALUES ($1, 'configuration.updated', $2, $3)`,
		revision, payload, cfgID,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func scopedSnapshotPayload(ctx context.Context, tx pgx.Tx, application, environment, configuration string, configurationID int64, revision uint64) ([]byte, error) {
	rows, err := tx.Query(ctx, `SELECT key, enabled, value FROM feature_flags WHERE configuration_id=$1 ORDER BY key`, configurationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	flags := []control.ScopedFlag{}
	for rows.Next() {
		var f control.ScopedFlag
		var v []byte
		if err := rows.Scan(&f.Key, &f.Enabled, &v); err != nil {
			return nil, err
		}
		f.Value = v
		flags = append(flags, f)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return json.Marshal(control.ConfigurationSnapshot{
		Application:   application,
		Environment:   environment,
		Configuration: configuration,
		Revision:      revision,
		Flags:         flags,
	})
}
