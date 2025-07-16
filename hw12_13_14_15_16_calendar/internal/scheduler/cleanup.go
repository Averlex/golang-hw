package scheduler

import (
	"context"
	"log/slog"
	"time"
)

// StartCleanup starts the cleanup goroutine. Non-blocking. Requires call to Scheduler.Wait().
//
// Goroutine observes the storage periodaclly and deletes old events.
func (sch *Scheduler) StartCleanup(ctx context.Context) {
	sch.wg.Add(1)
	go func() {
		defer sch.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(sch.cleanupInterval):
				sch.handleStorageCleanup(ctx)
			}
		}
	}()
}

// handleStorageCleanup deletes old events from the storage.
// Method logs the actual number of deleted events.
// If no events were deleted, method logs with a debug level.
func (sch *Scheduler) handleStorageCleanup(ctx context.Context) {
	var deletedCount int64
	err := sch.withRetries(ctx, "DeleteOldEvents", func() error {
		cleanupTime := time.Now().AddDate(-1, 0, 0)
		count, localErr := sch.s.DeleteOldEvents(ctx, cleanupTime)
		if localErr != nil {
			return localErr
		}
		deletedCount = count
		return nil
	})
	if err != nil {
		sch.l.Error(ctx, "delete old events", slog.Any("error", err))
		return
	}

	if deletedCount == 0 {
		sch.l.Debug(ctx, "no old events to delete")
		return
	}
	sch.l.Info(
		ctx,
		"deleted old events",
		slog.Int64("count", deletedCount),
	)
}
