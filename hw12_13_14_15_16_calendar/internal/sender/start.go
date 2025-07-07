package sender

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/types" //nolint:depguard,nolintlint
)

// Start starts a goroutine which listens to the message queue and logs notifications.
func (s *Sender) Start(ctx context.Context) error {
	ch, err := s.broker.Consume(ctx)
	if err != nil {
		s.l.Error(ctx, "unexpected error on broker consume", slog.Any("error", err))
		return err
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case data, ok := <-ch:
				if !ok {
					return
				}
				s.handleEventSending(ctx, data)
			}
		}
	}()

	return nil
}

// handleEventSending logs the notification.
// If any error occurs on umnarshalling or data parsing, it will be logged with WARN level.
func (s *Sender) handleEventSending(ctx context.Context, data []byte) {
	var message types.Notification
	err := json.Unmarshal(data, &message)
	if err != nil {
		s.l.Warn(ctx, "unmarshal notification", slog.Any("error", err))
		return
	}
	datetime, err := message.GetDatetime()
	if err != nil {
		s.l.Warn(ctx, "invalid datetime format", slog.String("datetime", message.Datetime), slog.Any("error", err))
	}
	s.l.Info(
		ctx, "notification sent",
		slog.Group(
			"notification",
			slog.String("id", message.ID),
			slog.String("event_id", message.Title),
			slog.String("user_id", message.UserID),
			slog.Time("notification_type", datetime),
		),
	)
}
