package pipeline

var _ StageParams = (*workerParams)(nil)

type workerParams struct {
	stage int

	input  <-chan Payload
	output chan<- Payload
	errCh  chan<- error
}

// Error implements StageParams.
func (wp *workerParams) Error() chan<- error {
	return wp.errCh
}

// Input implements StageParams.
func (wp *workerParams) Input() <-chan Payload {
	return wp.input
}

// Output implements StageParams.
func (wp *workerParams) Output() chan<- Payload {
	return wp.output
}

// StageIndex implements StageParams.
func (wp *workerParams) StageIndex() int {
	return wp.stage
}
