package aqm

import "context"

// Runner represents a lifecycle-managed component such as an HTTP or gRPC server.
type Runner interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}
