package render

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo *Repo
}

func NewHandler(repo *Repo) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Register(rg *gin.RouterGroup) {
	rg.GET("/content-items/:id/render", h.latest)
}

func (h *Handler) latest(c *gin.Context) {
	job, err := h.repo.LatestForContentItem(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no render job found"})
		return
	}
	c.JSON(http.StatusOK, job)
}
