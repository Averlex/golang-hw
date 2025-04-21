package hw05parallelexecution

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const timeout = 100 * time.Millisecond

func formText(val int, text string) string {
	if val == 1 {
		return fmt.Sprintf("%d %s", val, text)
	}
	return fmt.Sprintf("%d %ss", val, text)
}

func TestInternal(t *testing.T) {
	t.Run("taskGenerator", generatorTests)
	t.Run("worker", workerTests)
	t.Run("mux", muxTests)
}

type generatorSuite struct {
	suite.Suite
	wg       *sync.WaitGroup
	n        int
	tasks    []Task
	stop     chan struct{}
	taskPool <-chan Task
}

func newGeneratorSuite(n int) *generatorSuite {
	return &generatorSuite{
		n: n,
	}
}

func generatorTests(t *testing.T) {
	t.Helper()

	testCases := []int{1, 2, 5, 10, 1000}

	for _, tC := range testCases {
		tc := tC
		t.Run(formText(tc, "task"), func(t *testing.T) {
			suite.Run(t, newGeneratorSuite(tc))
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
			s.Require().True(ok, "taskPool closed prematurely (got %d/%d tasks)", i+1, s.n)
		case <-time.After(timeout):
			s.Require().Fail("test timeout exceeded")
			return
		}
	}

	select {
	case _, ok := <-s.taskPool:
		if ok {
			s.Require().Fail("taskPool should be closed but isn't")
		}
		s.Require().False(ok, "taskPool should be closed after %d tasks", s.n)
	case <-time.After(timeout):
		s.Require().Fail("test timeout exceeded")
		return
	}
}

func (s *generatorSuite) TestStopBeforeExctracting() {
	close(s.stop)

	select {
	case _, ok := <-s.taskPool:
		s.Require().False(ok, "taskPool should be closed before sending tasks")
	case <-time.After(timeout):
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
			s.Require().True(ok, "taskPool closed prematurely (got %d/%d tasks)", counter+1, stopValue)
		case <-time.After(timeout):
			s.Require().Fail("test timeout exceeded")
			return
		}
	}

	s.Require().Equal(counter, stopValue)

	select {
	case _, ok := <-s.taskPool:
		if ok {
			s.Require().Fail("taskPool should be closed but isn't")
		}
		s.Require().False(ok, "taskPool should be closed after %d tasks", stopValue)
	case <-time.After(timeout):
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
			s.Require().True(ok, "taskPool closed prematurely (got %d/%d tasks)", counter+1, s.n)
		case <-time.After(timeout):
			s.Require().Fail("test timeout exceeded")
			return
		}
	}

	close(s.stop)

	s.Require().Equal(counter, stopValue)

	select {
	case _, ok := <-s.taskPool:
		if ok {
			s.Require().Fail("taskPool should be closed but isn't")
		}
		s.Require().False(ok, "taskPool should be closed after %d tasks", stopValue)
	case <-time.After(timeout):
		s.Require().Fail("test timeout exceeded")
		return
	}
}

type workerSuite struct {
	suite.Suite
	wg       *sync.WaitGroup
	n        int
	isError  bool
	isPanic  bool
	tasks    []Task
	stop     chan struct{}
	taskPool <-chan Task
	taskRes  chan error
}

func newWorkerSuite(n int, isError, isPanic bool) *workerSuite {
	return &workerSuite{
		n:       n,
		isError: isError,
		isPanic: isPanic,
	}
}

func workerTests(t *testing.T) {
	t.Helper()

	testCases := []int{1, 2, 5, 10, 1000}

	for _, tC := range testCases {
		tc := tC

		t.Run(formText(tc, "task"), func(t *testing.T) {
			t.Run("normal", func(t *testing.T) {
				suite.Run(t, newWorkerSuite(tc, false, false))
			})
			t.Run("with error", func(t *testing.T) {
				suite.Run(t, newWorkerSuite(tc, true, false))
			})
			t.Run("with panic", func(t *testing.T) {
				suite.Run(t, newWorkerSuite(tc, false, true))
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
			s.Require().True(ok, "taskRes closed prematurely (got %d/%d tasks)", i+1, s.n)
			s.Require().Equal(erroneousCond, v != nil,
				`taskRes received incorrect value on task %d/%d:
				- got %v
				- isPanic=%v
				- isError=%v`,
				i+1, s.n, v, s.isPanic, s.isError)
		case <-time.After(timeout):
			s.Require().Fail("test timeout exceeded")
			return
		}
	}

	select {
	case _, ok := <-s.taskRes:
		if ok {
			s.Require().Fail("taskRes should be closed but isn't")
		}
		s.Require().False(ok, "taskRes should be closed after receiving %d tasks", s.n)
	case <-time.After(timeout):
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
		case <-time.After(timeout):
			s.Require().Fail("test timeout exceeded")
			return
		}
	}

	s.Require().Equal(counter, 0, "tasks done before stop - %v, expected - %v", counter, 0)
	select {
	case _, ok := <-s.taskRes:
		s.Require().False(ok, "taskRes should be closed before sending tasks")
	case <-time.After(timeout):
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
		case <-time.After(timeout):
			s.Require().Fail("test timeout exceeded")
			return
		}
	}

	// Might have done stopValue and stopValue+1 tasks as well.
	s.Require().LessOrEqual(counter, stopValue+1)

	select {
	case _, ok := <-s.taskRes:
		if ok {
			s.Require().Fail("taskRes should be closed but isn't")
		}
		s.Require().False(ok, "taskRes should be closed after %d tasks", stopValue)
	case <-time.After(timeout):
		s.Require().Fail("taskRes should be closed but isn't")
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
		case <-time.After(timeout):
			s.Require().Fail("test timeout exceeded")
			return
		}
	}

	close(s.stop)

	s.Require().Equal(counter, s.n)

	select {
	case _, ok := <-s.taskRes:
		if ok {
			s.Require().Fail("taskRes should be closed but isn't")
		}
		s.Require().False(ok, "taskRes should be closed after %d tasks", s.n)
	case <-time.After(timeout):
		s.Require().Fail("taskRes should be closed but isn't")
	}
}

type muxSuite struct {
	suite.Suite
	wg       *sync.WaitGroup
	chanNum  int
	taskNum  int
	channels []chan error
	res      <-chan error
}

func (s *muxSuite) distributeTasksEvenly() []int {
	distribution := make([]int, s.chanNum)
	if s.chanNum == 0 {
		return distribution
	}
	base := s.taskNum / s.chanNum
	remainder := s.taskNum % s.chanNum
	for i := 0; i < s.chanNum; i++ {
		distribution[i] = base
		if i < remainder {
			distribution[i]++
		}
	}
	return distribution
}

func newMuxSuite(chanNum, taskNum int) *muxSuite {
	return &muxSuite{
		chanNum: chanNum,
		taskNum: taskNum,
	}
}

func muxTests(t *testing.T) {
	t.Helper()

	chanNum := []int{0, 1, 2, 10, 1000}
	taskNum := []int{0, 1, 2, 100, 10000}

	for _, cN := range chanNum {
		cn := cN
		t.Run(formText(cn, "channel"), func(t *testing.T) {
			for _, tN := range taskNum {
				tn := tN
				t.Run(formText(tn, "task"), func(t *testing.T) { suite.Run(t, newMuxSuite(cn, tn)) })
			}
		})
	}
}

func (s *muxSuite) SetupTest() {
	s.wg = &sync.WaitGroup{}
	s.channels = make([]chan error, s.chanNum)
	arg := make([]<-chan error, s.chanNum)
	for i := range s.channels {
		s.channels[i] = make(chan error)
		arg[i] = s.channels[i]
	}

	s.res = muxChannels(s.wg, arg...)
}

func (s *muxSuite) TearDownTest() {
	s.wg.Wait()
}

func (s *muxSuite) TestMux() {
	testWG := &sync.WaitGroup{}
	defer testWG.Wait()

	simpleWorker := func(wg *sync.WaitGroup, ch chan<- error, taskNum int) {
		defer wg.Done()
		defer close(ch)
		for range taskNum {
			ch <- nil
		}
	}

	// If chanNum > taskNum, then distribution will be like [n, n, n, ..., 0, 0, 0, 0, 0].
	distributedTasks := s.distributeTasksEvenly()

	// Loading channels with evenly distributed tasks.
	for i := range s.chanNum {
		testWG.Add(1)
		go simpleWorker(testWG, s.channels[i], distributedTasks[i])
	}

	counter := 0
	for ; counter < s.taskNum; counter++ {
		select {
		case _, ok := <-s.res:
			if !ok {
				s.Require().Equal(counter+1, s.taskNum, "channel closed prematurely (got %d/%d tasks)", counter+1, s.taskNum)
			}
		case <-time.After(timeout):
			s.Require().Fail("test timeout exceeded")
		}
	}

	s.Require().Equal(counter, s.taskNum, "channel received %d tasks, expected %d", counter, s.taskNum)

	select {
	case _, ok := <-s.res:
		if ok {
			s.Require().Fail("channel should be closed but isn't")
		}
		s.Require().False(ok, "channel should be closed after %d tasks", s.taskNum)
	case <-time.After(timeout):
		s.Require().Fail("channel should be closed but isn't")
	}
	// Проверка значений (nil || error).
}
