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
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// TestInstallFlags tests the installFlags function
func TestInstallFlags(t *testing.T) {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)
	opts := &Opts{}

	installFlags(flags, opts)

	// Check that flags were registered
	expectedFlags := []string{"scale-set-name", "runner-name", "actions-runner-input-jitconfig"}
	for _, flagName := range expectedFlags {
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Flag %q was not registered", flagName)
		}
	}

	// Check short flags exist by looking up by shorthand
	scaleSetFlag := flags.Lookup("scale-set-name")
	if scaleSetFlag == nil || scaleSetFlag.Shorthand != "s" {
		t.Error("Short flag 's' for scale-set-name was not registered correctly")
	}
	runnerNameFlagCheck := flags.Lookup("runner-name")
	if runnerNameFlagCheck == nil || runnerNameFlagCheck.Shorthand != "r" {
		t.Error("Short flag 'r' for runner-name was not registered correctly")
	}
	jitConfigFlag := flags.Lookup("actions-runner-input-jitconfig")
	if jitConfigFlag == nil || jitConfigFlag.Shorthand != "c" {
		t.Error("Short flag 'c' for actions-runner-input-jitconfig was not registered correctly")
	}

	// Check default value for runner-name
	runnerNameFlag := flags.Lookup("runner-name")
	if runnerNameFlag.DefValue != "runner" {
		t.Errorf("runner-name default value = %q, want %q", runnerNameFlag.DefValue, "runner")
	}
}

// TestInitializeConfig tests the initializeConfig function
func TestInitializeConfig(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
	}

	opts := &Opts{}
	installFlags(cmd.Flags(), opts)

	err := initializeConfig(cmd)
	if err != nil {
		t.Errorf("initializeConfig() error = %v, want nil", err)
	}
}

// TestBindFlags tests the bindFlags function
func TestBindFlags(t *testing.T) {
	tests := []struct {
		name           string
		flagName       string
		flagValue      string
		viperValue     string
		flagChanged    bool
		expectedResult string
	}{
		{
			name:           "Flag not set, viper has value",
			flagName:       "test-flag",
			flagValue:      "",
			viperValue:     "viper-value",
			flagChanged:    false,
			expectedResult: "viper-value",
		},
		{
			name:           "Flag already set, viper has different value",
			flagName:       "test-flag",
			flagValue:      "flag-value",
			viperValue:     "viper-value",
			flagChanged:    true,
			expectedResult: "flag-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "test",
			}

			// Add a test flag
			var testValue string
			cmd.Flags().StringVar(&testValue, tt.flagName, tt.flagValue, "test flag")

			// Set flag as changed if needed
			if tt.flagChanged {
				_ = cmd.Flags().Set(tt.flagName, tt.flagValue)
			}

			// Create viper instance and set value
			v := viper.New()
			if tt.viperValue != "" {
				v.Set(tt.flagName, tt.viperValue)
			}

			// Bind flags
			bindFlags(cmd, v)

			// Check result
			result, err := cmd.Flags().GetString(tt.flagName)
			if err != nil {
				t.Fatalf("Failed to get flag value: %v", err)
			}

			if result != tt.expectedResult {
				t.Errorf("Flag value = %q, want %q", result, tt.expectedResult)
			}
		})
	}
}

// TestBindFlagsWithEnvVar tests bindFlags with environment variables
func TestBindFlagsWithEnvVar(t *testing.T) {
	// Set environment variable
	envKey := "TEST_FLAG"
	envValue := "env-value"
	_ = os.Setenv(envKey, envValue)
	defer func() { _ = os.Unsetenv(envKey) }()

	cmd := &cobra.Command{
		Use: "test",
	}

	var testValue string
	cmd.Flags().StringVar(&testValue, "test-flag", "", "test flag")

	// Create viper instance with env support (matching initializeConfig behavior)
	v := viper.New()
	v.AutomaticEnv()
	_ = v.BindEnv("test-flag", envKey)

	// Bind flags
	bindFlags(cmd, v)

	// Check result
	result, err := cmd.Flags().GetString("test-flag")
	if err != nil {
		t.Fatalf("Failed to get flag value: %v", err)
	}

	if result != envValue {
		t.Errorf("Flag value = %q, want %q (from env var)", result, envValue)
	}
}
