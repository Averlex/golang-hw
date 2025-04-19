// Package hw05parallelexecution implements Run function - a custom worker pool implementation.
package hw05parallelexecution

import (
	"errors"
)

// ErrErrorsLimitExceeded is an error indicating that the limit of errors is exceeded.
var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

// Task is a function that is executed in a separate worker.
type Task func() error

// Реализовать воркер с 3 каналами: внешний стоп, канал задач, канал done
// Реализовать мультиплексор каналов And-channel с лимитом m

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if len(tasks) == 0 || n <= 0 {
		return nil
	}

	// Place your code here.
	return nil
}
