package hw05parallelexecution

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const testTimeout = 100 * time.Millisecond

func taskText(val int) string {
	if val == 1 {
		return fmt.Sprintf("%d task", val)
	}
	return fmt.Sprintf("%d tasks", val)
}

func TestInternal(t *testing.T) {
	t.Run("taskGenerator", generatorTests)
	t.Run("worker", workerTests)
}

type generatorSuite struct {
	suite.Suite
	wg        *sync.WaitGroup
	suiteName string
	n         int
	tasks     []Task
	stop      chan struct{}
	taskPool  <-chan Task
}

func newGeneratorSuite(suiteName string, n int) *generatorSuite {
	return &generatorSuite{
		suiteName: suiteName,
		n:         n,
	}
}

func generatorTests(t *testing.T) {
	t.Helper()

	testCases := []int{1, 2, 5, 10, 1000}

	for _, tC := range testCases {
		tc := tC
		t.Run(taskText(tc), func(t *testing.T) {
			suite.Run(t, newGeneratorSuite(taskText(tc), tc))
		})
	}
}

func (s *generatorSuite) SetupTest() {
	s.wg = &sync.WaitGroup{}
	s.stop = make(chan struct{})

	s.tasks = make([]Task, s.n)
	for i := range s.tasks {
		s.tasks[i] = func() error { return nil }
	}

	s.taskPool = taskGenerator(s.wg, s.tasks, s.stop)
}

func (s *generatorSuite) TearDownTest() {
	s.wg.Wait()
	select {
	case _, ok := <-s.stop:
		if ok {
			close(s.stop)
		}
	default:
		close(s.stop)
	}
}

func (s *generatorSuite) TestBuildingPool() {
	for i := 0; i < s.n; i++ {
		select {
		case _, ok := <-s.taskPool:
			s.Require().True(ok, "taskPool closed prematurely (got %d/%d tasks)", s.suiteName, i+1, s.n)
		case <-time.After(testTimeout):
			s.Require().Fail("test timeout exceeded")
			return
		}
	}

	select {
	case _, ok := <-s.taskPool:
		if ok {
			s.Require().Fail("taskPool should be closed but isn't", s.suiteName)
		}
		s.Require().False(ok, "taskPool should be closed after %d tasks", s.suiteName, s.n)
	case <-time.After(testTimeout):
		s.Require().Fail("test timeout exceeded")
		return
	}
}

func (s *generatorSuite) TestStopBeforeExctracting() {
	close(s.stop)

	select {
	case _, ok := <-s.taskPool:
		s.Require().False(ok, "taskPool should be closed before sending tasks", s.suiteName)
	case <-time.After(testTimeout):
		s.Require().Fail("test timeout exceeded")
		return
	}
}

func (s *generatorSuite) TestStopDuringExctracting() {
	counter := 0
	stopValue := s.n / 2

	for ; counter < s.n; counter++ {
		if counter == stopValue {
			close(s.stop)
			break
		}
		select {
		case _, ok := <-s.taskPool:
			s.Require().True(ok, "taskPool closed prematurely (got %d/%d tasks)", s.suiteName, counter+1, stopValue)
		case <-time.After(testTimeout):
			s.Require().Fail("test timeout exceeded")
			return
		}
	}

	s.Require().Equal(counter, stopValue)

	select {
	case _, ok := <-s.taskPool:
		if ok {
			s.Require().Fail("taskPool should be closed but isn't", s.suiteName)
		}
		s.Require().False(ok, "taskPool should be closed after %d tasks", s.suiteName, stopValue)
	case <-time.After(testTimeout):
		s.Require().Fail("test timeout exceeded")
		return
	}
}

func (s *generatorSuite) TestStopAfterExctracting() {
	counter := 0
	stopValue := s.n

	for ; counter < s.n; counter++ {
		select {
		case _, ok := <-s.taskPool:
			s.Require().True(ok, "taskPool closed prematurely (got %d/%d tasks)", s.suiteName, counter+1, s.n)
		case <-time.After(testTimeout):
			s.Require().Fail("test timeout exceeded")
			return
		}
	}

	close(s.stop)

	s.Require().Equal(counter, stopValue)

	select {
	case _, ok := <-s.taskPool:
		if ok {
			s.Require().Fail("taskPool should be closed but isn't", s.suiteName)
		}
		s.Require().False(ok, "taskPool should be closed after %d tasks", s.suiteName, stopValue)
	case <-time.After(testTimeout):
		s.Require().Fail("test timeout exceeded")
		return
	}
}

type workerSuite struct {
	suite.Suite
	wg        *sync.WaitGroup
	suiteName string
	n         int
	isError   bool
	isPanic   bool
	tasks     []Task
	stop      chan struct{}
	taskPool  <-chan Task
	taskRes   chan error
}

func newWorkerSuite(suiteName string, n int, isError, isPanic bool) *workerSuite {
	return &workerSuite{
		suiteName: suiteName,
		n:         n,
		isError:   isError,
		isPanic:   isPanic,
	}
}

func workerTests(t *testing.T) {
	t.Helper()

	testCases := []int{1, 2, 5, 10, 1000}

	for _, tC := range testCases {
		tc := tC

		t.Run(taskText(tc), func(t *testing.T) {
			t.Run("normal", func(t *testing.T) {
				suite.Run(t, newWorkerSuite(taskText(tc)+"/normal", tc, false, false))
			})
			t.Run("with error", func(t *testing.T) {
				suite.Run(t, newWorkerSuite(taskText(tc)+"/with error", tc, true, false))
			})
			t.Run("with panic", func(t *testing.T) {
				suite.Run(t, newWorkerSuite(taskText(tc)+"/with panic", tc, false, true))
			})
		})
	}
}

func (s *workerSuite) SetupTest() {
	s.wg = &sync.WaitGroup{}
	s.stop = make(chan struct{})
	s.taskRes = make(chan error)

	s.tasks = make([]Task, s.n)
	for i := range s.tasks {
		switch {
		case s.isPanic:
			s.tasks[i] = func() error { panic("panic") }
		case s.isError:
			s.tasks[i] = func() error { return fmt.Errorf("error") }
		default:
			s.tasks[i] = func() error { return nil }
		}
	}

	s.taskPool = taskGenerator(s.wg, s.tasks, s.stop)
}

func (s *workerSuite) TearDownTest() {
	s.wg.Wait()
	select {
	case _, ok := <-s.stop:
		if ok {
			close(s.stop)
		}
	default:
		close(s.stop)
	}
}

func (s *workerSuite) TestWorker() {
	erroneousCond := s.isPanic || s.isError

	s.wg.Add(1)
	go worker(s.wg, s.stop, s.taskPool, s.taskRes)

	for i := 0; i < s.n; i++ {
		select {
		case v, ok := <-s.taskRes:
			s.Require().True(ok, "taskRes closed prematurely (got %d/%d tasks)", s.suiteName, i+1, s.n)
			s.Require().Equal(erroneousCond, v != nil,
				`taskRes received incorrect value on task %d/%d:
			- got %v
			- isPanic=%v
			- isError=%v`,
				s.suiteName, i+1, s.n, v, s.isPanic, s.isError)
		case <-time.After(testTimeout):
			s.Require().Fail("test timeout exceeded")
			return
		}
	}

	select {
	case _, ok := <-s.taskRes:
		if ok {
			s.Require().Fail("taskRes should be closed but isn't", s.suiteName)
		}
		s.Require().False(ok, "taskRes should be closed after receiving %d tasks", s.suiteName, s.n)
	case <-time.After(testTimeout):
		s.Require().Fail("test timeout exceeded")
		return
	}
}

func (s *workerSuite) TestStopBeforeExecuting() {
	close(s.stop)
	counter := 0

	s.wg.Add(1)
	go worker(s.wg, s.stop, s.taskPool, s.taskRes)

ForLoop:
	for ; counter < s.n; counter++ {
		select {
		case _, ok := <-s.taskRes:
			if !ok {
				break ForLoop
			}
		case <-time.After(testTimeout):
			s.Require().Fail("test timeout exceeded")
			return
		}
	}

	s.Require().Equal(counter, 0, "tasks done before stop - %v, expected - %v", s.suiteName, counter, 0)
	select {
	case _, ok := <-s.taskRes:
		s.Require().False(ok, "taskRes should be closed before sending tasks", s.suiteName)
	case <-time.After(testTimeout):
		s.Require().Fail("test timeout exceeded")
		return
	}
}

func (s *workerSuite) TestStopDuringExecuting() {
	counter := 0
	stopValue := s.n / 2

	s.wg.Add(1)
	go worker(s.wg, s.stop, s.taskPool, s.taskRes)

ForLoop:
	for ; counter < s.n; counter++ {
		if counter == stopValue {
			close(s.stop)
		}
		select {
		case _, ok := <-s.taskRes:
			if !ok {
				break ForLoop
			}
		case <-time.After(testTimeout):
			s.Require().Fail("test timeout exceeded")
			return
		}
	}

	// Might have done stopValue and stopValue+1 tasks as well.
	s.Require().LessOrEqual(counter, stopValue+1)

	select {
	case _, ok := <-s.taskRes:
		if ok {
			s.Require().Fail("taskRes should be closed but isn't", s.suiteName)
		}
		s.Require().False(ok, "taskRes should be closed after %d tasks", s.suiteName, stopValue)
	case <-time.After(testTimeout):
		s.Require().Fail("taskRes should be closed but isn't", s.suiteName)
	}
}

func (s *workerSuite) TestStopAfterExecuting() {
	counter := 0

	s.wg.Add(1)
	go worker(s.wg, s.stop, s.taskPool, s.taskRes)

ForLoop:
	for ; counter < s.n; counter++ {
		select {
		case _, ok := <-s.taskRes:
			if !ok {
				break ForLoop
			}
		case <-time.After(testTimeout):
			s.Require().Fail("test timeout exceeded")
			return
		}
	}

	close(s.stop)

	s.Require().Equal(counter, s.n)

	select {
	case _, ok := <-s.taskRes:
		if ok {
			s.Require().Fail("taskRes should be closed but isn't", s.suiteName)
		}
		s.Require().False(ok, "taskRes should be closed after %d tasks", s.suiteName, s.n)
	case <-time.After(testTimeout):
		s.Require().Fail("taskRes should be closed but isn't", s.suiteName)
	}
}
