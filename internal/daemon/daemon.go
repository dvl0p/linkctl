package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/dvl0p/linkctl/internal/store"
)

type daemon struct {
	eventQueue chan event
	logger     *slog.Logger
	store      *store.Store
	ctx        context.Context
	wg         sync.WaitGroup
}

func New(store *store.Store, logger *slog.Logger) *daemon {
	return &daemon{
		eventQueue: make(chan event, 100),
		logger:     logger,
		store:      store,
	}
}

func (d *daemon) CloseQueue() {
	d.logger.Debug("closing daemon event queue")
	close(d.eventQueue)
}

func (d *daemon) Start(ctx context.Context) error {
	d.ctx = ctx
	linksDB, err := d.store.Queries.ListLinks(d.ctx)
	if err != nil {
		return fmt.Errorf("could not retrieve links from db: %w", err)
	}
	workMap := make(map[int64]context.CancelFunc, len(linksDB))
	for _, linkDB := range linksDB {
		d.logger.Debug("initializing worker for link in db",
			slog.Int64("link_id", linkDB.ID),
			slog.String("link_url", linkDB.Url),
			slog.Int64("link_interval_seconds", linkDB.IntervalSeconds),
		)
		workCtx, workCancel := context.WithCancel(ctx)
		workMap[linkDB.ID] = workCancel
		d.wg.Go(func() {
			d.worker(workCtx, linkDB.ID, linkDB.Url, linkDB.IntervalSeconds)
		})
	}
	d.logger.Info("started linkctld daemon",
		slog.Int("workers", len(workMap)),
	)
	d.run(workMap)
	return nil
}

func (d *daemon) run(workMap map[int64]context.CancelFunc) {
	for {
		select {
		case <-d.ctx.Done():
			d.shutdown(workMap)
			return
		case ev, ok := <-d.eventQueue:
			if !ok {
				d.shutdown(workMap)
				return
			}
			switch ev := ev.(type) {
			case eventStart:
				d.starter(ev, workMap)
			case eventStop:
				d.stopper(ev, workMap)
			case eventCount:
				d.counter(ev, workMap)
			}
		}
	}
}

func (d *daemon) shutdown(workMap map[int64]context.CancelFunc) {
	for _, cancel := range workMap {
		cancel()
	}
	d.wg.Wait()
	d.logger.Info("shut down linkctld daemon",
		slog.Int("workers", len(workMap)),
	)
}
