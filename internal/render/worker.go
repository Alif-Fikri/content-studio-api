package render

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/alchemist/content-studio-api/internal/content"
)

type Storage interface {
	Download(ctx context.Context, key, destPath string) error
	Upload(ctx context.Context, key, srcPath string) error
}

type Pool struct {
	jobs             *Repo
	items            *content.Repo
	storage          Storage
	backgroundAudio  string
	pollInterval     time.Duration
	workerCount      int
}

func NewPool(jobs *Repo, items *content.Repo, storage Storage, backgroundAudio string, workerCount int) *Pool {
	return &Pool{
		jobs:            jobs,
		items:           items,
		storage:         storage,
		backgroundAudio: backgroundAudio,
		pollInterval:    3 * time.Second,
		workerCount:     workerCount,
	}
}

func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workerCount; i++ {
		go p.loop(ctx)
	}
}

func (p *Pool) loop(ctx context.Context) {
	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.tryClaimAndRun(ctx)
		}
	}
}

func (p *Pool) tryClaimAndRun(ctx context.Context) {
	job, err := p.jobs.Claim(ctx)
	if err != nil || job == nil {
		return
	}
	p.run(ctx, job)
}

func (p *Pool) run(ctx context.Context, job *Job) {
	item, err := p.items.Get(ctx, job.ContentItemID)
	if err != nil {
		_ = p.jobs.MarkFailed(ctx, job.ID, err.Error())
		return
	}

	workDir, err := os.MkdirTemp("", "render-"+job.ID)
	if err != nil {
		_ = p.jobs.MarkFailed(ctx, job.ID, err.Error())
		return
	}
	defer os.RemoveAll(workDir)

	inputPath := filepath.Join(workDir, "input.mp4")
	outputPath := filepath.Join(workDir, "output.mp4")

	if item.RawVideoKey == nil {
		_ = p.jobs.MarkFailed(ctx, job.ID, "content item has no raw video")
		return
	}

	if err := p.storage.Download(ctx, *item.RawVideoKey, inputPath); err != nil {
		_ = p.jobs.MarkFailed(ctx, job.ID, err.Error())
		return
	}

	args := BuildFFmpegArgs(inputPath, p.backgroundAudio, outputPath, item.Script)
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		_ = p.jobs.MarkFailed(ctx, job.ID, fmt.Sprintf("%v: %s", err, string(output)))
		return
	}

	renderedKey := "rendered/" + item.ID + ".mp4"
	if err := p.storage.Upload(ctx, renderedKey, outputPath); err != nil {
		_ = p.jobs.MarkFailed(ctx, job.ID, err.Error())
		return
	}

	if err := p.items.SetRenderedVideoKey(ctx, item.ID, renderedKey); err != nil {
		_ = p.jobs.MarkFailed(ctx, job.ID, err.Error())
		return
	}

	_ = p.jobs.MarkDone(ctx, job.ID)
}
