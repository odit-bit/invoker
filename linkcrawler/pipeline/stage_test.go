package pipeline

import (
	"context"
	"fmt"
	"testing"
	"time"
)

var _ Payload = (*stubPayload)(nil)

type stubPayload struct {
	processed bool
	value     string
}

// Clone implements Payload.
func (p *stubPayload) Clone() Payload {
	return &stubPayload{value: p.value}
}

// MarkAsProcessed implements Payload.
func (p *stubPayload) MarkAsProcessed() {
	p.processed = true
}

var proc_Sleep1000Second = func() ProcessorFunc {

	return func(ctx context.Context, payload Payload) (Payload, error) {
		select {
		case <-time.After(1000 * time.Second):
			return payload, nil
		case <-ctx.Done():
			return nil, nil
		}

	}

}

var proc_return_error = func() ProcessorFunc {

	return func(ctx context.Context, payload Payload) (Payload, error) {
		return nil, fmt.Errorf("process error")

	}

}

var proc_passThrough = func() ProcessorFunc {
	return func(ctx context.Context, payload Payload) (Payload, error) {
		return payload, nil
	}
}

func Test_context_cancel(t *testing.T) {
	input := make(chan Payload)
	output := make(chan Payload)
	errCh := make(chan error)

	wp := workerParams{
		stage:  0,
		input:  input,
		output: output,
		errCh:  errCh,
	}

	payload := stubPayload{
		processed: false,
		value:     "",
	}

	expect := expect{
		err:   nil,
		value: nil,
	}

	runner := FIFO(proc_Sleep1000Second())
	ctx, cancel := context.WithCancel(context.Background())
	go runner.Run(ctx, &wp)

	input <- &payload
	cancel()
	actual := actual{}
	select {
	case actual.value = <-output:
	case actual.err = <-errCh:
	case <-ctx.Done():
		return
	}

	assert(t, &expect, &actual)
	close(output)
	close(input)
	close(errCh)
}

func Test_context_timeout(t *testing.T) {
	input := make(chan Payload)
	output := make(chan Payload)
	errCh := make(chan error)

	wp := workerParams{
		stage:  0,
		input:  input,
		output: output,
		errCh:  errCh,
	}

	payload := stubPayload{
		processed: false,
		value:     "",
	}

	expect := expect{
		err:   nil,
		value: nil,
	}

	runner := FIFO(proc_Sleep1000Second())
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	go runner.Run(ctx, &wp)

	input <- &payload

	actual := actual{}
	select {
	case actual.value = <-output:
	case actual.err = <-errCh:
	case <-ctx.Done():
		return
	}

	assert(t, &expect, &actual)
	close(output)
	close(input)
	close(errCh)
}

func Test_return_error(t *testing.T) {
	input := make(chan Payload)
	output := make(chan Payload)
	errCh := make(chan error)

	wp := workerParams{
		stage:  0,
		input:  input,
		output: output,
		errCh:  errCh,
	}

	payload := stubPayload{
		processed: false,
		value:     "",
	}

	expect := expect{
		err:   fmt.Errorf("process error"),
		value: nil,
	}

	runner := FIFO(proc_return_error())
	go func() {
		runner.Run(context.Background(), &wp)
	}()

	input <- &payload

	actual := actual{}

	select {
	case actual.value = <-output:
	case actual.err = <-errCh:
	case <-time.After(2 * time.Second):
		t.Fatalf("chan is blocking")
	}

	if actual.err != expect.err {
		t.Fatal("expecting error ")
	}
	// assert(t, &expect, &actual)
	close(output)
	close(input)
	close(errCh)
}

func Test_passthrough(t *testing.T) {
	input := make(chan Payload)
	output := make(chan Payload)
	errCh := make(chan error)

	wp := workerParams{
		stage:  0,
		input:  input,
		output: output,
		errCh:  errCh,
	}

	payload := stubPayload{
		processed: false,
		value:     "",
	}

	expect := expect{
		err:   nil,
		value: &payload,
	}

	runner := FIFO(proc_passThrough())
	go func() {
		runner.Run(context.Background(), &wp)
	}()

	input <- &payload
	actual := actual{}

	select {
	case actual.value = <-output:
	case actual.err = <-errCh:
	case <-time.After(2 * time.Second):
		t.Error("chan is blocking")
	}

	if actual.err != nil {
		assert(t, &expect, &actual)
	}
	close(output)
	close(input)
	close(errCh)
}

type expect struct {
	err   error
	value Payload
}

type actual struct {
	err   error
	value Payload
}

func assert(t *testing.T, expect *expect, actual *actual) {
	if actual.err.Error() != expect.err.Error() {
		t.Fatalf("\nactual: %v\nexpect: %v\n", actual.err, expect.err)
	}
	if actual.value != expect.value {
		p1 := actual.value.(*stubPayload)
		p2 := expect.value.(*stubPayload)

		if p1.value != p2.value {
			t.Fatalf("\nactual: %v\nexpect: %v\n", p1.value, p2.value)
		}
		// t.Errorf("\nactual: %v\nexpect: %v\n", actual.value, expect.value)
	}
}

// func createContext(useCancel bool) (context.Context, context.CancelFunc) {
// 	switch useCancel {
// 	case true:
// 		return context.WithCancel(context.Background())
// 	case false:
// 		return context.WithTimeout(context.Background(), 2*time.Second)
// 	default:
// 		panic("panic")
// 	}
// }

// func Test_table(t *testing.T) {
// 	var tt = []struct {
// 		name   string
// 		runner func(Processor) StageRunner
// 		proc   Processor
// 		// ctxTimeout time.Duration
// 		ctxCancel bool
// 		expect    expect
// 	}{
// 		// {
// 		// 	name:   "test context canceled",
// 		// 	runner: FIFO,
// 		// 	proc:   proc_Sleep1000Second(),
// 		// 	// ctxTimeout: 0,
// 		// 	ctxCancel: true,
// 		// 	expect: expect{
// 		// 		err:   context.Canceled,
// 		// 		value: nil,
// 		// 	},
// 		// },

// 		{
// 			name:   "test context timeout",
// 			runner: FIFO,
// 			proc:   proc_Sleep1000Second(),
// 			// ctxTimeout: 100 * time.Millisecond,
// 			ctxCancel: false,
// 			expect: expect{
// 				err:   nil,
// 				value: nil,
// 			},
// 		},

// 		{
// 			name:   "test process error",
// 			runner: FIFO,
// 			proc:   proc_return_error(),
// 			// ctxTimeout: 3,
// 			ctxCancel: false,
// 			expect: expect{
// 				err:   fmt.Errorf("process error"),
// 				value: nil,
// 			},
// 		},

// 		{
// 			name:   "test process success",
// 			runner: FIFO,
// 			proc:   proc_passThrough(),
// 			// ctxTimeout: 3,
// 			ctxCancel: false,
// 			expect: expect{
// 				err: nil,
// 				value: &stubPayload{
// 					value: "",
// 				},
// 			},
// 		},
// 	}

// 	for _, tc := range tt {
// 		t.Run(tc.name, func(t *testing.T) {
// 			input := make(chan Payload)
// 			output := make(chan Payload)
// 			errCh := make(chan error)

// 			payload := stubPayload{
// 				processed: false,
// 				value:     "",
// 			}

// 			runner := tc.runner(tc.proc)

// 			ctx, cancel := createContext(tc.ctxCancel)
// 			defer cancel()

// 			wp := workerParams{
// 				stage:  0,
// 				input:  input,
// 				output: output,
// 				errCh:  errCh,
// 			}
// 			go func() {
// 				runner.Run(ctx, &wp)

// 			}()

// 			input <- &payload
// 			if tc.ctxCancel {
// 				cancel()
// 			}

// 			actual := actual{}
// 			select {
// 			case actual.value = <-output:
// 			case actual.err = <-errCh:
// 				// case <-ctx.Done():
// 			}

// 			assert(t, &tc.expect, &actual)
// 			close(output)
// 			close(input)
// 			close(errCh)
// 		})
// 	}
// }

// func Test_FIFO(t *testing.T) {
// 	proc1 := func() ProcessorFunc {
// 		return func(ctx context.Context, payload Payload) (Payload, error) {
// 			p := payload.Clone()
// 			v, ok := p.(*stubPayload)
// 			if !ok {
// 				return nil, fmt.Errorf("invalid payload underlying value %T", payload)
// 			}
// 			v.value = "uye"
// 			v.processed = true
// 			return v, nil
// 		}
// 	}

// 	p := stubPayload{
// 		processed: false,
// 		value:     "lala",
// 	}

// 	stage := FIFO(proc1())

// 	src := make(chan Payload)
// 	dst := make(chan Payload)
// 	errC := make(chan error)
// 	go stage.Run(context.Background(), &workerParams{
// 		stage:  0,
// 		input:  src,
// 		output: dst,
// 		errCh:  errC,
// 	})

// 	src <- &p
// 	res := <-dst

// 	val := res.(*stubPayload)

// 	if !val.processed {
// 		t.Error("not preceed")
// 	}
// }

// func Test_WorkerPool(t *testing.T) {
// 	p := stubPayload{
// 		processed: false,
// 		value:     "lala",
// 	}

// 	proc1 := func() ProcessorFunc {
// 		return func(ctx context.Context, payload Payload) (Payload, error) {
// 			p := payload.Clone()
// 			v, ok := p.(*stubPayload)
// 			if !ok {
// 				return nil, fmt.Errorf("invalid payload underlying value %T", payload)
// 			}
// 			v.value = fmt.Sprintf("%v", "uye")
// 			v.processed = true
// 			return v, nil
// 		}
// 	}
// 	payloadIn := make(chan Payload)
// 	payloadOut := make(chan Payload)
// 	errC := make(chan error)
// 	wp := workerParams{
// 		stage:  0,
// 		input:  payloadIn,
// 		output: payloadOut,
// 		errCh:  errC,
// 	}
// 	f := WorkerPool(proc1(), 1000)
// 	go f.Run(context.Background(), &wp)

// 	payloadIn <- &p
// 	resP := <-payloadOut

// 	val := resP.(*stubPayload)

// 	if !val.processed {
// 		t.Error("false")
// 	}

// 	if val.value == p.value {
// 		t.Error("stage not processed", val.value, p.value)
// 	}

// 	close(payloadIn)
// }

// func Test_broadcast(t *testing.T) {

// 	procMarkProcessed := func() ProcessorFunc {
// 		return func(ctx context.Context, payload Payload) (Payload, error) {
// 			if payload == nil {
// 				t.Error("cannot be nil")
// 			}
// 			payload.MarkAsProcessed()
// 			return payload, nil
// 		}
// 	}

// 	procUnmarkProcessed := func() ProcessorFunc {
// 		return func(ctx context.Context, payload Payload) (Payload, error) {
// 			if payload == nil {
// 				t.Error("cannot be nil")
// 			}
// 			return payload, nil
// 		}
// 	}

// 	p := stubPayload{
// 		processed: false,
// 		value:     "",
// 	}

// 	payloadIn := make(chan Payload)
// 	payloadOut := make(chan Payload)
// 	errC := make(chan error)

// 	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
// 	defer cancel()

// 	b := Broadcast(procMarkProcessed(), procUnmarkProcessed())
// 	go b.Run(ctx, &workerParams{
// 		stage:  0,
// 		input:  payloadIn,
// 		output: payloadOut,
// 		errCh:  errC,
// 	})

// 	payloadIn <- &p
// 	res := <-payloadOut
// 	if res == nil {
// 		t.Error("result cannot be nil")
// 	}

// 	// close(payloadIn)
// }
