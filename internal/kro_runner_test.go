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

package runner

import (
	"context"
	"testing"
)

// TestToResourceName tests the toResourceName function
func TestToResourceName(t *testing.T) {
	tests := []struct {
		name     string
		kind     string
		expected string
	}{
		{
			name:     "PodRunner",
			kind:     "PodRunner",
			expected: "podrunners",
		},
		{
			name:     "VMRunner",
			kind:     "VMRunner",
			expected: "vmrunners",
		},
		{
			name:     "EC2Runner",
			kind:     "EC2Runner",
			expected: "ec2runners",
		},
		{
			name:     "Empty string",
			kind:     "",
			expected: "s",
		},
		{
			name:     "Already lowercase",
			kind:     "podrunner",
			expected: "podrunners",
		},
		{
			name:     "Mixed case",
			kind:     "MyCustomRunner",
			expected: "mycustomrunners",
		},
		{
			name:     "Single character",
			kind:     "A",
			expected: "as",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toResourceName(tt.kind)
			if result != tt.expected {
				t.Errorf("toResourceName(%q) = %q, want %q", tt.kind, result, tt.expected)
			}
		})
	}
}

// TestNewAppContext tests the NewAppContext function
func TestNewAppContext(t *testing.T) {
	tests := []struct {
		name           string
		vmiName        string
		dataVolumeName string
	}{
		{
			name:           "Valid context",
			vmiName:        "test-runner",
			dataVolumeName: "test-secret",
		},
		{
			name:           "Empty names",
			vmiName:        "",
			dataVolumeName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewAppContext(tt.vmiName, tt.dataVolumeName)
			if ctx == nil {
				t.Fatal("NewAppContext returned nil")
			}
			if ctx.GetVMIName() != tt.vmiName {
				t.Errorf("GetVMIName() = %q, want %q", ctx.GetVMIName(), tt.vmiName)
			}
			if ctx.GetDataVolumeName() != tt.dataVolumeName {
				t.Errorf("GetDataVolumeName() = %q, want %q", ctx.GetDataVolumeName(), tt.dataVolumeName)
			}
		})
	}
}

// TestGetAppContext tests the GetAppContext function
func TestGetAppContext(t *testing.T) {
	// Reset global state
	appContext = nil

	ctx := GetAppContext()
	if ctx == nil {
		t.Fatal("GetAppContext returned nil when appContext was nil")
	}

	// Set a context
	expectedVMI := "test-vmi"
	expectedDV := "test-dv"
	NewAppContext(expectedVMI, expectedDV)

	ctx = GetAppContext()
	if ctx.GetVMIName() != expectedVMI {
		t.Errorf("GetVMIName() = %q, want %q", ctx.GetVMIName(), expectedVMI)
	}
	if ctx.GetDataVolumeName() != expectedDV {
		t.Errorf("GetDataVolumeName() = %q, want %q", ctx.GetDataVolumeName(), expectedDV)
	}
}

// TestNewKRORunner tests the NewKRORunner constructor
func TestNewKRORunner(t *testing.T) {
	namespace := "default"
	scaleSetName := "test-scale-set"

	runner := NewKRORunner(namespace, nil, nil, scaleSetName)
	if runner == nil {
		t.Fatal("NewKRORunner returned nil")
	}

	if runner.namespace != namespace {
		t.Errorf("namespace = %q, want %q", runner.namespace, namespace)
	}

	if runner.scaleSetName != scaleSetName {
		t.Errorf("scaleSetName = %q, want %q", runner.scaleSetName, scaleSetName)
	}
}

// TestCreateResourcesValidation tests input validation for CreateResources
func TestCreateResourcesValidation(t *testing.T) {
	runner := NewKRORunner("default", nil, nil, "test-scale-set")

	tests := []struct {
		name        string
		runnerName  string
		jitConfig   string
		expectedErr error
	}{
		{
			name:        "Empty runner name",
			runnerName:  "",
			jitConfig:   "test-config",
			expectedErr: ErrEmptyRunnerName,
		},
		{
			name:        "Empty JIT config",
			runnerName:  "test-runner",
			jitConfig:   "",
			expectedErr: ErrEmptyJitConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runner.CreateResources(context.TODO(), tt.runnerName, tt.jitConfig)
			if err != tt.expectedErr {
				t.Errorf("CreateResources() error = %v, want %v", err, tt.expectedErr)
			}
		})
	}
}

// TestErrorConstants tests that error constants are defined
func TestErrorConstants(t *testing.T) {
	if ErrEmptyRunnerName == nil {
		t.Error("ErrEmptyRunnerName is nil")
	}
	if ErrEmptyJitConfig == nil {
		t.Error("ErrEmptyJitConfig is nil")
	}
	if ErrRunnerFailed == nil {
		t.Error("ErrRunnerFailed is nil")
	}

	// Verify error messages
	if ErrEmptyRunnerName.Error() != "empty runner name" {
		t.Errorf("ErrEmptyRunnerName message = %q, want %q", ErrEmptyRunnerName.Error(), "empty runner name")
	}
	if ErrEmptyJitConfig.Error() != "empty JIT config" {
		t.Errorf("ErrEmptyJitConfig message = %q, want %q", ErrEmptyJitConfig.Error(), "empty JIT config")
	}
	if ErrRunnerFailed.Error() != "runner execution failed" {
		t.Errorf("ErrRunnerFailed message = %q, want %q", ErrRunnerFailed.Error(), "runner execution failed")
	}
}

// TestAppContextMethods tests all AppContext methods
func TestAppContextMethods(t *testing.T) {
	vmiName := "test-vmi"
	dataVolumeName := "test-dv"

	ctx := NewAppContext(vmiName, dataVolumeName)

	// Test getters
	if ctx.GetVMIName() != vmiName {
		t.Errorf("GetVMIName() = %q, want %q", ctx.GetVMIName(), vmiName)
	}

	if ctx.GetDataVolumeName() != dataVolumeName {
		t.Errorf("GetDataVolumeName() = %q, want %q", ctx.GetDataVolumeName(), dataVolumeName)
	}

	// Test that context is accessible via GetAppContext
	retrievedCtx := GetAppContext()
	if retrievedCtx.GetVMIName() != vmiName {
		t.Errorf("GetAppContext().GetVMIName() = %q, want %q", retrievedCtx.GetVMIName(), vmiName)
	}
}

// TestRGDInfo tests the RGDInfo struct
func TestRGDInfo(t *testing.T) {
	info := &RGDInfo{
		Name:      "test-rgd",
		Namespace: "default",
		Kind:      "PodRunner",
	}

	if info.Name != "test-rgd" {
		t.Errorf("RGDInfo.Name = %q, want %q", info.Name, "test-rgd")
	}
	if info.Namespace != "default" {
		t.Errorf("RGDInfo.Namespace = %q, want %q", info.Namespace, "default")
	}
	if info.Kind != "PodRunner" {
		t.Errorf("RGDInfo.Kind = %q, want %q", info.Kind, "PodRunner")
	}
}
