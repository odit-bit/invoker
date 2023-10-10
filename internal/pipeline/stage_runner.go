package pipeline

import (
	"context"
	"sync"
)

// the type that can process incoming payloads
// as synchronous like dispatch

var _ StageRunner = (*fifo)(nil)

type fifo struct {
	proc Processor
}

// instantiate the StageRunner that proceess payload as First-In-First-Out fashion
func FIFO(proc Processor) StageRunner {
	return &fifo{
		proc: proc,
	}
}

// Run implements StageRunner.
func (f *fifo) Run(ctx context.Context, param StageParams) {
	// should return error if <-ctx.done or canceled, and payload return from processor is nil

	for {
		select {
		case <-ctx.Done():
			return
		case PayloadIn, ok := <-param.Input():
			if !ok {
				//chan maybe closed
				return
			}
			payloadOut, err := f.proc.Process(ctx, PayloadIn)
			if err != nil {
				emitError(err, param.Error())
				return
			}
			// If the processor did not output a payload for the
			// next stage there is nothing we need to do.
			if payloadOut == nil {
				PayloadIn.MarkAsProcessed()
				continue
			}
			select {
			case <-ctx.Done():
				return
			case param.Output() <- payloadOut:
				//payload go to next stage
			}
		}

	}
}

var _ StageRunner = (*workerPool)(nil)

// like worker pool is a list of fifo runner but it worker is predefined
// because it implemented the loop like fifo, we instantiate the fifo to reduce code duplication
type workerPool struct {
	fifos []StageRunner
}

func WorkerPool(proc Processor, numWorker int) StageRunner {
	fifos := make([]StageRunner, numWorker)

	for i := range fifos {
		fifos[i] = FIFO(proc)
	}

	wp := workerPool{
		fifos: fifos,
	}

	return &wp
}

// Run implements StageRunner.
func (wp *workerPool) Run(ctx context.Context, param StageParams) {
	var wg sync.WaitGroup

	for i := 0; i < len(wp.fifos); i++ {
		wg.Add(1)
		go func(fifoIndex int) {
			wp.fifos[fifoIndex].Run(ctx, param)
			wg.Done()
		}(i)
	}

	wg.Wait()
}

func emitError(err error, errC chan<- error) {
	select {
	case errC <- err:
	// case <-time.After(2 * time.Second):
	// 	log.Fatal("stage error chan is blocking")
	// }
	default:
		// 	//error chan full
	}
}

var _ StageRunner = (*broadcast)(nil)

type broadcast struct {
	runners []StageRunner
}

func Broadcast(procs ...Processor) *broadcast {
	fifos := make([]StageRunner, len(procs))
	for i, v := range procs {
		fifos[i] = FIFO(v)
	}

	return &broadcast{
		runners: fifos,
	}
}

// Run implements StageRunner.
func (br *broadcast) Run(ctx context.Context, param StageParams) {
	var (
		wg sync.WaitGroup
		//list of chan (stil empty)
		inC = make([]chan Payload, len(br.runners))
	)

	// setup go routine
	for i := 0; i < len(br.runners); i++ {
		wg.Add(1)
		inC[i] = make(chan Payload)

		go func(index int) {
			br.runners[index].Run(ctx, &workerParams{
				stage:  param.StageIndex(),
				input:  inC[index],
				output: param.Output(),
				errCh:  param.Error(),
			})
			wg.Done()
		}(i)

	}

	//send clone of payload to precede(inC) list of channel
done:
	for {
		select {
		case <-ctx.Done():
			break done
		case input, ok := <-param.Input():
			if !ok {
				break done
			}
			if ok := br.proceedNextStage(ctx, input, inC); !ok {
				break done
			}
		}
	}

	for _, c := range inC {
		close(c)
	}

	wg.Wait()
}

// helper

func (br *broadcast) proceedNextStage(ctx context.Context, input Payload, inC []chan Payload) bool {
	// for i := 0; i < len(inC); i++ {
	// 	var cloneP = input
	// 	if i != 0 {
	// 		cloneP = input.Clone()
	// 	}
	// 	select {
	// 	case <-ctx.Done():
	// 		return
	// 	case inC[i] <- cloneP:
	// 	}
	// }

	for i := len(br.runners) - 1; i >= 0; i-- {
		var cloneP = input
		if i != 0 {
			cloneP = input.Clone()
		}
		select {
		case <-ctx.Done():
			return false
		case inC[i] <- cloneP:
		}
	}
	return true
}
