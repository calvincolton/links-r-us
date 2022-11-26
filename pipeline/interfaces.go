package pipeline

import "context"

// Payload is implemented by values that can be sent through a pipeline
type Payload interface {
	Clone() Payload
	MarkAsProcessed()
}

// Processor is implemented by types that can process Payloads as part of a pipeline stage
type Processor interface {
	Process(context.Context, Payload) (Payload, error)
}

// ProcessorFunc is an adapter to allow the use of plain functions as Processor instances
// If f is a function with the appropriate signature, ProcessorFunc(f) is a Processor that calls f
type ProcessorFunc func(context.Context, Payload) (Payload, error)

// Procss calls f(ctx, p)
func (f ProcessorFunc) Process(ctx context.Context, p Payload) (Payload, error) {
	return f(ctx, p)
}

// StageParams encapsulates the information required for executing a pipeline stage
// The pipeline passes a StageParams instance to the Run() method of each stage
type StageParams interface {
	StageIndex() int
	Input() <-chan Payload
	Output() chan<- Payload
	Error() chan<- error
}

// StageRunner is implemented by types that can be strung together to form a multi-stage pipeline
type StageRunner interface {
	// Run implements the processing logic for this stage by reading
	// incoming Payloads from an input channel, processing them and
	// outputting the results to an output channel.
	//
	// NOTE: Calls to Run are expected to block until:
	// - the stage input channel is closed OR
	// - the provided context expires OR
	// - an error occurs while process payloads
	Run(context.Context, StageParams)
}

// Source is implemented by types that generate Payload instances which can be used as inputs to a Pipeline instance
type Source interface {
	Next(context.Context) bool
	Payload() Payload
	Error() error
}

// Sink is implemented by types that can operate at the tail of a pipeline
type Sink interface {
	// Consume processes a Payload instance that has been emitted out of a Pipeline instance
	Consume(context.Context, Payload) error
}
