//nolint:revive
package hw06pipelineexecution

import "context"

//nolint:revive
type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

//nolint:revive
type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	if in == nil || len(stages) == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	prevStageChan := listenInput(ctx, in)

	for _, stage := range stages {
		prevStageChan = runStage(stage, prevStageChan)
	}

	return sendToOutput(cancel, done, prevStageChan)
}

func runStage(stage Stage, in In) Out {
	// Stage panic protection.
	return func() (out Out) {
		// Returning a closed channel on stage running panic.
		defer func() {
			if r := recover(); r != nil {
				tmpChan := make(Bi)
				close(tmpChan)
				out = tmpChan
				awaitChannel(in)
			}
		}()

		out = stage(in)
		return
	}()
}

func sendToOutput(cancel context.CancelFunc, done In, prevStageChan In) Out {
	res := make(Bi)
	go func() {
		// Lazy init of done channel.
		if done == nil {
			tmpDone := make(Bi)
			defer close(tmpDone)
			done = tmpDone
		}
		defer awaitChannel(prevStageChan)
		defer close(res)
		defer cancel()
		for {
			select {
			case <-done:
				return
			case v, ok := <-prevStageChan:
				if !ok {
					return
				}
				select {
				case <-done:
					return
				case res <- v:
					// Passing the value to the next stage.
				}
			}
		}
	}()
	return res
}

func listenInput(ctx context.Context, in In) Out {
	out := make(Bi)
	go func() {
		defer awaitChannel(in)
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					return
				case out <- v:
				}
			}
		}
	}()
	return out
}

func awaitChannel(in In) {
	go func() {
		for range in {
		}
	}()
}
