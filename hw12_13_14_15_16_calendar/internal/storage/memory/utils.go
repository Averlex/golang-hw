package memory

import (
	"sort"

	projectErrors "github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/errors" //nolint:depguard,nolintlint
	"github.com/Averlex/golang-hw/hw12_13_14_15_16_calendar/pkg/types"                //nolint:depguard,nolintlint
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

// checkState checks if the storage is initialized and ready for operations.
// Returns ErrStorageUninitialized on failure.
func (s *Storage) checkState() error {
	if s.events == nil || s.userIndex == nil || s.idIndex == nil {
		return projectErrors.ErrStorageUninitialized
	}
	return nil
}

func (s *Storage) isOverlaps(arr []*types.Event, elem *types.Event) int {
	// Нужно проверить, что новый элемент не пересекается с уже существующими.
	// В том числе исключить кейс, когда он пересекается сам с собой.
	// Подумать на тему того, что кэш по id с индексом будет лучше, чем с указателем.

	pos := s.findInsertPosition(arr, elem)
	switch {
		case pos == 0 && len(arr) != 0:
			if arr[0].Datetime.Before(elem.Datetime.Add(elem.Duration)) {
				return -1
			}
		case pos == len(arr) && len(arr) != 0:
			if arr[len(arr)-1].Datetime.Before(elem.Datetime.Add(elem.Duration)) {
				return pos - 1
			}
		case pos
		default:
			if arr[pos-1].Datetime.Add(arr[pos-1].Duration).After(elem.Datetime) ||
				arr[pos].Datetime.Before(elem.Datetime.Add(elem.Duration)) {
				return pos - 1
			}
	}
	if pos == 0 && len(arr) != 0 {
		if arr[1].Datetime.Add(arr[1].Duration).After(elem.Datetime) {
			return pos
		}
	}
}
