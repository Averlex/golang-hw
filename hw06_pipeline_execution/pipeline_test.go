package hw06pipelineexecution

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
)

var isFullTesting = true

func TestPipeline(t *testing.T) {
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("simple case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})
}

func TestAllStageStop(t *testing.T) {
	if !isFullTesting {
		return
	}
	wg := sync.WaitGroup{}
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()

		result := make([]string, 0, 10)
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		wg.Wait()

		require.Len(t, result, 0)
	})
}

func TestIncorrectArguments(t *testing.T) {
	wg := &sync.WaitGroup{}

	in := make(Bi)
	defer close(in)

	testCases := []struct {
		name   string
		in     Bi
		done   Bi
		stages []Stage
	}{
		{"nil input channel", nil, nil, []Stage{g(wg, "Dummy", func(v interface{}) interface{} { return v })}},
		{"empty stages slice", in, nil, nil},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			got := ExecutePipeline(tC.in, tC.done, tC.stages...)
			require.Nil(t, got, "unexpected not-nil channel received from the pipeline")
		})
	}

	wg.Wait()
}

func TestPanicingStages(t *testing.T) {
	wg := &sync.WaitGroup{}

	testCases := []struct {
		name   string
		done   Bi
		stages []Stage
	}{
		{"1 stage", nil, []Stage{panicingG("Dummy", func(v interface{}) interface{} { return v })}},
		{"2 stages/1st panics", nil, []Stage{
			panicingG("Dummy", func(v interface{}) interface{} { return v }),
			g(wg, "Dummy", func(v interface{}) interface{} { return v }),
		}},
		{"2 stages/2nd panics", nil, []Stage{
			g(wg, "Dummy", func(v interface{}) interface{} { return v }),
			panicingG("Dummy", func(v interface{}) interface{} { return v }),
		}},
		{"many stages/many panics", nil, []Stage{
			g(wg, "Dummy", func(v interface{}) interface{} { return v }),
			panicingG("Dummy", func(v interface{}) interface{} { return v }),
			g(wg, "Dummy", func(v interface{}) interface{} { return v }),
			panicingG("Dummy", func(v interface{}) interface{} { return v }),
			g(wg, "Dummy", func(v interface{}) interface{} { return v }),
			panicingG("Dummy", func(v interface{}) interface{} { return v }),
			g(wg, "Dummy", func(v interface{}) interface{} { return v }),
		}},
	}

	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			in := make(Bi)
			data := []int{1, 2, 3, 4, 5}
			go func() {
				for _, v := range data {
					in <- v
				}
				close(in)
			}()

			result := make([]string, 0, 10)

			for s := range ExecutePipeline(in, tC.done, tC.stages...) {
				result = append(result, s.(string))
			}

			require.Len(t, result, 0)
		})
	}

	wg.Wait()
}

// Stage generator.
func g(wg *sync.WaitGroup, _ string, f func(v interface{}) interface{}) Stage {
	return func(in In) Out {
		out := make(Bi)
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(out)
			for v := range in {
				time.Sleep(sleepPerStage)
				out <- f(v)
			}
		}()
		return out
	}
}

// Stage generator which panics in process.
func panicingG(_ string, _ func(v interface{}) interface{}) Stage {
	return func(_ In) Out {
		panic("42")
	}
}
