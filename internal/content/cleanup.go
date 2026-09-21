package content

import (
	"context"
	"time"
)

const cleanupInterval = 6 * time.Hour

type CleanupStorage interface {
	Delete(ctx context.Context, key string) error
}

type Cleanup struct {
	repo          *Repo
	storage       CleanupStorage
	retentionDays int
}

func NewCleanup(repo *Repo, storage CleanupStorage, retentionDays int) *Cleanup {
	return &Cleanup{repo: repo, storage: storage, retentionDays: retentionDays}
}

func (c *Cleanup) Start(ctx context.Context) {
	ticker := time.NewTicker(cleanupInterval)
	go func() {
		defer ticker.Stop()
		c.run(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.run(ctx)
			}
		}
	}()
}

func (c *Cleanup) run(ctx context.Context) {
	items, err := c.repo.ListExpiredRendered(ctx, c.retentionDays)
	if err != nil {
		return
	}

	for _, item := range items {
		if err := c.storage.Delete(ctx, item.RenderedVideoKey); err != nil {
			continue
		}
		_ = c.repo.ClearRenderedVideoKey(ctx, item.ID)
	}
}
