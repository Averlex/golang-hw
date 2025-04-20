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

// +++ Реализовать воркер с 3 каналами: внешний стоп, канал задач, канал done
// +++ Реализовать передачу задач в канал
// +++ Убедиться, что генератор безопасен
// +++ Реализовать мультиплексор каналов
// Обработать лимит m
// m <= 0 - не использовать лимит ошибок

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if len(tasks) == 0 || n <= 0 {
		return nil
	}

	wg := &sync.WaitGroup{}

	stop := make(chan struct{})
	errors := make([]chan error, n)

	taskPool := taskGenerator(tasks, stop)

	runWorkers := func() {
		// Running n workers regardless of the len(tasks).
		for i := 0; i < n; i++ {
			errors[i] = make(chan error) // Initializing worker response channel.
			wg.Add(1)
			go worker(wg, stop, taskPool, errors[i])
		}
	}

	return nil
}

// taskGenerator creates a goroutine that sends each task in the tasks slice
// to the returned channel. If the stop channel is closed or the stop signal received,
// the goroutine will stop sending tasks and close the channel.
func taskGenerator(tasks []Task, stop <-chan struct{}) <-chan Task {
	taskPool := make(chan Task)

	go func() {
		defer close(taskPool)
		for _, task := range tasks {
			select {
			case <-stop:
				return
			case taskPool <- task:
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

		for len(closedChannels) == len(channels) {
			// Listening to each channel once without blocking.
			for i, ch := range channels {
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
