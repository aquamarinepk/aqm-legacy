package aqm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCServiceRegistrar can register itself with a gRPC server.
type GRPCServiceRegistrar interface {
	RegisterGRPCService(server *grpc.Server)
}

// GRPCServiceFactory constructs a GRPCServiceRegistrar from the shared dependency container.
type GRPCServiceFactory func(*Deps) (GRPCServiceRegistrar, error)

// WithGRPCServer wires a gRPC server runner. It instantiates the provided service
// factories, registers their services with the gRPC server, and mounts the resulting
// server as a lifecycle-managed runner.
//
// Usage:
//   aqm.WithGRPCServer("grpc.port", serviceFactory1, serviceFactory2)
//
// The addrKey is used to look up the server address from config (e.g., "grpc.port" -> ":50051").
// If the config key is not found, it defaults to ":50051".
func WithGRPCServer(addrKey string, factories ...GRPCServiceFactory) Option {
	return func(ms *Micro) error {
		if addrKey == "" {
			return errors.New("grpc addr property key required")
		}

		grpcServer := grpc.NewServer()

		// Enable reflection for easier debugging with grpcurl/grpcui
		reflection.Register(grpcServer)

		// Instantiate and register all gRPC services
		for _, factory := range factories {
			if factory == nil {
				return errors.New("nil grpc service factory")
			}
			service, err := factory(ms.deps)
			if err != nil {
				return fmt.Errorf("building grpc service: %w", err)
			}
			if service == nil {
				return errors.New("grpc service factory returned nil service")
			}
			service.RegisterGRPCService(grpcServer)

			// Support lifecycle hooks
			if startable, ok := service.(Startable); ok {
				ms.addStart(startable.Start)
			}
			if stoppable, ok := service.(Stoppable); ok {
				ms.addStop(stoppable.Stop)
			}
		}

		addr := ms.deps.Config.GetPort(addrKey, ":50051")

		ms.runners = append(ms.runners, newGRPCServerRunner(addr, grpcServer))
		return nil
	}
}

// WithGRPCServerModules is a convenience helper for the common case where
// services do not need to access the shared dependency container during
// construction. It wraps the provided services into factories and delegates to
// WithGRPCServer.
func WithGRPCServerModules(addrKey string, services ...GRPCServiceRegistrar) Option {
	factories := make([]GRPCServiceFactory, len(services))
	for i, svc := range services {
		service := svc
		factories[i] = func(*Deps) (GRPCServiceRegistrar, error) {
			if service == nil {
				return nil, errors.New("nil grpc service provided")
			}
			return service, nil
		}
	}
	return WithGRPCServer(addrKey, factories...)
}

type grpcServerRunner struct {
	addr   string
	server *grpc.Server
	errCh  chan error
}

func newGRPCServerRunner(addr string, server *grpc.Server) Runner {
	return &grpcServerRunner{
		addr:   addr,
		server: server,
		errCh:  make(chan error, 1),
	}
}

func (r *grpcServerRunner) Start(_ context.Context) error {
	lis, err := net.Listen("tcp", r.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", r.addr, err)
	}

	go func() {
		if err := r.server.Serve(lis); err != nil {
			r.errCh <- err
		}
		close(r.errCh)
	}()
	return nil
}

func (r *grpcServerRunner) Stop(ctx context.Context) error {
	stopped := make(chan struct{})
	go func() {
		r.server.GracefulStop()
		close(stopped)
	}()

	// Wait for graceful stop with timeout
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	select {
	case <-stopped:
		// Graceful stop completed
	case <-timer.C:
		// Timeout - force stop
		r.server.Stop()
	case <-ctx.Done():
		// Context cancelled - force stop
		r.server.Stop()
	}

	// Check for any errors from Serve
	var err error
	select {
	case srvErr, ok := <-r.errCh:
		if ok && srvErr != nil {
			err = srvErr
		}
	default:
	}
	return err
}
