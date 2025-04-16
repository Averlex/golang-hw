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
		// Comparing values since we can't compare pointers due to method's behavior.
		require.Equal(t, l.Front().Value, elem.Value)
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
		require.Equal(t, l.Front().Value, elem.Value)
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

func twoElemListDataStructureTest(t *testing.T) {
	t.Helper()
	t.Run("data structure", func(t *testing.T) {
		l := NewList()
		testVal := 42
		firstElem := l.PushFront(testVal)
		secondElem := l.PushBack(testVal)

		require.Equal(t, 2, l.Len())
		require.Equal(t, firstElem, l.Front())
		require.Equal(t, secondElem, l.Back())
		require.Equal(t, l.Front().Next, l.Back())
		require.Equal(t, l.Back().Prev, l.Front())
		require.Nil(t, l.Front().Prev)
		require.Nil(t, l.Back().Next)
	})
}

func twoElemListPushTests(t *testing.T) {
	t.Helper()

	testVal := 42

	testCases := []struct {
		name                       string
		len                        int
		method                     func(List, any) *ListItem
		expectedHead, expectedTail *ListItem
	}{
		{"push front", 3, List.PushFront, nil, nil},
		{"push back", 3, List.PushBack, nil, nil},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			l := NewList()
			firstElem := l.PushFront(testVal)
			secondElem := l.PushBack(testVal)
			newElem := tC.method(l, testVal)

			require.Equal(t, tC.len, l.Len())

			if tC.name == "push front" {
				require.Equal(t, newElem, l.Front())
				require.Equal(t, secondElem, l.Back())
			} else {
				require.Equal(t, newElem, l.Back())
				require.Equal(t, firstElem, l.Front())
			}

			// Проверка связей между элементами
			require.Equal(t, l.Front().Next, l.Back().Prev)
			require.Equal(t, l.Front().Next.Next, l.Back())
			require.Equal(t, l.Back().Prev.Prev, l.Front())
		})
	}
}

func twoElemListRemoveTest(t *testing.T) {
	t.Helper()

	testVal := 42

	t.Run("remove first element", func(t *testing.T) {
		l := NewList()
		firstElem := l.PushFront(testVal)
		secondElem := l.PushBack(testVal)
		l.Remove(firstElem)

		require.Equal(t, 1, l.Len())
		require.Equal(t, secondElem, l.Front())
		require.Equal(t, secondElem, l.Back())
		require.Nil(t, l.Front().Prev)
		require.Nil(t, l.Back().Next)
	})

	t.Run("remove last element", func(t *testing.T) {
		l := NewList()
		firstElem := l.PushFront(testVal)
		secondElem := l.PushBack(testVal)
		l.Remove(secondElem)

		require.Equal(t, 1, l.Len())
		require.Equal(t, firstElem, l.Front())
		require.Equal(t, firstElem, l.Back())
		require.Nil(t, l.Front().Prev)
		require.Nil(t, l.Back().Next)
	})
}

func twoElemListMoveToFrontTest(t *testing.T) {
	t.Helper()

	testVal := 42

	t.Run("move first element to front", func(t *testing.T) {
		l := NewList()
		firstElem := l.PushFront(testVal)
		secondElem := l.PushBack(testVal)
		l.MoveToFront(firstElem)

		require.Equal(t, 2, l.Len())
		// Comparing values since we can't compare pointers due to method's behavior.
		require.Equal(t, firstElem.Value, l.Front().Value)
		require.Equal(t, secondElem, l.Back())
		require.Equal(t, l.Front().Next, l.Back())
		require.Equal(t, l.Back().Prev, l.Front())
	})

	t.Run("move last element to front", func(t *testing.T) {
		l := NewList()
		firstElem := l.PushFront(testVal * 1)
		secondElem := l.PushBack(testVal * 2)
		l.MoveToFront(secondElem)

		require.Equal(t, 2, l.Len())
		// Comparing values since we can't compare pointers due to method's behavior.
		require.Equal(t, secondElem.Value, l.Front().Value)
		require.Equal(t, firstElem, l.Back())
		require.Equal(t, l.Front().Next, l.Back())
		require.Equal(t, l.Back().Prev, l.Front())
		require.Nil(t, l.Back().Next)
		require.Nil(t, l.Front().Prev)
	})
}

func twoElemListTests(t *testing.T) {
	t.Helper()

	twoElemListDataStructureTest(t)
	twoElemListPushTests(t)
	twoElemListRemoveTest(t)
	twoElemListMoveToFrontTest(t)
}

func complexListTests(t *testing.T) {
	t.Helper()

	t.Run("behavioral", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})

	t.Run("cyclic moving last elem to front", func(t *testing.T) {
		l := NewList()

		cycleLen := 10
		expected := make([]int, 0, cycleLen)

		getList := func(l List) []int {
			elems := make([]int, 0, cycleLen)
			for i := l.Front(); i != nil; i = i.Next {
				elems = append(elems, i.Value.(int))
			}
			return elems
		}

		// expected: [0, 10, 20, 30, 40, 50, 60, 70, 80, 90]
		for i := range cycleLen {
			l.PushBack(i * 10)
			expected = append(expected, i*10)
		}

		require.Equal(t, cycleLen, l.Len())
		require.Equal(t, expected, getList(l))

		// Clearing
		for range cycleLen {
			l.MoveToFront(l.Back())
		}

		require.Equal(t, expected, getList(l))
	})

	t.Run("list clearing", func(t *testing.T) {
		l := NewList()

		cycleLen := 10
		expected := make([]int, 0, cycleLen)

		getList := func(l List) []int {
			elems := make([]int, 0, cycleLen)
			for i := l.Front(); i != nil; i = i.Next {
				elems = append(elems, i.Value.(int))
			}
			return elems
		}

		// expected: [0, 10, 20, 30, 40, 50, 60, 70, 80, 90]
		for i := range cycleLen {
			l.PushBack(i * 10)
			expected = append(expected, i*10)
		}

		require.Equal(t, cycleLen, l.Len())
		require.Equal(t, expected, getList(l))

		// Ciclyc shift
		for range cycleLen {
			l.Remove(l.Back())
		}

		require.Equal(t, 0, l.Len())
	})
}

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) { emptyListTests(t) })
	t.Run("list with a single element", func(t *testing.T) { singleElemListTests(t) })
	t.Run("list with 2 elements", func(t *testing.T) { twoElemListTests(t) })
	t.Run("complex", func(t *testing.T) { complexListTests(t) })
}
