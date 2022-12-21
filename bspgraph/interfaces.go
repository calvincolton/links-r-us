package bspgraph

import "github.com/calvincolton/links-r-us/bspgraph/message"

// Aggregator is implemented by types that provide concurrent-safe aggregation primitives (e.g. counters, min/max, topN)
type Aggregator interface {
	Type() string
	Set(val interface{})
	Get() interface{}
	Aggregate(val interface{})
	Delta() interface{}
}

// Relayer is implemented by types that can relay messages to vertices that are managed by a remote graph instance
type Relayer interface {
	Relay(dst string, msg message.Message) error
}

// The RelayerFunc type is an adapter to allow the use of ordinary functions as Relayers.
// If f is a function with the appropriate signature, RelayerFunc(f) is a Relayer that calls f.
type RelayerFunc func(string, message.Message) error

// Relay calls f(dst, msg)
func (f RelayerFunc) Relay(dst string, msg message.Message) error {
	return f(dst, msg)
}

// ComputeFunc is a function that a graph instance invokes on each vertex when executing a superstep
type ComputeFunc func(g *Graph, v *Vertex, msgIt message.Iterator) error
