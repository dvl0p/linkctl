package daemon

import (
	"context"
	"log/slog"
)

type event interface {
	isDaemonEvent()
}

type eventStart struct {
	linkID          int64
	url             string
	intervalSeconds int64
	done            chan struct{}
}

type eventStop struct {
	linkID int64
	done   chan struct{}
}

type eventCount struct {
	result chan int
}

func (ev eventStart) isDaemonEvent() {}
func (ev eventStop) isDaemonEvent()  {}
func (ev eventCount) isDaemonEvent() {}

func (d *daemon) starter(ev eventStart, ctx context.Context,
	workMap map[int64]context.CancelFunc) {
	defer close(ev.done)
	cancel, exists := workMap[ev.linkID]
	if exists {
		d.logger.Debug("stopping worker",
			slog.Int64("link_id", ev.linkID),
			slog.String("link_url", ev.url),
		)
		cancel()
	}
	d.logger.Debug("starting worker",
		slog.Int64("link_id", ev.linkID),
		slog.String("link_url", ev.url),
		slog.Int64("link_interval_seconds", ev.intervalSeconds),
	)
	workCtx, cancel := context.WithCancel(ctx)
	workMap[ev.linkID] = cancel
	d.wg.Go(func() {
		d.worker(workCtx, ev.linkID, ev.url, ev.intervalSeconds)
	})
}

func (d *daemon) stopper(ev eventStop, workMap map[int64]context.CancelFunc) {
	defer close(ev.done)
	if cancel, exists := workMap[ev.linkID]; exists {
		d.logger.Debug("stopping worker",
			slog.Int64("link_id", ev.linkID),
		)
		cancel()
		delete(workMap, ev.linkID)
	} else {
		d.logger.Debug("attepted stop on a non-running worker")
	}
}

func (d *daemon) counter(ev eventCount, workMap map[int64]context.CancelFunc) {
	ev.result <- len(workMap)
}
