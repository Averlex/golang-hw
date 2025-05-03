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
	transits := make([]Bi, len(stages))

	prevStageChan := getSource(ctx, in)

	for i, stage := range stages {
		//nolint:copyloopvar
		i := i
		transits[i] = make(Bi)
		go runStage(ctx, stage, prevStageChan, transits[i])
		prevStageChan = transits[i]
	}

	// Awaiting for cancel signal or end of processing.
	res := make(Bi)
	if done != nil {
		go listenRes(cancel, prevStageChan, res, done)
	} else {
		go listenRes(cancel, prevStageChan, res, make(In)) // Stubbing done channel if no stop is required.
	}
	return res
}

func runStage(ctx context.Context, stage Stage, in In, out chan<- interface{}) {
	defer close(out)

	var localOut Out

	// Stage panic protection.
	localOut = func(in In) (out Out) {
		// Returning a closed channel on stage running panic.
		defer func() {
			if r := recover(); r != nil {
				tmpChan := make(Bi)
				close(tmpChan)
				out = tmpChan
			}
		}()

		localOut = stage(in)
		return localOut
	}(in)

	// Delivering results from one stage to another.
	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-localOut:
			// Extracting the value from the current stage.
			if !ok {
				return
			}
			select {
			case <-ctx.Done():
				return
			case out <- v:
				// Passing the value to the next stage.
			}
		}
	}
}

func listenRes(cancel context.CancelFunc, prevStageChan In, res chan<- interface{}, done In) {
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
}

func getSource(ctx context.Context, in In) Out {
	out := make(Bi)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				for range in {
				} // Reading all values from the pipeline input channel.
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					for range in {
					}
					return
				case out <- v:
				}
			}
		}
	}()
	return out
}

// listenRes - переименовать, подумать, как сделать последние строчки ExecutePipeline чуть более graceful.
//
// Разделить runStage на 2 части???
