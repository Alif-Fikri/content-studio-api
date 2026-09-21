package render

import "time"

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRendering Status = "rendering"
	StatusDone      Status = "done"
	StatusFailed    Status = "failed"
)

type Job struct {
	ID            string     `json:"id"`
	ContentItemID string     `json:"content_item_id"`
	Status        Status     `json:"status"`
	Error         *string    `json:"error,omitempty"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
