// Package hw05parallelexecution implements Run function - a custom worker pool implementation.
package hw05parallelexecution

import (
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrErrorsLimitExceeded is an error indicating that the limit of errors is exceeded.
	ErrErrorsLimitExceeded = errors.New("errors limit exceeded")
	// ErrNoWorkers is an error indicating that no workers were set.
	ErrNoWorkers = errors.New("no workers set")
)

// Task is a function that is executed in a separate worker.
type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if len(tasks) == 0 {
		return nil
	}

	if n <= 0 {
		return ErrNoWorkers
	}

	wg := &sync.WaitGroup{}
	stop := make(chan struct{})
	taskPool := taskGenerator(wg, tasks, stop)
	taskResults := runWorkers(wg, stop, taskPool, n)
	muxedResults := muxChannels(wg, taskResults...)
	res := processTaskResults(muxedResults, m)

	// if res != nil {
	// 	select {
	// 	case stop <- struct{}{}: // Sending a singnal if at least 1 of listeners up.
	// 	default: // All workers are already done.
	// 	}
	// }

	if res != nil {
		close(stop)
	} else {
		defer close(stop)
	}

	wg.Wait()

	return res
}

// taskGenerator creates a goroutine that sends each task in the tasks slice
// to the returned channel. If the stop channel is closed or the stop signal received,
// the goroutine will stop sending tasks and close the channel.
func taskGenerator(wg *sync.WaitGroup, tasks []Task, stop <-chan struct{}) <-chan Task {
	taskPool := make(chan Task)
	wg.Add(1)

	go func() {
		defer close(taskPool)
		defer wg.Done()
		for _, task := range tasks {
			select {
			case <-stop:
				return
			default:
				select {
				case <-stop:
					return
				case taskPool <- task:
				}
			}
		}
	}()

	return taskPool
}

// worker starts a goroutine that consumes tasks from taskPool and sends the results to taskRes.
// It stops working when the stop signal is received or when there are no more tasks.
// It handles panics in tasks and sends the error to the taskRes channel.
func worker(wg *sync.WaitGroup, stop <-chan struct{}, taskPool <-chan Task, taskRes chan<- error) {
	if taskRes != nil {
		defer close(taskRes)
	}
	defer wg.Done()

	if stop == nil || taskPool == nil {
		return
	}

	for {
		select {
		case <-stop:
			return
		default:
			task, ok := <-taskPool
			// No tasks left.
			if !ok {
				return
			}

			// Ensuring that the worker continues to work after any of the tasks panics.
			func() {
				defer func() {
					if r := recover(); r != nil {
						taskRes <- fmt.Errorf("one of the tasks panicked: %v", r)
					}
				}()
				taskRes <- task()
			}()
		}
	}
}

// muxChannels starts a goroutine that listens to all given channels and multiplexes them to a single one.
// It closes the result channel when all input channels are closed.
func muxChannels(wg *sync.WaitGroup, channels ...<-chan error) <-chan error {
	if len(channels) == 0 {
		return nil
	}

	res := make(chan error)
	closedChannels := make(map[int]struct{}, len(channels))
	wg.Add(1)

	// Starting separate goroutine to listen to all channels and multiplexing them to a single one.
	go func() {
		defer wg.Done()
		defer close(res)

		for len(closedChannels) < len(channels) {
			// Listening to each channel once without blocking.
			for i, ch := range channels {
				if _, ok := closedChannels[i]; ok {
					continue
				}
				select {
				case v, ok := <-ch:
					if !ok {
						closedChannels[i] = struct{}{}
						continue
					}
					res <- v
				default:
					continue
				}
			}
		}
	}()

	return res
}

// runWorkers starts n goroutines that consume tasks from taskPool and send the results to the returned channel.
// It stops working when the stop signal is received or when there are no more tasks.
func runWorkers(wg *sync.WaitGroup, stop <-chan struct{}, taskPool <-chan Task, n int) []<-chan error {
	res := make([]<-chan error, n)

	for i := 0; i < n; i++ {
		workerRes := make(chan error) // Initializing worker response channel.
		wg.Add(1)
		go worker(wg, stop, taskPool, workerRes)
		res[i] = workerRes // Casting workerRes to <-workerRes.
	}

	return res
}

// processTaskResults processes the results from the provided error channel res.
// It listens to the channel until it is closed, counting the number of errors received.
//
// If m is greater than 0, it treats m as the maximum allowable error count.
//
// If the count of errors reaches or exceeds m, it returns ErrErrorsLimitExceeded.
//
// If m is 0 or negative, it ignores the error count and processes all results until the channel is closed.
func processTaskResults(receiver <-chan error, m int) error {
	ignoreErrors := m <= 0
	var res error

	counter := 0

	for v := range receiver {
		// Task completed without errors.
		if v == nil {
			continue
		}

		counter++
		// Limit of errors exceeded.
		if !ignoreErrors && counter >= m {
			res = ErrErrorsLimitExceeded
		}
	}

	return res
}
