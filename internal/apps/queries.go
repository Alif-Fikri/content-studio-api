package apps

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) CreateApp(ctx context.Context, name, packageName string, playConsoleURL *string) (*App, error) {
	row := r.pool.QueryRow(ctx, `
		insert into apps (name, package_name, play_console_url)
		values ($1, $2, $3)
		returning id, name, package_name, play_console_url, created_at, updated_at
	`, name, packageName, playConsoleURL)
	return scanApp(row)
}

func (r *Repo) ListApps(ctx context.Context) ([]*App, error) {
	rows, err := r.pool.Query(ctx, `
		select id, name, package_name, play_console_url, created_at, updated_at
		from apps
		order by name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*App
	for rows.Next() {
		app, err := scanApp(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, app)
	}
	return result, rows.Err()
}

func (r *Repo) GetApp(ctx context.Context, id string) (*App, error) {
	row := r.pool.QueryRow(ctx, `
		select id, name, package_name, play_console_url, created_at, updated_at
		from apps
		where id = $1
	`, id)
	return scanApp(row)
}

func (r *Repo) DeleteApp(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `delete from apps where id = $1`, id)
	return err
}

func (r *Repo) CreateRelease(ctx context.Context, appID string, track ReleaseTrack, versionCode int, versionName, releaseNotes *string) (*Release, error) {
	row := r.pool.QueryRow(ctx, `
		insert into app_releases (app_id, track, version_code, version_name, release_notes)
		values ($1, $2, $3, $4, $5)
		returning id, app_id, track, version_code, version_name, release_notes, bundle_key, status, error, created_at, updated_at, released_at
	`, appID, track, versionCode, versionName, releaseNotes)
	return scanRelease(row)
}

func (r *Repo) ListReleases(ctx context.Context, appID string) ([]*Release, error) {
	rows, err := r.pool.Query(ctx, `
		select id, app_id, track, version_code, version_name, release_notes, bundle_key, status, error, created_at, updated_at, released_at
		from app_releases
		where app_id = $1
		order by created_at desc
	`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Release
	for rows.Next() {
		release, err := scanRelease(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, release)
	}
	return result, rows.Err()
}

func (r *Repo) GetRelease(ctx context.Context, id string) (*Release, error) {
	row := r.pool.QueryRow(ctx, `
		select id, app_id, track, version_code, version_name, release_notes, bundle_key, status, error, created_at, updated_at, released_at
		from app_releases
		where id = $1
	`, id)
	return scanRelease(row)
}

func (r *Repo) SetReleaseBundleKey(ctx context.Context, id, key string) error {
	_, err := r.pool.Exec(ctx, `
		update app_releases set bundle_key = $2, status = 'uploaded', updated_at = now() where id = $1
	`, id, key)
	return err
}

func (r *Repo) SetReleaseStatus(ctx context.Context, id string, status ReleaseStatus) error {
	_, err := r.pool.Exec(ctx, `
		update app_releases set status = $2, updated_at = now() where id = $1
	`, id, status)
	return err
}

func (r *Repo) MarkReleaseRolledOut(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		update app_releases set status = 'rolled_out', released_at = now(), updated_at = now() where id = $1
	`, id)
	return err
}

func (r *Repo) MarkReleaseFailed(ctx context.Context, id, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
		update app_releases set status = 'failed', error = $2, updated_at = now() where id = $1
	`, id, errMsg)
	return err
}

func (r *Repo) UpsertMetric(ctx context.Context, appID, date string, crashRate, anrRate, ratingAvg *float64, ratingCount *int) error {
	_, err := r.pool.Exec(ctx, `
		insert into app_metrics (app_id, date, crash_rate, anr_rate, rating_avg, rating_count)
		values ($1, $2, $3, $4, $5, $6)
		on conflict (app_id, date)
		do update set crash_rate = excluded.crash_rate, anr_rate = excluded.anr_rate,
			rating_avg = excluded.rating_avg, rating_count = excluded.rating_count
	`, appID, date, crashRate, anrRate, ratingAvg, ratingCount)
	return err
}

func (r *Repo) ListMetrics(ctx context.Context, appID string) ([]*Metric, error) {
	rows, err := r.pool.Query(ctx, `
		select id, app_id, date, crash_rate, anr_rate, rating_avg, rating_count, created_at
		from app_metrics
		where app_id = $1
		order by date
	`, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Metric
	for rows.Next() {
		var m Metric
		if err := rows.Scan(&m.ID, &m.AppID, &m.Date, &m.CrashRate, &m.AnrRate, &m.RatingAvg, &m.RatingCount, &m.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, &m)
	}
	return result, rows.Err()
}

type row interface {
	Scan(dest ...any) error
}

func scanApp(r row) (*App, error) {
	var a App
	err := r.Scan(&a.ID, &a.Name, &a.PackageName, &a.PlayConsoleURL, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func scanRelease(r row) (*Release, error) {
	var rel Release
	err := r.Scan(
		&rel.ID, &rel.AppID, &rel.Track, &rel.VersionCode, &rel.VersionName, &rel.ReleaseNotes,
		&rel.BundleKey, &rel.Status, &rel.Error, &rel.CreatedAt, &rel.UpdatedAt, &rel.ReleasedAt,
	)
	if err != nil {
		return nil, err
	}
	return &rel, nil
}
