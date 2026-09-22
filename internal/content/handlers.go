package content

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Generator interface {
	Generate(ctx context.Context, provider, model, prompt string) (caption string, script []ScriptBeat, err error)
}

type Storage interface {
	PresignUpload(ctx context.Context, key string) (string, error)
	PresignDownload(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
}

type RenderJobCreator interface {
	Create(ctx context.Context, contentItemID string) error
}

type Handler struct {
	repo       *Repo
	generator  Generator
	storage    Storage
	renderJobs RenderJobCreator
}

func NewHandler(repo *Repo, generator Generator, storage Storage, renderJobs RenderJobCreator) *Handler {
	return &Handler{repo: repo, generator: generator, storage: storage, renderJobs: renderJobs}
}

func (h *Handler) Register(rg *gin.RouterGroup) {
	rg.POST("/content-items", h.create)
	rg.POST("/content-items/:id/upload-complete", h.uploadComplete)
	rg.GET("/content-items", h.list)
	rg.GET("/content-items/:id", h.get)
	rg.GET("/content-items/:id/download-url", h.downloadURL)
	rg.DELETE("/content-items/:id", h.delete)
	rg.POST("/content-items/:id/generate", h.generate)
	rg.POST("/content-items/:id/approve", h.approve)
}

type createRequest struct {
	Product string `json:"product" binding:"required"`
	Title   string `json:"title" binding:"required"`
	Brief   string `json:"brief" binding:"required"`
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.repo.Create(c.Request.Context(), req.Product, req.Title, req.Brief)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uploadKey := "raw/" + item.ID
	uploadURL, err := h.storage.PresignUpload(c.Request.Context(), uploadKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": item.ID, "upload_url": uploadURL})
}

func (h *Handler) uploadComplete(c *gin.Context) {
	id := c.Param("id")
	key := "raw/" + id
	if err := h.repo.SetRawVideoKey(c.Request.Context(), id, key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) list(c *gin.Context) {
	items, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *Handler) get(c *gin.Context) {
	item, err := h.repo.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) downloadURL(c *gin.Context) {
	item, err := h.repo.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	key := item.RenderedVideoKey
	if key == nil {
		key = item.RawVideoKey
	}
	if key == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no video available for this item"})
		return
	}

	url, err := h.storage.PresignDownload(c.Request.Context(), *key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"download_url": url})
}

func (h *Handler) delete(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")

	item, err := h.repo.Get(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if item.Status == StatusRendering {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot delete while a render is in progress"})
		return
	}

	if item.RawVideoKey != nil {
		if err := h.storage.Delete(ctx, *item.RawVideoKey); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	if item.RenderedVideoKey != nil {
		if err := h.storage.Delete(ctx, *item.RenderedVideoKey); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := h.repo.Delete(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type generateRequest struct {
	Prompt   string `json:"prompt" binding:"required"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

func (h *Handler) generate(c *gin.Context) {
	id := c.Param("id")

	var req generateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.repo.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if item.Status != StatusDraft {
		c.JSON(http.StatusConflict, gin.H{"error": "generate is only allowed while the item is in draft status"})
		return
	}

	if err := h.repo.SetStatus(c.Request.Context(), id, StatusGenerating); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go h.runGenerate(id, req.Provider, req.Model, req.Prompt)

	c.Status(http.StatusAccepted)
}

func (h *Handler) runGenerate(id, provider, model, prompt string) {
	ctx := context.Background()
	caption, script, err := h.generator.Generate(ctx, provider, model, prompt)
	if err != nil {
		_ = h.repo.SetStatus(ctx, id, StatusDraft)
		return
	}
	_ = h.repo.SetGenerated(ctx, id, caption, script, StatusReadyToRender)
}

type approveRequest struct {
	Caption string       `json:"caption" binding:"required"`
	Script  []ScriptBeat `json:"script" binding:"required"`
}

func (h *Handler) approve(c *gin.Context) {
	id := c.Param("id")
	var req approveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item, err := h.repo.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	switch item.Status {
	case StatusRendering:
		c.JSON(http.StatusConflict, gin.H{"error": "a render is already in progress for this item"})
		return
	case StatusDraft, StatusGenerating:
		c.JSON(http.StatusConflict, gin.H{"error": "approve requires a generated script first"})
		return
	}

	if err := h.repo.SetApproved(c.Request.Context(), id, req.Caption, req.Script); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.renderJobs.Create(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusAccepted)
}
