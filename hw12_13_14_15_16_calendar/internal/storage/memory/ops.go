package memory

import (
	"sort"

	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/internal/types" //nolint:depguard,nolintlint
)

// findInsertPosition finds the index in the events slice where a new event should be inserted
// to maintain ascending order by Datetime, using binary search.
// Returns the index where the event should be inserted.
//
// Method is not bound to any concrete slice, so it can be used for any slice of events.
func (s *Storage) findInsertPosition(arr []*types.Event, elem *types.Event) int {
	return sort.Search(len(arr), func(i int) bool {
		if arr[i].Datetime.Equal(elem.Datetime) {
			// If Datetime is equal, compare IDs for deterministic ordering.
			return arr[i].ID.String() > elem.ID.String()
		}
		return arr[i].Datetime.After(elem.Datetime)
	})
}

// isOverlaps checks if the given event overlaps with any event in the sorted slice at the specified insertion position.
// Returns true if there is an overlap (excluding the event itself), false otherwise.
//
// IMPORTANT: case, where nextStart == prevEnd is not considered as overlap.
// Therefore, if the event starts at the same time as the previous event ends,
// it is considered as non-overlapping.
// This behavior is consistent with the original implementation and allows for events to be scheduled back-to-back.
func (s *Storage) isOverlaps(arr []*types.Event, elem *types.Event, pos int) bool {
	elemEnd := elem.Datetime.Add(elem.Duration)

	// Check for overlap with the previous event (if it exists).
	if pos > 0 {
		prev := arr[pos-1]
		// Skip if prev is the same event.
		if prev.ID != elem.ID {
			prevEnd := prev.Datetime.Add(prev.Duration)
			if prevEnd.After(elem.Datetime) && elemEnd.After(prev.Datetime) {
				return true
			}
		}
	}

	// Check for overlap with the next event (if it exists).
	if pos < len(arr) {
		next := arr[pos]
		// Skip if next is the same event.
		if next.ID != elem.ID {
			nextEnd := next.Datetime.Add(next.Duration)
			if nextEnd.After(elem.Datetime) && elemEnd.After(next.Datetime) {
				return true
			}
		}
	}

	return false
}

// insertElem inserts a new event into the sorted slice at the specified position.
// Returns a new slice with the event inserted. Method avoid additional allocations.
func (s *Storage) insertElem(arr []*types.Event, elem *types.Event, pos int) []*types.Event {
	arr = append(arr, nil)
	copy(arr[pos+1:], arr[pos:])
	arr[pos] = elem
	return arr
}

// deleteElem removes an event from the sorted slice at the specified position.
// Returns a new slice with the event removed. Method avoid additional allocations.
func (s *Storage) deleteElem(arr []*types.Event, pos int) []*types.Event {
	copy(arr[pos:], arr[pos+1:])
	arr[len(arr)-1] = nil
	arr = arr[:len(arr)-1]
	return arr
}

// getIndex returns the index of the event in the sorted slice.
func (s *Storage) getIndex(arr []*types.Event, elem *types.Event) int {
	return sort.Search(len(arr), func(i int) bool {
		if arr[i].Datetime.After(elem.Datetime) {
			return true
		}
		if arr[i].Datetime.Equal(elem.Datetime) {
			return arr[i].ID.String() >= elem.ID.String()
		}
		return false
	})
}
