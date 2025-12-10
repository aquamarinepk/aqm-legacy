package aqm

import (
	"context"
	"errors"
	"testing"
)

func TestRunnerInterface(t *testing.T) {
	var _ Runner = &testRunnerImpl{}
}

type testRunnerImpl struct {
	startCalled bool
	stopCalled  bool
	startErr    error
	stopErr     error
}

func (r *testRunnerImpl) Start(ctx context.Context) error {
	r.startCalled = true
	return r.startErr
}

func (r *testRunnerImpl) Stop(ctx context.Context) error {
	r.stopCalled = true
	return r.stopErr
}

func TestRunnerImplStart(t *testing.T) {
	runner := &testRunnerImpl{}

	err := runner.Start(context.Background())
	if err != nil {
		t.Errorf("Start error: %v", err)
	}
	if !runner.startCalled {
		t.Error("Start should have been called")
	}
}

func TestRunnerImplStartError(t *testing.T) {
	expectedErr := errors.New("start failed")
	runner := &testRunnerImpl{startErr: expectedErr}

	err := runner.Start(context.Background())
	if err != expectedErr {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

func TestRunnerImplStop(t *testing.T) {
	runner := &testRunnerImpl{}

	err := runner.Stop(context.Background())
	if err != nil {
		t.Errorf("Stop error: %v", err)
	}
	if !runner.stopCalled {
		t.Error("Stop should have been called")
	}
}

func TestRunnerImplStopError(t *testing.T) {
	expectedErr := errors.New("stop failed")
	runner := &testRunnerImpl{stopErr: expectedErr}

	err := runner.Stop(context.Background())
	if err != expectedErr {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

func TestRunnerImplLifecycle(t *testing.T) {
	runner := &testRunnerImpl{}

	err := runner.Start(context.Background())
	if err != nil {
		t.Errorf("Start error: %v", err)
	}

	err = runner.Stop(context.Background())
	if err != nil {
		t.Errorf("Stop error: %v", err)
	}

	if !runner.startCalled {
		t.Error("Start should have been called")
	}
	if !runner.stopCalled {
		t.Error("Stop should have been called")
	}
}
