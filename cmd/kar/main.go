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
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/electrocucaracha/kubevirt-actions-runner/cmd/kar/app"
	runner "github.com/electrocucaracha/kubevirt-actions-runner/internal"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const defaultCleanupTimeout = 5 * time.Minute

type buildInfo struct {
	gitCommit       string
	gitTreeModified string
	buildDate       string
	goVersion       string
}

func getBuildInfo() buildInfo {
	out := buildInfo{}

	if info, ok := debug.ReadBuildInfo(); ok {
		out.goVersion = info.GoVersion

		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				out.gitCommit = setting.Value
			case "vcs.time":
				out.buildDate = setting.Value
			case "vcs.modified":
				out.gitTreeModified = setting.Value
			}
		}
	}

	return out
}

func getCleanupTimeout() time.Duration {
	if val := os.Getenv("KAR_CLEANUP_TIMEOUT"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}

		log.Printf("Invalid KAR_CLEANUP_TIMEOUT value: %q, using default %s", val, defaultCleanupTimeout)
	}

	return defaultCleanupTimeout
}

func ensureValidCleanupContext(parent context.Context) (context.Context, context.CancelFunc) {
	if parent.Err() != nil {
		return context.WithTimeout(context.TODO(), getCleanupTimeout())
	}

	return context.WithTimeout(parent, getCleanupTimeout())
}

func main() {
	var (
		opts app.Opts
		err  error
	)

	buildInfo := getBuildInfo()
	log.Printf("starting kro-actions-runner\ncommit: %v\tmodified: %v\tdate: %v\tgo: %v\n",
		buildInfo.gitCommit, buildInfo.gitTreeModified, buildInfo.buildDate, buildInfo.goVersion)

	// Parse flags
	pflag.StringVar(&opts.ScaleSetName, "scale-set-name", os.Getenv("ACTIONS_RUNNER_SCALE_SET_NAME"), "Scale set name")
	pflag.StringVar(&opts.VMTemplate, "kubevirt-vm-template", "vm-template", "VM template")
	pflag.StringVar(&opts.RunnerName, "runner-name", os.Getenv("RUNNER_NAME"), "Runner name")
	pflag.StringVar(&opts.JitConfig, "actions-runner-input-jitconfig", os.Getenv("ACTIONS_RUNNER_INPUT_JITCONFIG"), "JIT config")
	pflag.Parse()

	// Get kubeconfig and namespace
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		log.Fatalf("error in namespace : %v\n", err)
	}
	if namespace == "" {
		namespace = "default"
	}

	// KRO mode only (KubeVirt support removed)
	log.Printf("Using KRO mode with scale-set-name: %s", opts.ScaleSetName)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		log.Fatalf("cannot obtain kubeconfig: %v\n", err)
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("cannot create dynamic client: %v\n", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("cannot create kubernetes client: %v\n", err)
	}

	r := runner.NewKRORunner(namespace, dynamicClient, kubeClient, opts.ScaleSetName)

	log.Printf("cleanup timeout is set to: %s", getCleanupTimeout())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()

		cleanupCtx, cancel := ensureValidCleanupContext(ctx)
		defer cancel()

		// Call DeleteResources
		if err := r.DeleteResources(cleanupCtx); err != nil {
			log.Println("cleanup failed:", err)
		}
	}()

	rootCmd := app.NewRootCommand(ctx, r, opts)

	if err := rootCmd.Execute(); err != nil && !errors.Is(errors.Cause(err), context.Canceled) {
		log.Println("execute command failed:", err)
	}
}
