package ads

import "time"

type Platform string

const (
	PlatformInstagram Platform = "instagram"
	PlatformFacebook  Platform = "facebook"
)

type Entry struct {
	ID            string    `json:"id"`
	ContentItemID *string   `json:"content_item_id,omitempty"`
	Platform      Platform  `json:"platform"`
	ExternalAdID  *string   `json:"external_ad_id,omitempty"`
	Spend         float64   `json:"spend"`
	StartedAt     time.Time `json:"started_at"`
	CreatedAt     time.Time `json:"created_at"`
}

type Metric struct {
	ID          string    `json:"id"`
	AdEntryID   string    `json:"ad_entry_id"`
	Date        time.Time `json:"date"`
	Impressions int       `json:"impressions"`
	Reach       int       `json:"reach"`
	Clicks      int       `json:"clicks"`
	Spend       float64   `json:"spend"`
	CreatedAt   time.Time `json:"created_at"`
}
