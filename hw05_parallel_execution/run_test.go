package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	//nolint:depguard
	"go.uber.org/goleak"
)

// concurrency test without time.Sleep

func makeTasks(n int, isError, isPanic bool) []Task {
	tasks := make([]Task, 0, n)
	for range n {
		switch {
		case isError:
			tasks = append(tasks, func() error { return fmt.Errorf("error") })
		case isPanic:
			tasks = append(tasks, func() error { panic("panic") })
		default:
			tasks = append(tasks, func() error { return nil })
		}
	}
	return tasks
}

func formText(val int, text string) string {
	if val == 1 {
		return fmt.Sprintf("%d %s", val, text)
	}
	return fmt.Sprintf("%d %ss", val, text)
}

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})
}

func TestRun_IncorrectParams(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name        string
		tasks       []Task
		n           int
		expectedErr error
	}{
		{"nil tasks slice", nil, 1, ErrNoTasks},
		{"n < 0", make([]Task, 1), -5, ErrNoWorkers},
		{"n = 0", make([]Task, 1), 0, ErrNoWorkers},
		{"nil task", []Task{nil}, 1, nil},
		{"panicing task", []Task{func() error { panic(42) }}, 1, nil},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			err := Run(tC.tasks, tC.n, 42)
			require.True(t, errors.Is(err, tC.expectedErr), "actual err - %v", err)
		})
	}
}

func TestRun_TasksWorkersMux(t *testing.T) {
	t.Helper()

	workerNum := []int{0, 1, 2, 10, 1000}
	taskNum := []int{0, 1, 2, 100, 10000}

	for _, wN := range workerNum {
		t.Run(formText(wN, "worker"), func(t *testing.T) {
			for _, tN := range taskNum {
				tasks := makeTasks(tN, false, false)

				t.Run(formText(tN, "task"), func(t *testing.T) {
					err := Run(tasks, wN, wN)
					switch {
					case tN == 0:
						require.True(t, errors.Is(err, ErrNoTasks), "actual err - %v", err)
					case wN == 0:
						require.True(t, errors.Is(err, ErrNoWorkers), "actual err - %v", err)
					default:
						require.NoError(t, Run(tasks, wN, wN))
					}
				})
			}
		})
	}
}

func TestRun_ErrorParameter(t *testing.T) {
	t.Helper()

	testCases := []struct {
		name        string
		tasks       []Task
		n           int
		m           int
		expectedErr error
	}{
		{"m<0, n=10, all errors", makeTasks(10, true, false), 10, -1, nil},
		{"m=0, n=10, all errors", makeTasks(10, false, true), 10, 0, nil},
		{"m=1, n=1, all errors", makeTasks(1, true, false), 1, 1, ErrErrorsLimitExceeded},
		{"m=1, n=2, all errors", makeTasks(2, false, true), 2, 1, ErrErrorsLimitExceeded},
		{"m=2, n=2, all errors", makeTasks(2, true, false), 2, 2, ErrErrorsLimitExceeded},
		{"m=5, n=10, all errors", makeTasks(10, false, true), 10, 5, ErrErrorsLimitExceeded},
		{"m=10, n=5, all errors", makeTasks(5, true, false), 5, 10, nil},

		{"m<0, n=10, no errors", makeTasks(10, false, false), 10, -1, nil},
		{"m=0, n=10, no errors", makeTasks(10, false, false), 10, 0, nil},
		{"m=1, n=1, no errors", makeTasks(1, false, false), 1, 1, nil},
		{"m=1, n=2, no errors", makeTasks(2, false, false), 2, 1, nil},
		{"m=2, n=2, no errors", makeTasks(2, false, false), 2, 2, nil},
		{"m=5, n=10, no errors", makeTasks(10, false, false), 10, 5, nil},
		{"m=10, n=5, no errors", makeTasks(5, false, false), 5, 10, nil},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			err := Run(tC.tasks, tC.n, tC.m)
			if tC.expectedErr == nil {
				require.NoError(t, err, "actual err - %v", err)
				return
			}
			require.ErrorIs(t, err, tC.expectedErr, "actual err - %v", err)
		})
	}
}

func TestRun_ConcurrentExecution(t *testing.T) {
	var activeGoroutines int32
	var maxActiveGoroutines int32
	var mu sync.Mutex

	tasks := make([]Task, 10)
	for i := 0; i < 10; i++ {
		tasks[i] = func() error {
			current := atomic.AddInt32(&activeGoroutines, 1)

			mu.Lock()
			if current > maxActiveGoroutines {
				maxActiveGoroutines = current
			}
			mu.Unlock()

			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))

			atomic.AddInt32(&activeGoroutines, -1)
			return nil
		}
	}

	m := 1
	testCases := []int{1, 2, 5, 10, 1000}

	for _, n := range testCases {
		t.Run(formText(n, "goroutine"), func(t *testing.T) {
			err := Run(tasks, n, m)
			require.NoError(t, err, "actual err - %v", err)
			require.LessOrEqual(t, maxActiveGoroutines, int32(n),
				"max concurrent goroutines: %d, expected less or equal: %d", maxActiveGoroutines, n)
			require.Eventually(t, func() bool {
				return atomic.LoadInt32(&activeGoroutines) == 0
			}, time.Second, time.Millisecond*100, "%d goroutines are still active", activeGoroutines)
		})
	}
}
