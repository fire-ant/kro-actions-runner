/*
Copyright © 2024

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"context"
	"errors"
	"testing"
)

// mockRunner implements the required interface for testing
type mockRunner struct {
	createErr error
	waitErr   error
	deleteErr error
	called    struct {
		create bool
		wait   bool
		delete bool
	}
}

func (m *mockRunner) CreateResources(_ context.Context, _ string, _ string) error {
	m.called.create = true
	return m.createErr
}

func (m *mockRunner) WaitForResourceGraph(_ context.Context) error {
	m.called.wait = true
	return m.waitErr
}

func (m *mockRunner) DeleteResources(_ context.Context) error {
	m.called.delete = true
	return m.deleteErr
}

// TestNewRootCommand tests the NewRootCommand function
func TestNewRootCommand(t *testing.T) {
	ctx := context.Background()
	runner := &mockRunner{}
	opts := Opts{
		ScaleSetName: "test-scale-set",
		RunnerName:   "test-runner",
		JitConfig:    "test-jit-config",
	}

	cmd := NewRootCommand(ctx, runner, opts)
	if cmd == nil {
		t.Fatal("NewRootCommand returned nil")
	}

	if cmd.Use != "kar" {
		t.Errorf("cmd.Use = %q, want %q", cmd.Use, "kar")
	}

	if cmd.Short == "" {
		t.Error("cmd.Short is empty")
	}
}

// TestRunSuccess tests successful run execution
func TestRunSuccess(t *testing.T) {
	ctx := context.Background()
	runner := &mockRunner{}
	opts := Opts{
		RunnerName: "test-runner",
		JitConfig:  "test-jit-config",
	}

	err := run(ctx, runner, opts)
	if err != nil {
		t.Errorf("run() error = %v, want nil", err)
	}

	if !runner.called.create {
		t.Error("CreateResources was not called")
	}
	if !runner.called.wait {
		t.Error("WaitForResourceGraph was not called")
	}
	if !runner.called.delete {
		t.Error("DeleteResources was not called")
	}
}

// TestRunCreateError tests run with CreateResources error
func TestRunCreateError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("create error")
	runner := &mockRunner{
		createErr: expectedErr,
	}
	opts := Opts{
		RunnerName: "test-runner",
		JitConfig:  "test-jit-config",
	}

	err := run(ctx, runner, opts)
	if err == nil {
		t.Fatal("run() error = nil, want error")
	}

	if !runner.called.create {
		t.Error("CreateResources was not called")
	}
	if runner.called.wait {
		t.Error("WaitForResourceGraph should not be called after CreateResources error")
	}
	if runner.called.delete {
		t.Error("DeleteResources should not be called after CreateResources error")
	}
}

// TestRunWaitError tests run with WaitForResourceGraph error
func TestRunWaitError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("wait error")
	runner := &mockRunner{
		waitErr: expectedErr,
	}
	opts := Opts{
		RunnerName: "test-runner",
		JitConfig:  "test-jit-config",
	}

	err := run(ctx, runner, opts)
	if err == nil {
		t.Fatal("run() error = nil, want error")
	}

	if !runner.called.create {
		t.Error("CreateResources was not called")
	}
	if !runner.called.wait {
		t.Error("WaitForResourceGraph was not called")
	}
	if runner.called.delete {
		t.Error("DeleteResources should not be called after WaitForResourceGraph error")
	}
}

// TestRunDeleteError tests run with DeleteResources error
func TestRunDeleteError(t *testing.T) {
	ctx := context.Background()
	expectedErr := errors.New("delete error")
	runner := &mockRunner{
		deleteErr: expectedErr,
	}
	opts := Opts{
		RunnerName: "test-runner",
		JitConfig:  "test-jit-config",
	}

	err := run(ctx, runner, opts)
	if err == nil {
		t.Fatal("run() error = nil, want error")
	}

	if !runner.called.create {
		t.Error("CreateResources was not called")
	}
	if !runner.called.wait {
		t.Error("WaitForResourceGraph was not called")
	}
	if !runner.called.delete {
		t.Error("DeleteResources was not called")
	}
}

// TestRunInvalidRunner tests run with invalid runner type
func TestRunInvalidRunner(t *testing.T) {
	ctx := context.Background()
	invalidRunner := "not a runner" // String doesn't implement the required interface
	opts := Opts{
		RunnerName: "test-runner",
		JitConfig:  "test-jit-config",
	}

	err := run(ctx, invalidRunner, opts)
	if err == nil {
		t.Fatal("run() error = nil, want error for invalid runner")
	}

	expectedMsg := "runner does not implement required KRO interface"
	if err.Error() != expectedMsg {
		t.Errorf("run() error message = %q, want %q", err.Error(), expectedMsg)
	}
}
