package ads

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo *Repo
	meta *MetaClient
}

func NewHandler(repo *Repo, meta *MetaClient) *Handler {
	return &Handler{repo: repo, meta: meta}
}

func (h *Handler) Register(rg *gin.RouterGroup) {
	rg.POST("/ads", h.create)
	rg.GET("/ads", h.list)
	rg.GET("/ads/:id", h.get)
	rg.GET("/ads/:id/metrics", h.metrics)
	rg.POST("/ads/sync", h.sync)
}

type createRequest struct {
	ContentItemID *string  `json:"content_item_id"`
	Platform      Platform `json:"platform" binding:"required"`
	Spend         float64  `json:"spend"`
	StartedAt     string   `json:"started_at" binding:"required"`
	ExternalAdID  *string  `json:"external_ad_id"`
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry, err := h.repo.Create(c.Request.Context(), req.ContentItemID, req.Platform, req.Spend, req.StartedAt, req.ExternalAdID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

func (h *Handler) list(c *gin.Context) {
	entries, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entries)
}

func (h *Handler) get(c *gin.Context) {
	entry, err := h.repo.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, entry)
}

func (h *Handler) metrics(c *gin.Context) {
	metrics, err := h.repo.Metrics(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, metrics)
}

const defaultLookbackDate = "2024-01-01"

func (h *Handler) sync(c *gin.Context) {
	entries, err := h.repo.ListWithExternalID(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	synced := 0
	for _, entry := range entries {
		if err := h.syncEntry(c.Request.Context(), entry); err != nil {
			continue
		}
		synced++
	}

	c.JSON(http.StatusOK, gin.H{"synced": synced, "total": len(entries)})
}

func (h *Handler) syncEntry(ctx context.Context, entry *Entry) error {
	since, err := h.repo.LatestMetricDate(ctx, entry.ID)
	if err != nil {
		return err
	}

	sinceDate := defaultLookbackDate
	if since != nil {
		sinceDate = *since
	}

	rows, err := h.meta.FetchInsights(ctx, *entry.ExternalAdID, sinceDate)
	if err != nil {
		return err
	}

	for _, row := range rows {
		impressions, _ := strconv.Atoi(row.Impressions)
		reach, _ := strconv.Atoi(row.Reach)
		clicks, _ := strconv.Atoi(row.Clicks)
		spend, _ := strconv.ParseFloat(row.Spend, 64)

		if err := h.repo.UpsertMetric(ctx, entry.ID, row.Date, impressions, reach, clicks, spend); err != nil {
			return err
		}
	}

	return nil
}
