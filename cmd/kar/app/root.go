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
	"log"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewRootCommand(ctx context.Context, r interface{}, opts Opts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kar",
		Short: "Tool that creates a GitHub Self-Host runner with KRO or Kubevirt",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return initializeConfig(cmd)
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(ctx, r, opts)
		},
	}

	installFlags(cmd.Flags(), &opts)

	return cmd
}

func run(ctx context.Context, r interface{}, opts Opts) error {
	// KRO mode (only mode supported)
	kroRunner, ok := r.(interface {
		CreateResources(ctx context.Context, runnerName string, jitConfig string) error
		WaitForResourceGraph(ctx context.Context) error
		DeleteResources(ctx context.Context) error
	})
	if !ok {
		return errors.New("runner does not implement required KRO interface")
	}

	if err := kroRunner.CreateResources(ctx, opts.RunnerName, opts.JitConfig); err != nil {
		return errors.Wrap(err, "fail to create resources")
	}

	log.Println("ResourceGraph runner resources created successfully")

	if err := kroRunner.WaitForResourceGraph(ctx); err != nil {
		return errors.Wrap(err, "fail to wait for resources")
	}

	log.Println("ResourceGraph runner completed successfully")

	if err := kroRunner.DeleteResources(ctx); err != nil {
		return errors.Wrap(err, "fail to delete resources")
	}

	log.Println("ResourceGraph runner deleted successfully")

	return nil
}
