// Package hw05parallelexecution implements Run function - a custom worker pool implementation.
package hw05parallelexecution

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
)

var (
	// ErrErrorsLimitExceeded is an error indicating that the limit of errors is exceeded.
	ErrErrorsLimitExceeded = errors.New("errors limit exceeded")
	// ErrNoWorkers is an error indicating that no workers were set.
	ErrNoWorkers = errors.New("no workers set")
	// ErrNoTasks is an error indicating that no tasks were set.
	ErrNoTasks = errors.New("no tasks set")
)

// Task is a function that is executed in a separate worker.
type Task func() error

type errorCounter struct {
	count int64
	m     int64
}

// Inc atomically increments the error counter and returns its new value.
func (ec *errorCounter) Inc() int64 {
	return atomic.AddInt64(&ec.count, 1)
}

// IsExceeded checks if the number of errors exceeds the limit.
// Returns false if the number of errors is less than limit, true otherwise.
func (ec *errorCounter) IsExceeded() bool {
	if ec.m <= 0 {
		return false
	}
	return atomic.LoadInt64(&ec.count) >= ec.m
}

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if len(tasks) == 0 {
		return ErrNoTasks
	}

	if n <= 0 {
		return ErrNoWorkers
	}

	// n workers + current goroutine + tasks sender.
	runtime.GOMAXPROCS(n + 2)

	wg := &sync.WaitGroup{}
	taskPool := make(chan Task)
	errCounter := &errorCounter{count: 0, m: int64(m)}

	// Starting workers.
	for range n {
		wg.Add(1)
		go worker(wg, taskPool, errCounter)
	}

	// Sending tasks.
	wg.Add(1)
	go sendTasks(wg, taskPool, errCounter, tasks)

	wg.Wait()

	if errCounter.IsExceeded() {
		return ErrErrorsLimitExceeded
	}

	return nil
}

// worker takes tasks from the taskPool and executes them.
// The worker continues to work after any of the tasks panics.
// If the number of errors exceeds the limit, the worker stops working.
func worker(wg *sync.WaitGroup, taskPool <-chan Task, errCounter *errorCounter) {
	defer wg.Done()
	for task := range taskPool {
		// Ensuring that the worker continues to work after any of the tasks panics.
		func() {
			defer func() {
				if r := recover(); r != nil {
					errCounter.Inc()
				}
			}()
			if errCounter.IsExceeded() {
				return
			}
			if res := task(); res != nil {
				errCounter.Inc()
			}
		}()
		if errCounter.IsExceeded() {
			return
		}
	}
}

// sendTasks sends tasks to the taskPool and stops when all tasks have been sent or the error limit has been exceeded.
func sendTasks(wg *sync.WaitGroup, taskPool chan<- Task, errCounter *errorCounter, tasks []Task) {
	defer wg.Done()
	defer close(taskPool)
	sentCount := 0
	for sentCount < len(tasks) {
		if errCounter.IsExceeded() {
			return
		}
		select {
		case taskPool <- tasks[sentCount]:
			sentCount++
		default:
		}
	}
}
