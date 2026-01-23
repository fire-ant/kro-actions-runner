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

package main

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestGetCleanupTimeout tests the getCleanupTimeout function
func TestGetCleanupTimeout(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected time.Duration
	}{
		{
			name:     "Valid duration string",
			envValue: "10m",
			expected: 10 * time.Minute,
		},
		{
			name:     "Valid duration in seconds",
			envValue: "30s",
			expected: 30 * time.Second,
		},
		{
			name:     "Valid duration in hours",
			envValue: "2h",
			expected: 2 * time.Hour,
		},
		{
			name:     "Invalid duration falls back to default",
			envValue: "invalid",
			expected: defaultCleanupTimeout,
		},
		{
			name:     "Empty env var falls back to default",
			envValue: "",
			expected: defaultCleanupTimeout,
		},
		{
			name:     "Zero duration",
			envValue: "0s",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("KAR_CLEANUP_TIMEOUT", tt.envValue)
			} else {
				os.Unsetenv("KAR_CLEANUP_TIMEOUT")
			}
			defer os.Unsetenv("KAR_CLEANUP_TIMEOUT")

			result := getCleanupTimeout()
			if result != tt.expected {
				t.Errorf("getCleanupTimeout() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGetBuildInfo tests the getBuildInfo function
func TestGetBuildInfo(t *testing.T) {
	// This function reads from runtime debug info
	// We can't control the build info in tests, but we can verify it doesn't panic
	info := getBuildInfo()

	// Just verify we get a struct back (fields may be empty in test environment)
	if info.goVersion == "" {
		t.Log("goVersion is empty (expected in test environment)")
	}

	// The function should at least return a valid struct
	_ = info.gitCommit
	_ = info.gitTreeModified
	_ = info.buildDate
}

// TestEnsureValidCleanupContext tests the ensureValidCleanupContext function
func TestEnsureValidCleanupContext(t *testing.T) {
	tests := []struct {
		name          string
		parentContext context.Context
		expectTimeout bool
		minDuration   time.Duration
	}{
		{
			name:          "Valid parent context",
			parentContext: context.Background(),
			expectTimeout: true,
			minDuration:   defaultCleanupTimeout - time.Second,
		},
		{
			name: "Cancelled parent context",
			parentContext: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			expectTimeout: true,
			minDuration:   defaultCleanupTimeout - time.Second,
		},
		{
			name: "Context with expired deadline",
			parentContext: func() context.Context {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
				defer cancel()
				return ctx
			}(),
			expectTimeout: true,
			minDuration:   defaultCleanupTimeout - time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := ensureValidCleanupContext(tt.parentContext)
			defer cancel()

			if ctx == nil {
				t.Fatal("ensureValidCleanupContext returned nil context")
			}

			// Check that we have a deadline
			deadline, ok := ctx.Deadline()
			if tt.expectTimeout && !ok {
				t.Error("Expected context with deadline, but got none")
			}

			if ok {
				duration := time.Until(deadline)
				if duration < tt.minDuration {
					t.Errorf("Timeout duration %v is less than minimum expected %v", duration, tt.minDuration)
				}
			}
		})
	}
}
