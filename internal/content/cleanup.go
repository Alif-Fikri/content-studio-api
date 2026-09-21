package content

import (
	"context"
	"log"
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
		log.Printf("cleanup: list expired rendered items: %v", err)
		return
	}

	for _, item := range items {
		if item.RawVideoKey != nil {
			if err := c.storage.Delete(ctx, *item.RawVideoKey); err != nil {
				log.Printf("cleanup: delete raw video %s for item %s: %v", *item.RawVideoKey, item.ID, err)
			}
		}

		if err := c.storage.Delete(ctx, item.RenderedVideoKey); err != nil {
			log.Printf("cleanup: delete rendered video %s for item %s: %v", item.RenderedVideoKey, item.ID, err)
			continue
		}

		if err := c.repo.ClearVideoKeys(ctx, item.ID); err != nil {
			log.Printf("cleanup: clear video keys for item %s: %v", item.ID, err)
		}
	}
}
