package memory

import (
	"sort"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/types" //nolint:depguard,nolintlint
)

// findInsertPosition finds the index in the events slice where a new event should be inserted
// to maintain ascending order by Datetime, using binary search.
// Returns the index where the event should be inserted.
//
//nolint:unused
func findInsertPosition(events []*types.Event, event *types.Event) int {
	return sort.Search(len(events), func(i int) bool {
		if events[i].Datetime.Equal(event.Datetime) {
			// If Datetime is equal, compare IDs for deterministic ordering
			return events[i].ID.String() > event.ID.String()
		}
		return events[i].Datetime.After(event.Datetime)
	})
}
