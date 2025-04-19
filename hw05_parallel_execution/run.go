// Package hw05parallelexecution implements Run function - a custom worker pool implementation.
package hw05parallelexecution

import (
	"errors"
	"fmt"
	"sync"
)

// ErrErrorsLimitExceeded is an error indicating that the limit of errors is exceeded.
var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

// Task is a function that is executed in a separate worker.
type Task func() error

// Реализовать воркер с 3 каналами: внешний стоп, канал задач, канал done
// Реализовать мультиплексор каналов And-channel с лимитом m - в отдельной горутине

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if len(tasks) == 0 || n <= 0 {
		return nil
	}

	// stop := make(chan struct{})
	// errors := make([]chan error, n)
	// taskPool := make(chan Task)

	// runWorkers := func() {
	// 	for i := 0; i < n; i++ {
	// 		// Initializing worker response channel.
	// 		errors[i] := make(chan error)
	// 		go func(res chan error) {

	// 		}(errors[i])
	// 	}
	// }

	return nil
}

func worker(wg *sync.WaitGroup, stop <-chan struct{}, taskPool <-chan Task, taskRes chan<- error) {
	defer close(taskRes)
	defer wg.Done()

	defer func() {
		if r := recover(); r != nil {
			taskRes <- fmt.Errorf("one of the tasks panicked: %v", r)
		}
		close(taskRes)
	}()

	for {
		// Prioritizing stop signals.
		select {
		case <-stop:
			return
		case task, ok := <-taskPool:
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
