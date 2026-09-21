package ads

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

func (r *Repo) Create(ctx context.Context, contentItemID *string, platform Platform, spend float64, startedAt string, externalAdID *string) (*Entry, error) {
	row := r.pool.QueryRow(ctx, `
		insert into ad_entries (content_item_id, platform, spend, started_at, external_ad_id)
		values ($1, $2, $3, $4, $5)
		returning id, content_item_id, platform, external_ad_id, spend, started_at, created_at
	`, contentItemID, platform, spend, startedAt, externalAdID)
	return scanEntry(row)
}

func (r *Repo) List(ctx context.Context) ([]*Entry, error) {
	rows, err := r.pool.Query(ctx, `
		select id, content_item_id, platform, external_ad_id, spend, started_at, created_at
		from ad_entries
		order by created_at desc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*Entry
	for rows.Next() {
		entry, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (r *Repo) Get(ctx context.Context, id string) (*Entry, error) {
	row := r.pool.QueryRow(ctx, `
		select id, content_item_id, platform, external_ad_id, spend, started_at, created_at
		from ad_entries
		where id = $1
	`, id)
	return scanEntry(row)
}

func (r *Repo) ListWithExternalID(ctx context.Context) ([]*Entry, error) {
	rows, err := r.pool.Query(ctx, `
		select id, content_item_id, platform, external_ad_id, spend, started_at, created_at
		from ad_entries
		where external_ad_id is not null
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*Entry
	for rows.Next() {
		entry, err := scanEntry(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func (r *Repo) Metrics(ctx context.Context, adEntryID string) ([]*Metric, error) {
	rows, err := r.pool.Query(ctx, `
		select id, ad_entry_id, date, impressions, reach, clicks, spend, created_at
		from ad_metrics
		where ad_entry_id = $1
		order by date
	`, adEntryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*Metric
	for rows.Next() {
		var m Metric
		if err := rows.Scan(&m.ID, &m.AdEntryID, &m.Date, &m.Impressions, &m.Reach, &m.Clicks, &m.Spend, &m.CreatedAt); err != nil {
			return nil, err
		}
		metrics = append(metrics, &m)
	}
	return metrics, rows.Err()
}

func (r *Repo) LatestMetricDate(ctx context.Context, adEntryID string) (*string, error) {
	var date *string
	err := r.pool.QueryRow(ctx, `
		select max(date)::text from ad_metrics where ad_entry_id = $1
	`, adEntryID).Scan(&date)
	return date, err
}

func (r *Repo) UpsertMetric(ctx context.Context, adEntryID, date string, impressions, reach, clicks int, spend float64) error {
	_, err := r.pool.Exec(ctx, `
		insert into ad_metrics (ad_entry_id, date, impressions, reach, clicks, spend)
		values ($1, $2, $3, $4, $5, $6)
		on conflict (ad_entry_id, date)
		do update set impressions = excluded.impressions, reach = excluded.reach, clicks = excluded.clicks, spend = excluded.spend
	`, adEntryID, date, impressions, reach, clicks, spend)
	return err
}

type row interface {
	Scan(dest ...any) error
}

func scanEntry(r row) (*Entry, error) {
	var e Entry
	err := r.Scan(&e.ID, &e.ContentItemID, &e.Platform, &e.ExternalAdID, &e.Spend, &e.StartedAt, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}
