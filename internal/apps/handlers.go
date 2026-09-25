package apps

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/alchemist/content-studio-api/internal/playstore"
)

type Storage interface {
	PresignUpload(ctx context.Context, key string) (string, error)
	Download(ctx context.Context, key, destPath string) error
}

type Publisher interface {
	PublishBundle(ctx context.Context, in playstore.ReleaseInput) error
	FetchRatingSummary(ctx context.Context, packageName string) (playstore.RatingSummary, error)
	FetchVitalsRates(ctx context.Context, packageName string) (playstore.VitalsRates, error)
}

type Handler struct {
	repo      *Repo
	storage   Storage
	publisher Publisher
}

func NewHandler(repo *Repo, storage Storage, publisher Publisher) *Handler {
	return &Handler{repo: repo, storage: storage, publisher: publisher}
}

func (h *Handler) Register(rg *gin.RouterGroup) {
	rg.POST("/apps", h.createApp)
	rg.GET("/apps", h.listApps)
	rg.GET("/apps/:id", h.getApp)
	rg.DELETE("/apps/:id", h.deleteApp)

	rg.POST("/apps/:id/releases", h.createRelease)
	rg.GET("/apps/:id/releases", h.listReleases)
	rg.GET("/apps/:id/releases/:release_id", h.getRelease)
	rg.POST("/apps/:id/releases/:release_id/upload-complete", h.releaseUploadComplete)

	rg.GET("/apps/:id/metrics", h.listMetrics)
	rg.POST("/apps/:id/metrics/sync", h.syncMetrics)
}

type createAppRequest struct {
	Name           string  `json:"name" binding:"required"`
	PackageName    string  `json:"package_name" binding:"required"`
	PlayConsoleURL *string `json:"play_console_url"`
}

func (h *Handler) createApp(c *gin.Context) {
	var req createAppRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	app, err := h.repo.CreateApp(c.Request.Context(), req.Name, req.PackageName, req.PlayConsoleURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, app)
}

func (h *Handler) listApps(c *gin.Context) {
	appsList, err := h.repo.ListApps(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, appsList)
}

func (h *Handler) getApp(c *gin.Context) {
	app, err := h.repo.GetApp(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, app)
}

func (h *Handler) deleteApp(c *gin.Context) {
	if err := h.repo.DeleteApp(c.Request.Context(), c.Param("id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type createReleaseRequest struct {
	Track        ReleaseTrack `json:"track" binding:"required"`
	VersionCode  int          `json:"version_code" binding:"required"`
	VersionName  *string      `json:"version_name"`
	ReleaseNotes *string      `json:"release_notes"`
}

func (h *Handler) createRelease(c *gin.Context) {
	if h.publisher == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google Play integration is not configured"})
		return
	}

	appID := c.Param("id")

	app, err := h.repo.GetApp(c.Request.Context(), appID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
		return
	}

	var req createReleaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	release, err := h.repo.CreateRelease(c.Request.Context(), app.ID, req.Track, req.VersionCode, req.VersionName, req.ReleaseNotes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	bundleKey := "bundles/" + release.ID + ".aab"
	uploadURL, err := h.storage.PresignUpload(c.Request.Context(), bundleKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": release.ID, "upload_url": uploadURL})
}

func (h *Handler) listReleases(c *gin.Context) {
	releases, err := h.repo.ListReleases(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, releases)
}

func (h *Handler) getRelease(c *gin.Context) {
	release, err := h.repo.GetRelease(c.Request.Context(), c.Param("release_id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, release)
}

func (h *Handler) releaseUploadComplete(c *gin.Context) {
	ctx := c.Request.Context()
	appID := c.Param("id")
	releaseID := c.Param("release_id")

	app, err := h.repo.GetApp(ctx, appID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "app not found"})
		return
	}

	release, err := h.repo.GetRelease(ctx, releaseID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "release not found"})
		return
	}

	bundleKey := "bundles/" + release.ID + ".aab"
	if err := h.repo.SetReleaseBundleKey(ctx, release.ID, bundleKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go h.runPublish(app, release.ID, bundleKey, string(release.Track), release.VersionCode, derefStr(release.ReleaseNotes))

	c.Status(http.StatusAccepted)
}

func (h *Handler) runPublish(app *App, releaseID, bundleKey, track string, versionCode int, releaseNotes string) {
	ctx := context.Background()

	_ = h.repo.SetReleaseStatus(ctx, releaseID, ReleaseStatusPublishing)

	workDir, err := os.MkdirTemp("", "release-"+releaseID)
	if err != nil {
		_ = h.repo.MarkReleaseFailed(ctx, releaseID, err.Error())
		return
	}
	defer os.RemoveAll(workDir)

	bundlePath := workDir + "/bundle.aab"
	if err := h.storage.Download(ctx, bundleKey, bundlePath); err != nil {
		_ = h.repo.MarkReleaseFailed(ctx, releaseID, err.Error())
		return
	}

	err = h.publisher.PublishBundle(ctx, playstore.ReleaseInput{
		PackageName:  app.PackageName,
		Track:        track,
		VersionCode:  int64(versionCode),
		ReleaseNotes: releaseNotes,
		BundlePath:   bundlePath,
	})
	if err != nil {
		_ = h.repo.MarkReleaseFailed(ctx, releaseID, err.Error())
		return
	}

	_ = h.repo.MarkReleaseRolledOut(ctx, releaseID)
}

func (h *Handler) listMetrics(c *gin.Context) {
	metrics, err := h.repo.ListMetrics(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

func (h *Handler) syncMetrics(c *gin.Context) {
	if h.publisher == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google Play integration is not configured"})
		return
	}

	ctx := c.Request.Context()
	app, err := h.repo.GetApp(ctx, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	rating, ratingErr := h.publisher.FetchRatingSummary(ctx, app.PackageName)
	rates, ratesErr := h.publisher.FetchVitalsRates(ctx, app.PackageName)

	if ratingErr != nil && ratesErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": ratingErr.Error()})
		return
	}

	date := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")

	var ratingAvg *float64
	var ratingCount *int
	if ratingErr == nil && rating.Count > 0 {
		ratingAvg = &rating.Average
		ratingCount = &rating.Count
	}

	var crashRate, anrRate *float64
	if ratesErr == nil {
		crashRate = &rates.CrashRate
		anrRate = &rates.AnrRate
	}

	if err := h.repo.UpsertMetric(ctx, app.ID, date, crashRate, anrRate, ratingAvg, ratingCount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"date":      date,
		"rating_ok": ratingErr == nil,
		"vitals_ok": ratesErr == nil,
	})
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
