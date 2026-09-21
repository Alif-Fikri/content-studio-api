package render

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

func (r *Repo) Create(ctx context.Context, contentItemID string) error {
	_, err := r.pool.Exec(ctx, `
		insert into render_jobs (content_item_id) values ($1)
	`, contentItemID)
	return err
}

func (r *Repo) LatestForContentItem(ctx context.Context, contentItemID string) (*Job, error) {
	row := r.pool.QueryRow(ctx, `
		select id, content_item_id, status, error, started_at, finished_at, created_at
		from render_jobs
		where content_item_id = $1
		order by created_at desc
		limit 1
	`, contentItemID)
	return scanJob(row)
}

func (r *Repo) Claim(ctx context.Context) (*Job, error) {
	row := r.pool.QueryRow(ctx, `
		update render_jobs
		set status = 'rendering', started_at = now()
		where id = (
			select id from render_jobs
			where status = 'queued'
			order by created_at
			limit 1
			for update skip locked
		)
		returning id, content_item_id, status, error, started_at, finished_at, created_at
	`)
	return scanJob(row)
}

func (r *Repo) MarkDone(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		update render_jobs set status = 'done', finished_at = now() where id = $1
	`, id)
	return err
}

func (r *Repo) MarkFailed(ctx context.Context, id, errMsg string) error {
	_, err := r.pool.Exec(ctx, `
		update render_jobs set status = 'failed', error = $2, finished_at = now() where id = $1
	`, id, errMsg)
	return err
}

type row interface {
	Scan(dest ...any) error
}

func scanJob(r row) (*Job, error) {
	var job Job
	err := r.Scan(&job.ID, &job.ContentItemID, &job.Status, &job.Error, &job.StartedAt, &job.FinishedAt, &job.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &job, nil
}
