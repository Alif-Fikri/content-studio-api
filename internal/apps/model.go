package apps

import "time"

type ReleaseTrack string

const (
	TrackInternal   ReleaseTrack = "internal"
	TrackClosed     ReleaseTrack = "closed"
	TrackOpen       ReleaseTrack = "open"
	TrackProduction ReleaseTrack = "production"
)

type ReleaseStatus string

const (
	ReleaseStatusDraft      ReleaseStatus = "draft"
	ReleaseStatusUploaded   ReleaseStatus = "uploaded"
	ReleaseStatusPublishing ReleaseStatus = "publishing"
	ReleaseStatusInReview   ReleaseStatus = "in_review"
	ReleaseStatusRolledOut  ReleaseStatus = "rolled_out"
	ReleaseStatusHalted     ReleaseStatus = "halted"
	ReleaseStatusFailed     ReleaseStatus = "failed"
)

type App struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	PackageName    string    `json:"package_name"`
	PlayConsoleURL *string   `json:"play_console_url,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Release struct {
	ID           string        `json:"id"`
	AppID        string        `json:"app_id"`
	Track        ReleaseTrack  `json:"track"`
	VersionCode  int           `json:"version_code"`
	VersionName  *string       `json:"version_name,omitempty"`
	ReleaseNotes *string       `json:"release_notes,omitempty"`
	BundleKey    *string       `json:"-"`
	Status       ReleaseStatus `json:"status"`
	Error        *string       `json:"error,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	ReleasedAt   *time.Time    `json:"released_at,omitempty"`
}

type Metric struct {
	ID          string    `json:"id"`
	AppID       string    `json:"app_id"`
	Date        time.Time `json:"date"`
	CrashRate   *float64  `json:"crash_rate,omitempty"`
	AnrRate     *float64  `json:"anr_rate,omitempty"`
	RatingAvg   *float64  `json:"rating_avg,omitempty"`
	RatingCount *int      `json:"rating_count,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
