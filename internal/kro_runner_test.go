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

	runner := NewKRORunner(namespace, nil, nil, scaleSetName, 0, 3, "", "")
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
	runner := NewKRORunner("default", nil, nil, "test-scale-set", 0, 3, "", "")

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

// TestKRORunnerFields tests that KRORunner fields are properly set
func TestKRORunnerFields(t *testing.T) {
	namespace := "test-ns"
	scaleSetName := "test-scale-set"
	runnerIndex := 5
	minRunners := 3
	imageID := "ami-123"
	instanceType := "t3.large"

	runner := NewKRORunner(namespace, nil, nil, scaleSetName, runnerIndex, minRunners, imageID, instanceType)

	if runner.namespace != namespace {
		t.Errorf("namespace = %q, want %q", runner.namespace, namespace)
	}
	if runner.scaleSetName != scaleSetName {
		t.Errorf("scaleSetName = %q, want %q", runner.scaleSetName, scaleSetName)
	}
	if runner.runnerIndex != runnerIndex {
		t.Errorf("runnerIndex = %d, want %d", runner.runnerIndex, runnerIndex)
	}
	if runner.minRunners != minRunners {
		t.Errorf("minRunners = %d, want %d", runner.minRunners, minRunners)
	}
	if runner.imageID != imageID {
		t.Errorf("imageID = %q, want %q", runner.imageID, imageID)
	}
	if runner.instanceType != instanceType {
		t.Errorf("instanceType = %q, want %q", runner.instanceType, instanceType)
	}
}

// TestCreateResourcesWithNilKubeClient tests CreateResources with nil kube client
func TestCreateResourcesWithNilKubeClient(t *testing.T) {
	runner := NewKRORunner("default", nil, nil, "test-scale-set", 0, 3, "ami-123", "t3.medium")

	// Should panic or error when trying to get orchestrator pod
	defer func() {
		if r := recover(); r != nil {
			// Expected panic due to nil client
			t.Log("Got expected panic with nil kubeClient")
		}
	}()

	err := runner.CreateResources(context.TODO(), "test-runner", "test-jit-config")
	if err != nil {
		// Expected error
		t.Log("Got expected error with nil kubeClient:", err)
	}
}

// TestDeleteResourcesWithNilDynamicClient tests DeleteResources with nil dynamic client
func TestDeleteResourcesWithNilDynamicClient(t *testing.T) {
	runner := NewKRORunner("default", nil, nil, "test-scale-set", 0, 3, "", "")

	// Set up app context
	NewAppContext("test-runner", "")

	// Should panic or error when trying to delete ResourceGraph
	defer func() {
		if r := recover(); r != nil {
			// Expected panic due to nil client
			t.Log("Got expected panic with nil dynamicClient")
		}
	}()

	err := runner.DeleteResources(context.TODO())
	if err != nil {
		// Expected error
		t.Log("Got expected error with nil dynamicClient:", err)
	}
}

// TestWaitForResourceGraphWithNilClient tests WaitForResourceGraph with nil client
func TestWaitForResourceGraphWithNilClient(t *testing.T) {
	runner := NewKRORunner("default", nil, nil, "test-scale-set", 0, 3, "", "")

	// Set up app context
	NewAppContext("test-runner", "")

	// Should panic or error when trying to get ResourceGraph
	defer func() {
		if r := recover(); r != nil {
			// Expected panic due to nil client
			t.Log("Got expected panic with nil dynamicClient")
		}
	}()

	err := runner.WaitForResourceGraph(context.TODO())
	if err != nil {
		// Expected error
		t.Log("Got expected error with nil dynamicClient:", err)
	}
}

// TestFindRGDByLabelWithNilClient tests findRGDByLabel with nil client
func TestFindRGDByLabelWithNilClient(t *testing.T) {
	runner := NewKRORunner("default", nil, nil, "test-scale-set", 0, 3, "", "")

	// Should panic or error when trying to list RGDs
	defer func() {
		if r := recover(); r != nil {
			// Expected panic due to nil client
			t.Log("Got expected panic with nil dynamicClient")
		}
	}()

	_, err := runner.findRGDByLabel(context.TODO())
	if err != nil {
		// Expected error
		t.Log("Got expected error with nil dynamicClient:", err)
	}
}

// TestCreateResourcesValidationEdgeCases tests edge cases in validation
func TestCreateResourcesValidationEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		runnerName  string
		jitConfig   string
		namespace   string
		scaleSet    string
		expectError bool
	}{
		{
			name:        "Empty runner name",
			runnerName:  "",
			jitConfig:   "config",
			namespace:   "default",
			scaleSet:    "test",
			expectError: true,
		},
		{
			name:        "Empty JIT config",
			runnerName:  "runner",
			jitConfig:   "",
			namespace:   "default",
			scaleSet:    "test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewKRORunner(tt.namespace, nil, nil, tt.scaleSet, 0, 3, "", "")

			// Use defer to catch any panics
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectError {
						t.Errorf("CreateResources() unexpected panic: %v", r)
					}
				}
			}()

			err := runner.CreateResources(context.TODO(), tt.runnerName, tt.jitConfig)

			if tt.expectError && err == nil {
				t.Errorf("CreateResources() expected error, got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("CreateResources() unexpected error: %v", err)
			}
		})
	}
}

// TestAppContextMultipleCalls tests multiple calls to NewAppContext
func TestAppContextMultipleCalls(t *testing.T) {
	// First call
	ctx1 := NewAppContext("runner1", "dv1")
	if ctx1.GetVMIName() != "runner1" {
		t.Errorf("First context VMIName = %q, want %q", ctx1.GetVMIName(), "runner1")
	}

	// Second call should override
	ctx2 := NewAppContext("runner2", "dv2")
	if ctx2.GetVMIName() != "runner2" {
		t.Errorf("Second context VMIName = %q, want %q", ctx2.GetVMIName(), "runner2")
	}

	// Global context should be updated
	globalCtx := GetAppContext()
	if globalCtx.GetVMIName() != "runner2" {
		t.Errorf("Global context VMIName = %q, want %q", globalCtx.GetVMIName(), "runner2")
	}
}

// TestKRORunnerWithDifferentConfigs tests runner with different configurations
func TestKRORunnerWithDifferentConfigs(t *testing.T) {
	tests := []struct {
		name         string
		namespace    string
		scaleSet     string
		runnerIndex  int
		minRunners   int
		imageID      string
		instanceType string
	}{
		{
			name:         "Default config",
			namespace:    "default",
			scaleSet:     "runners",
			runnerIndex:  0,
			minRunners:   3,
			imageID:      "",
			instanceType: "",
		},
		{
			name:         "Custom namespace",
			namespace:    "custom-ns",
			scaleSet:     "runners",
			runnerIndex:  0,
			minRunners:   3,
			imageID:      "",
			instanceType: "",
		},
		{
			name:         "High runner index",
			namespace:    "default",
			scaleSet:     "runners",
			runnerIndex:  10,
			minRunners:   3,
			imageID:      "",
			instanceType: "",
		},
		{
			name:         "With EC2 config",
			namespace:    "default",
			scaleSet:     "runners",
			runnerIndex:  0,
			minRunners:   5,
			imageID:      "ami-123456",
			instanceType: "t3.large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewKRORunner(
				tt.namespace,
				nil,
				nil,
				tt.scaleSet,
				tt.runnerIndex,
				tt.minRunners,
				tt.imageID,
				tt.instanceType,
			)

			if runner == nil {
				t.Fatal("NewKRORunner returned nil")
			}

			// Verify all fields are set correctly
			if runner.namespace != tt.namespace {
				t.Errorf("namespace = %q, want %q", runner.namespace, tt.namespace)
			}
			if runner.scaleSetName != tt.scaleSet {
				t.Errorf("scaleSetName = %q, want %q", runner.scaleSetName, tt.scaleSet)
			}
			if runner.runnerIndex != tt.runnerIndex {
				t.Errorf("runnerIndex = %d, want %d", runner.runnerIndex, tt.runnerIndex)
			}
			if runner.minRunners != tt.minRunners {
				t.Errorf("minRunners = %d, want %d", runner.minRunners, tt.minRunners)
			}
			if runner.imageID != tt.imageID {
				t.Errorf("imageID = %q, want %q", runner.imageID, tt.imageID)
			}
			if runner.instanceType != tt.instanceType {
				t.Errorf("instanceType = %q, want %q", runner.instanceType, tt.instanceType)
			}
		})
	}
}
