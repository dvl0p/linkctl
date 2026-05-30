package daemon

import (
	"context"
	"log/slog"
	"time"
)

func (d *daemon) StartWorker(ctx context.Context,
	linkID int64, url string, intervalSeconds int64) {
	done := make(chan struct{})
	select {
	case <-ctx.Done():
		return
	case d.eventQueue <- eventStart{
		linkID:          linkID,
		url:             url,
		intervalSeconds: intervalSeconds,
		done:            done}:
	}
	select {
	case <-done:
	case <-ctx.Done():
	}
}

func (d *daemon) StopWorker(ctx context.Context, linkID int64) {
	done := make(chan struct{})
	select {
	case d.eventQueue <- eventStop{linkID: linkID, done: done}:
	case <-ctx.Done():
		return
	}
	select {
	case <-done:
	case <-ctx.Done():
	}
}

func (d *daemon) CountWorkers(ctx context.Context) int {
	ch := make(chan int, 1)
	select {
	case <-ctx.Done():
		return 0
	case d.eventQueue <- eventCount{result: ch}:
	}
	select {
	case <-ctx.Done():
		return 0
	case count := <-ch:
		return count
	}
}

func (d *daemon) worker(ctx context.Context,
	linkID int64, url string, intervalSeconds int64) {
	ticker := time.NewTicker(time.Second * time.Duration(intervalSeconds))
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.logger.Debug("triggered link health check",
				slog.Int64("link_id", linkID),
				slog.String("url", url),
			)
		}
	}
}
