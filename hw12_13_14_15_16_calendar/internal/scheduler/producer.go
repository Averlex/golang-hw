package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/errors" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/types"                //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                               //nolint:depguard,nolintlint
)

type queueTransport struct {
	Notifications []*types.Notification
	IDs           []uuid.UUID
}

// StartProducer starts the producer goroutine. Non-blocking. Requires call to Scheduler.Wait().
//
// Goroutine is receiving notifications from a separate queue, marshalling it and trying to
// send it to the broker.
//
// All IDs whose messages were successfully sent to the broker will be updated in the storage
// as notified ones.
func (sch *Scheduler) StartProducer(ctx context.Context) {
	ch := sch.runNotificationQueue(ctx)

	sch.wg.Add(1)

	go func() {
		defer sch.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case data := <-ch:
				// Sending messages to the broker.
				successIDs := sch.handleNotificationsSending(ctx, data)

				// Updating the events with the IDs that were successfully sent to the broker.
				sch.handleStorageUpdate(ctx, successIDs, data)
			}
		}
	}()
}

// runNotificationQueue starts notification queue in a separate goroutine. Non-blocking.
// Requires a Scheduler.Wait() call to wait for the goroutine to finish.
//
// Returns a read-only channel for []Notification transport.
func (sch *Scheduler) runNotificationQueue(ctx context.Context) <-chan *queueTransport {
	ch := make(chan *queueTransport)

	sch.mu.RLock()
	queueInterval := sch.queueInterval
	sch.mu.RUnlock()

	sch.wg.Add(1)

	go func() {
		defer close(ch)
		sch.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(queueInterval):
				events := sch.handleNotificationsGet(ctx)
				if len(events) == 0 {
					sch.l.Debug(ctx, "got no events for notification")
					continue
				}
				notifications, ids := convertEventsToNotifications(events)
				data := &queueTransport{
					Notifications: notifications,
					IDs:           ids,
				}
				select {
				case <-ctx.Done():
					return
				case ch <- data:
				}
			}
		}
	}()

	return ch
}

// handleNotificationsGet gets events from the storage for notification queue.
func (sch *Scheduler) handleNotificationsGet(ctx context.Context) []*types.Event {
	var events []*types.Event
	err := sch.withRetries(ctx, "GetEventsForNotification", func() error {
		localEvents, localErr := sch.s.GetEventsForNotification(ctx)
		if localErr != nil {
			if !errors.Is(localErr, projectErrors.ErrEventNotFound) {
				return localErr
			}
		}
		events = localEvents
		return nil
	})
	if err != nil {
		sch.l.Error(ctx, "get events for notification queue", slog.Any("error", err))
		return nil
	}
	return events
}

// handleNotificationsSending gets the events from the internal queue,
// marshals them and sends notifications to the broker.
// Returns a slice of IDs that were successfully sent to the broker.
func (sch *Scheduler) handleNotificationsSending(ctx context.Context, data *queueTransport) []uuid.UUID {
	successIDs := make([]uuid.UUID, 0, len(data.Notifications))
	for i, notification := range data.Notifications {
		messageData, err := json.Marshal(notification)
		if err != nil {
			sch.l.Warn(ctx, "marshal notification", slog.Any("error", err))
			continue
		}
		err = sch.broker.Produce(ctx, messageData)
		if err != nil {
			sch.l.Error(
				ctx,
				"unexpected error on broker produce",
				slog.String("id", notification.ID),
				slog.Any("error", err),
			)
			continue
		}
		successIDs = append(successIDs, data.IDs[i])
	}
	return successIDs
}

// handleStorageUpdate updates the events with the IDs that were successfully sent to the broker.
// Method logs the actual and expected numbers of updated events.
func (sch *Scheduler) handleStorageUpdate(ctx context.Context, successIDs []uuid.UUID, data *queueTransport) {
	var updatedCount int64
	err := sch.withRetries(ctx, "UpdateNotifiedEvents", func() error {
		count, localErr := sch.s.UpdateNotifiedEvents(ctx, successIDs)
		if localErr != nil {
			return localErr
		}
		updatedCount = count
		return nil
	})
	if err != nil {
		sch.l.Error(ctx, "update events for notification queue", slog.Any("error", err))
		return
	}
	sch.l.Debug(
		ctx,
		"updated notified events",
		slog.Int64("count", updatedCount),
		slog.Int64("failed to update", int64(len(data.Notifications))-updatedCount),
	)
}
