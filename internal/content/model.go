package content

import "time"

type Status string

const (
	StatusDraft         Status = "draft"
	StatusGenerating    Status = "generating"
	StatusReadyToRender Status = "ready_to_render"
	StatusRendering     Status = "rendering"
	StatusReady         Status = "ready"
	StatusFailed        Status = "failed"
)

type ScriptBeat struct {
	Text            string  `json:"text"`
	StartSeconds    float64 `json:"start_seconds"`
	DurationSeconds float64 `json:"duration_seconds"`
}

type Item struct {
	ID               string       `json:"id"`
	Product          string       `json:"product"`
	Title            string       `json:"title"`
	Brief            string       `json:"brief"`
	RawVideoKey      *string      `json:"raw_video_key,omitempty"`
	Status           Status       `json:"status"`
	Caption          *string      `json:"caption,omitempty"`
	Script           []ScriptBeat `json:"script,omitempty"`
	RenderedVideoKey *string      `json:"rendered_video_key,omitempty"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}
