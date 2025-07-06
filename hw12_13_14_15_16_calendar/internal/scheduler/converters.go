package scheduler

import (
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/types" //nolint:depguard,nolintlint
	"github.com/google/uuid"                                                //nolint:depguard,nolintlint
)

// convertEventsToNotifications converts a slice of internal events to the slice of notifications.
func convertEventsToNotifications(events []*types.Event) ([]*types.Notification, []uuid.UUID) {
	notifications := make([]*types.Notification, len(events))
	ids := make([]uuid.UUID, len(events))
	for i, event := range events {
		notifications[i] = event.ToNotification()
		ids[i] = event.ID
	}
	return notifications, ids
}
