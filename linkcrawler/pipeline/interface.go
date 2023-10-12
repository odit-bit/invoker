package pipeline

import (
	"context"
)

//defined API interface that client or user will use

type Payload interface {
	Clone() Payload
	MarkAsProcessed()
}

type Processor interface {
	Process(context.Context, Payload) (Payload, error)
}

/*
write as middleware:

	func Example(next ProcessorFunc) ProcessorFunc {

		//code that defined here will immutable for entire process

		return ProcessorFunc(func(ctx context.Context, payload Payload) (Payload, error) {

			//do something in here will invoked everytime this function called

			return next(ctx, payload)
		})

		//do something here for after stage
	}

or wrap as ProcessorFunc

	func Example2(proc Processor, args any) Processor {

		return ProcessorFunc(func(ctx context.Context, payload Payload) (Payload, error) {
			fmt.Println(args)
			return nil, nil
		})

	}

the distinguished  will be end up with fn(ctx, payload) or fn.Process(ctx, payload)
*/
type ProcessorFunc func(ctx context.Context, payload Payload) (Payload, error)

func (pf ProcessorFunc) Process(ctx context.Context, payload Payload) (Payload, error) {
	return pf(ctx, payload)
}

// encapsulate the parameter to run the processor
type StageParams interface {
	//return the position of current index
	StageIndex() int

	//return a channel to reading the input
	Input() <-chan Payload

	// return a channel to writing the output
	Output() chan<- Payload

	// return a channel to writing error
	Error() chan<- error
}

// implement by types that form multi-stage pipeline
type StageRunner interface {
	Run(ctx context.Context, param StageParams)
}

// represent the source of payloads that can use as input of pipelines instance
// under the hood it will act as adapter to get payload from external system
type Source interface {
	Next() bool

	Payload() Payload

	Error() error
}

// represent the destination of payloads (endpoint ?)
// it act opposite as Source, conver payload to format that receive or consume by external system
type Sink interface {
	Consume(ctx context.Context, payload Payload) error
}
