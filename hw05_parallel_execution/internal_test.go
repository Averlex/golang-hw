package hw05parallelexecution

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestInternal(t *testing.T) {
	t.Run("taskGenerator", generatorTests)
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

	testCases := []struct {
		name string
		n    int
	}{
		{name: "1 workers", n: 1},
		{name: "2 workers", n: 2},
		{name: "5 workers", n: 5},
		{name: "10 workers", n: 10},
		{name: "1000 workers", n: 1000},
	}

	for _, tC := range testCases {
		tc := tC
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			suite.Run(t, newGeneratorSuite(tc.name, tc.n))
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
	close(s.stop)
}

func (s *generatorSuite) TestBuildingPool() {
	for i := 0; i < s.n; i++ {
		_, ok := <-s.taskPool
		s.Require().True(ok, "[%s] taskPool closed prematurely (got %d/%d tasks)", s.suiteName, i+1, s.n)
	}

	select {
	case _, ok := <-s.taskPool:
		if ok {
			s.Fail("[%s] taskPool should be closed but isn't", s.suiteName)
		}
		s.Require().False(ok, "[%s] taskPool should be closed after %d tasks", s.suiteName, s.n)
	default:
	}
}

func (s *generatorSuite) TestStopBeforeExctracting() {
	s.stop <- struct{}{}
	counter := 0

	_, ok := <-s.taskPool
	s.Require().False(ok, "[%s] taskPool should be closed after %d tasks", s.suiteName, counter, s.n)

	s.Require().Equal(counter, 0)
}

func (s *generatorSuite) TestStopDuringExctracting() {
	counter := 0
	stopValue := s.n / 2

	for ; counter < s.n; counter++ {
		if counter == stopValue {
			s.stop <- struct{}{}
			break
		}
		_, ok := <-s.taskPool
		s.Require().True(ok, "[%s] taskPool closed prematurely (got %d/%d tasks)", s.suiteName, counter+1, stopValue)
	}

	s.Require().Equal(counter, stopValue)

	select {
	case _, ok := <-s.taskPool:
		if ok {
			s.Fail("[%s] taskPool should be closed but isn't", s.suiteName)
		}
		s.Require().False(ok, "[%s] taskPool should be closed after %d tasks", s.suiteName, stopValue)
	default:
	}
}

func (s *generatorSuite) TestStopAfterExctracting() {
	counter := 0
	stopValue := s.n

	for ; counter < s.n; counter++ {
		_, ok := <-s.taskPool
		s.Require().True(ok, "[%s] taskPool closed prematurely (got %d/%d tasks)", s.suiteName, counter+1, s.n)
	}

	select {
	case s.stop <- struct{}{}:
	default:
	}

	s.Require().Equal(counter, stopValue)

	select {
	case _, ok := <-s.taskPool:
		if ok {
			s.Fail("[%s] taskPool should be closed but isn't", s.suiteName)
		}
		s.Require().False(ok, "[%s] taskPool should be closed after %d tasks", s.suiteName, stopValue)
	default:
	}
}
