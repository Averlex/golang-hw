package hw04lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func emptyListDataStructureTest(t *testing.T) {
	t.Helper()
	t.Run("data structure", func(t *testing.T) {
		l := NewList()

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})
}

func emptyListPushTests(t *testing.T) {
	t.Helper()

	testVal := 42

	testCases := []struct {
		name   string
		len    int
		method func(List, any) *ListItem
	}{
		{"push front", 1, List.PushFront},
		{"push back", 1, List.PushBack},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			l := NewList()
			newElem := tC.method(l, testVal)

			require.Equal(t, tC.len, l.Len())
			require.Equal(t, l.Front(), newElem)
			require.Equal(t, l.Back(), newElem)
		})
	}
}

func emptyListRemoveTest(t *testing.T) {
	t.Helper()
	t.Run("remove", func(t *testing.T) {
		l := NewList()
		l.Remove(nil)

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})
}

func emptyListMoveToFrontTest(t *testing.T) {
	t.Helper()
	t.Run("move to front", func(t *testing.T) {
		testVal := 42
		l := NewList()
		elem := &ListItem{Value: testVal}
		l.MoveToFront(elem)

		require.Equal(t, 1, l.Len())
		require.Equal(t, l.Front(), l.Back())
		require.Equal(t, l.Front(), elem)
		require.Nil(t, l.Front().Prev)
		require.Nil(t, l.Back().Next)
	})
}

func emptyListTests(t *testing.T) {
	t.Helper()

	emptyListDataStructureTest(t)
	emptyListPushTests(t)
	emptyListRemoveTest(t)
	emptyListMoveToFrontTest(t)
}

func singleElemListDataStructureTest(t *testing.T) {
	t.Helper()
	t.Run("data structure", func(t *testing.T) {
		testVal := 42
		l := NewList()
		firstElem := l.PushFront(testVal)

		require.Equal(t, 1, l.Len())
		require.Equal(t, l.Front(), l.Back())
		require.Equal(t, l.Front(), firstElem)
	})
}

func singleElemListPushTests(t *testing.T) {
	t.Helper()

	testVal := 42

	testCases := []struct {
		name                   string
		len                    int
		method                 func(List, any) *ListItem
		headMethod, tailMethod func(List) *ListItem
	}{
		{"push front", 2, List.PushFront, List.Front, List.Back},
		{"push back", 2, List.PushBack, List.Back, List.Front},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			l := NewList()
			firstElem := tC.method(l, testVal)
			newElem := tC.method(l, testVal)

			// Data structure testing.
			require.Equal(t, tC.len, l.Len())
			require.Equal(t, tC.headMethod(l), newElem)
			require.Equal(t, tC.tailMethod(l), firstElem)

			// Linkage testing.
			require.Equal(t, l.Front().Next, l.Back())
			require.Equal(t, l.Back().Prev, l.Front())
		})
	}
}

func singleElemListRemoveTest(t *testing.T) {
	t.Helper()
	t.Run("remove", func(t *testing.T) {
		testVal := 42
		l := NewList()
		elem := l.PushFront(testVal)
		l.Remove(elem)

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})
}

func singleElemListMoveToFrontTest(t *testing.T) {
	t.Helper()
	t.Run("move to front", func(t *testing.T) {
		testVal := 42
		l := NewList()
		elem := l.PushFront(testVal)
		l.MoveToFront(elem)

		require.Equal(t, 1, l.Len())
		require.Equal(t, l.Front(), l.Back())
		require.Equal(t, l.Front(), elem)
		require.Nil(t, l.Front().Prev)
		require.Nil(t, l.Back().Next)
	})
}

func singleElemListTests(t *testing.T) {
	t.Helper()

	singleElemListDataStructureTest(t)
	singleElemListPushTests(t)
	singleElemListRemoveTest(t)
	singleElemListMoveToFrontTest(t)
}

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) { emptyListTests(t) })
	t.Run("list with a single element", func(t *testing.T) { singleElemListTests(t) })

	// t.Run("complex", func(t *testing.T) {
	// 	l := NewList()

	// 	l.PushFront(10) // [10]
	// 	l.PushBack(20)  // [10, 20]
	// 	l.PushBack(30)  // [10, 20, 30]
	// 	require.Equal(t, 3, l.Len())

	// 	middle := l.Front().Next // 20
	// 	l.Remove(middle)         // [10, 30]
	// 	require.Equal(t, 2, l.Len())

	// 	for i, v := range [...]int{40, 50, 60, 70, 80} {
	// 		if i%2 == 0 {
	// 			l.PushFront(v)
	// 		} else {
	// 			l.PushBack(v)
	// 		}
	// 	} // [80, 60, 40, 10, 30, 50, 70]

	// 	require.Equal(t, 7, l.Len())
	// 	require.Equal(t, 80, l.Front().Value)
	// 	require.Equal(t, 70, l.Back().Value)

	// 	l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
	// 	l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

	// 	elems := make([]int, 0, l.Len())
	// 	for i := l.Front(); i != nil; i = i.Next {
	// 		elems = append(elems, i.Value.(int))
	// 	}
	// 	require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	// })
}
