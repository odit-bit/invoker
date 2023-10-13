package pipeline_test

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/odit-bit/invoker/linkcrawler/pipeline"
)

var _ pipeline.Payload = (*stubPayload)(nil)

type stubPayload struct {
	processed bool
	value     string
}

// Clone implements Payload.
func (p *stubPayload) Clone() pipeline.Payload {
	return &stubPayload{value: p.value}
}

// MarkAsProcessed implements Payload.
func (p *stubPayload) MarkAsProcessed() {
	p.processed = true
}

// /================
var _ pipeline.Source = (*mockSrc)(nil)

type mockSrc struct {
	payload []pipeline.Payload
	idx     int
	err     error
}

func stubSource(num int) *mockSrc {
	payloads := make([]pipeline.Payload, num)
	for i := 0; i < num; i++ {
		payloads[i] = &stubPayload{
			processed: false,
			value:     fmt.Sprint(i),
		}
	}

	src := mockSrc{
		payload: payloads,
		idx:     0,
	}
	return &src
}

// Error implements Source.
func (ms *mockSrc) Error() error {
	return ms.err
}

// Next implements Source.
func (ms *mockSrc) Next() bool {
	return ms.idx < len(ms.payload)
}

// Payload implements Source.
func (ms *mockSrc) Payload() pipeline.Payload {
	p := ms.payload[ms.idx]
	ms.idx++
	return p
}

//==============

var _ pipeline.Sink = (*mockSink)(nil)

type mockSink struct {
	fromPipe      []any
	expectCounter int
	realCounter   int
}

func stubSink(expected int) *mockSink {
	ms := mockSink{
		expectCounter: expected,
		realCounter:   0,
	}
	return &ms
}

// Consume implements Sink.
func (ms *mockSink) Consume(ctx context.Context, payload pipeline.Payload) error {
	v := payload.(*stubPayload)
	ms.fromPipe = append(ms.fromPipe, v.value)
	ms.realCounter++
	return nil
}

// processor
var proc1 = func() pipeline.ProcessorFunc {
	count := 0
	return func(ctx context.Context, payload pipeline.Payload) (pipeline.Payload, error) {
		p := payload.Clone()
		v, ok := p.(*stubPayload)
		if !ok {
			return nil, fmt.Errorf("invalid payload underlying value %T", payload)
		}
		v.value = fmt.Sprint(count)
		v.processed = true
		count++
		return v, nil
	}
}

// processor
var proc2 = func() pipeline.ProcessorFunc {
	return func(ctx context.Context, payload pipeline.Payload) (pipeline.Payload, error) {
		//just passing through
		return payload, nil
	}
}

var proc3 = func(n int64) pipeline.ProcessorFunc {
	return func(ctx context.Context, payload pipeline.Payload) (pipeline.Payload, error) {
		dur := time.Duration(n) * time.Second

		select {
		case <-time.After(dur):
			// mock the long process duration
			return payload, nil
		case <-ctx.Done():
			// context being cancel
			return nil, fmt.Errorf("context canceled")
		}

	}
}

func test_pipeline(source pipeline.Source, sink pipeline.Sink, runner ...pipeline.StageRunner) func(t *testing.T) {

	pipe := pipeline.New(runner...)
	errCh := make(chan error)

	return func(t *testing.T) {
		go func(errC chan error) {
			errC <- pipe.Process(context.Background(), source, sink)
			close(errC)
		}(errCh)

		err := <-errCh
		if err != nil {
			t.Error(err)
		}

	}
}

var _ pipeline.StageRunner = (*ErrorStage)(nil)

type ErrorStage struct{}

// Run implements pipeline.StageRunner.
func (*ErrorStage) Run(ctx context.Context, param pipeline.StageParams) {
	param.Error() <- fmt.Errorf("should error")
}

func test_pipeline_error(src pipeline.Source, dst pipeline.Sink) func(t *testing.T) {
	p := pipeline.New(&ErrorStage{})
	return func(t *testing.T) {
		err := p.Process(context.TODO(), src, dst)
		if err == nil {
			t.Error("error cannot be nil")
		}
	}
}

func test_ctx_done(src pipeline.Source, dst pipeline.Sink) func(t *testing.T) {
	fifos := make([]pipeline.StageRunner, 1000)
	for i := range fifos {
		fifos[i] = pipeline.FIFO(proc3(1000))
	}

	p := pipeline.New(fifos...)
	return func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			err := p.Process(ctx, src, dst)
			if err != nil {
				t.Error(err)
			}
			wg.Done()
		}()

		wg.Wait()

	}
}

// ==============================

func Test_(t *testing.T) {

	t.Run("pipeline_fifo", func(t *testing.T) {
		source := stubSource(2)
		sink := stubSink(2)
		expected := []any{"0", "1"}

		//pipeline with multi fifo runner
		procList := []pipeline.StageRunner{
			pipeline.FIFO(proc1()),
			pipeline.FIFO(proc2()),
		}

		test_pipeline(source, sink, procList...)(t)
		if sink.expectCounter != sink.realCounter {
			t.Error("not reach sink")
		}

		if !reflect.DeepEqual(expected, sink.fromPipe) {
			t.Errorf("\n%v\n%v\n", expected, sink.fromPipe)
		}

	})

	t.Run("pipeline_no_stage", func(t *testing.T) {
		source := stubSource(2)
		sink := stubSink(2)
		//pipeline with no runner
		test_pipeline(source, sink)(t)
	})

	t.Run("pipeline_error", func(t *testing.T) {
		src := stubSource(0)
		src.err = fmt.Errorf("should error")

		sink := stubSink(0)
		test_pipeline_error(src, sink)(t)
	})

	// t.Run("pipeline_ctx_done", func(t *testing.T) {
	// 	src := stubSource(1000)
	// 	sink := stubSink(0)
	// 	test_ctx_done(src, sink)(t)
	// })

}
