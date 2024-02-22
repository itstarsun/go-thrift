package thrift

import (
	"context"
)

// A Client is the interface implemented by Thrift RPC clients.
// A Client is safe for concurrent use by multiple goroutines.
type Client interface {
	// Call invokes the named method. If the result is nil, a one-way message is sent.
	Call(ctx context.Context, method string, args, result any) error
}
