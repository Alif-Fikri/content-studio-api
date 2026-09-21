package content

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

func (r *Repo) Create(ctx context.Context, product, title, brief string) (*Item, error) {
	row := r.pool.QueryRow(ctx, `
		insert into content_items (product, title, brief)
		values ($1, $2, $3)
		returning id, product, title, brief, raw_video_key, status, caption, script, rendered_video_key, created_at, updated_at
	`, product, title, brief)
	return scanItem(row)
}

func (r *Repo) List(ctx context.Context) ([]*Item, error) {
	rows, err := r.pool.Query(ctx, `
		select id, product, title, brief, raw_video_key, status, caption, script, rendered_video_key, created_at, updated_at
		from content_items
		order by created_at desc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*Item
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repo) Get(ctx context.Context, id string) (*Item, error) {
	row := r.pool.QueryRow(ctx, `
		select id, product, title, brief, raw_video_key, status, caption, script, rendered_video_key, created_at, updated_at
		from content_items
		where id = $1
	`, id)
	return scanItem(row)
}

func (r *Repo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `delete from content_items where id = $1`, id)
	return err
}

func (r *Repo) SetRawVideoKey(ctx context.Context, id, key string) error {
	_, err := r.pool.Exec(ctx, `
		update content_items set raw_video_key = $2, updated_at = now() where id = $1
	`, id, key)
	return err
}

func (r *Repo) ClearVideoKeys(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		update content_items set raw_video_key = null, rendered_video_key = null, updated_at = now() where id = $1
	`, id)
	return err
}

type ExpiredRenderedItem struct {
	ID               string
	RawVideoKey      *string
	RenderedVideoKey string
}

func (r *Repo) ListExpiredRendered(ctx context.Context, retentionDays int) ([]ExpiredRenderedItem, error) {
	rows, err := r.pool.Query(ctx, `
		select id, raw_video_key, rendered_video_key
		from content_items
		where status = 'ready'
		  and rendered_video_key is not null
		  and updated_at < now() - make_interval(days => $1)
	`, retentionDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ExpiredRenderedItem
	for rows.Next() {
		var item ExpiredRenderedItem
		if err := rows.Scan(&item.ID, &item.RawVideoKey, &item.RenderedVideoKey); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repo) SetStatus(ctx context.Context, id string, status Status) error {
	_, err := r.pool.Exec(ctx, `
		update content_items set status = $2, updated_at = now() where id = $1
	`, id, status)
	return err
}

func (r *Repo) SetGenerated(ctx context.Context, id, caption string, script []ScriptBeat, status Status) error {
	scriptJSON, err := json.Marshal(script)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		update content_items
		set caption = $2, script = $3, status = $4, updated_at = now()
		where id = $1
	`, id, caption, scriptJSON, status)
	return err
}

func (r *Repo) SetApproved(ctx context.Context, id, caption string, script []ScriptBeat) error {
	scriptJSON, err := json.Marshal(script)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		update content_items
		set caption = $2, script = $3, status = $4, updated_at = now()
		where id = $1
	`, id, caption, scriptJSON, StatusRendering)
	return err
}

func (r *Repo) SetRenderedVideoKey(ctx context.Context, id, key string) error {
	_, err := r.pool.Exec(ctx, `
		update content_items set rendered_video_key = $2, status = $3, updated_at = now() where id = $1
	`, id, key, StatusReady)
	return err
}

type row interface {
	Scan(dest ...any) error
}

func scanItem(r row) (*Item, error) {
	var item Item
	var scriptJSON []byte
	err := r.Scan(
		&item.ID, &item.Product, &item.Title, &item.Brief, &item.RawVideoKey,
		&item.Status, &item.Caption, &scriptJSON, &item.RenderedVideoKey,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	if len(scriptJSON) > 0 {
		if err := json.Unmarshal(scriptJSON, &item.Script); err != nil {
			return nil, err
		}
	}
	return &item, nil
}
