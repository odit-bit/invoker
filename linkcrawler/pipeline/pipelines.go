package pipeline

// implementation code of pipeline

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/multierr"
)

type Pipeline struct {
	stages []StageRunner
}

func New(stage ...StageRunner) *Pipeline {
	return &Pipeline{
		stages: stage,
	}
}

func (p *Pipeline) Process(ctx context.Context, src Source, dst Sink) error {
	var wg sync.WaitGroup
	pCtx, cancel := context.WithCancel(ctx)

	//allocate channel for wiring source , stage , sink
	stageCh := make([]chan Payload, len(p.stages)+1)
	errCh := make(chan error, len(p.stages)+2)

	for i := 0; i < len(stageCh); i++ {
		stageCh[i] = make(chan Payload)
	}

	for i := 0; i < len(p.stages); i++ {
		wg.Add(1)

		go func(stageIndex int) {
			p.stages[stageIndex].Run(pCtx, &workerParams{
				stage:  stageIndex,
				input:  stageCh[stageIndex],
				output: stageCh[stageIndex+1],
				errCh:  errCh,
			})

			// when stagerunner return
			//it will close the output chan
			close(stageCh[stageIndex+1])
			wg.Done()
		}(i)

	}

	//source worker
	wg.Add(1)
	go func() {
		sourceWorker(pCtx, stageCh[0], errCh, src)
		//when sourceWorker return it will
		//close the source chan
		close(stageCh[0])
		wg.Done()
	}()

	// sink worker
	wg.Add(1)
	go func() {
		sinkWorker(pCtx, stageCh[len(stageCh)-1], errCh, dst)
		wg.Done()
	}()

	go func() {
		wg.Wait()
		close(errCh)
		cancel()
	}()

	var err error
	for pErr := range errCh {
		err = multierr.Append(err, pErr)
		cancel()
	}

	return err
}

// sourceWorker implements a worker that reads Payload instances from a Source
// and pushes them to an output channel that is used as input for the first
// stage of the pipeline.
func sourceWorker(ctx context.Context, stageCh chan<- Payload, errCh chan error, src Source) {
	for src.Next() {
		payload := src.Payload()
		select {
		// case <-ctx.Done():
		case stageCh <- payload:
		case <-ctx.Done():
			return
		}
	}

	if err := src.Error(); err != nil {
		errCh <- fmt.Errorf("pipeline source err: %v ", err)
		emitError(err, errCh)
	}

}

// sinkWorker implements a worker that reads Payload instances from an input
// channel (the output of the last pipeline stage) and passes them to the
// provided sink.
func sinkWorker(ctx context.Context, stageCh <-chan Payload, errCh chan error, dst Sink) {

	for {
		select {
		case <-ctx.Done():
			return
		case payload, ok := <-stageCh:
			if !ok {
				return
			}

			if err := dst.Consume(ctx, payload); err != nil {
				errCh <- fmt.Errorf("pipeline sink err: %v ", err)
				emitError(err, errCh)
				return
			}
			payload.MarkAsProcessed()
		}
	}

}
