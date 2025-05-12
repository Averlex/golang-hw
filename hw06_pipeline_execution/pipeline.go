//nolint:revive
package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	if in == nil || len(stages) == 0 {
		return nil
	}

	prevStageChan := in

	for _, stage := range stages {
		prevStageChan = runStage(stage, prevStageChan, done)
	}

	return prevStageChan
}

func runStage(stage Stage, in In, done In) Out {
	transit := make(Bi)

	go func() {
		if done == nil {
			ch := make(Bi)
			defer close(ch)
			done = ch
		}

		defer close(transit)

		// Draining the channel in case of stop signal.
		defer func() {
			for range in {
			}
		}()

		for {
			select {
			case <-done:
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case <-done:
					return
				case transit <- v:
				}
			}
		}
	}()

	out := stage(transit)

	return out
}
