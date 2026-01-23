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
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/ptr"
)

const (
	// Label for RGD discovery - matches scale set name
	rgdLabelKey = "actions.github.com/scale-set-name"

	// Annotation to store runner metadata
	runnerMetadataAnnotation = "actions.github.com/runner-metadata"
)

// Errors
var (
	ErrEmptyRunnerName = errors.New("empty runner name")
	ErrEmptyJitConfig  = errors.New("empty JIT config")
	ErrRunnerFailed    = errors.New("runner execution failed")
)

// AppContext stores runner context for cleanup
type AppContext struct {
	vmiName        string
	dataVolumeName string
}

var appContext *AppContext

// NewAppContext creates a new app context
func NewAppContext(vmiName, dataVolumeName string) *AppContext {
	appContext = &AppContext{
		vmiName:        vmiName,
		dataVolumeName: dataVolumeName,
	}
	return appContext
}

// GetAppContext returns the app context
func GetAppContext() *AppContext {
	if appContext == nil {
		appContext = &AppContext{}
	}
	return appContext
}

// GetVMIName returns the VMI name (repurposed for runner name)
func (ac *AppContext) GetVMIName() string {
	return ac.vmiName
}

// GetDataVolumeName returns the data volume name (repurposed for secret name)
func (ac *AppContext) GetDataVolumeName() string {
	return ac.dataVolumeName
}

// RGDInfo holds information about a discovered ResourceGraphDefinition
type RGDInfo struct {
	Name      string
	Namespace string
	Kind      string // The Kind from RGD schema (e.g., "PodRunner", "VMRunner")
}

// Runner interface for KRO-based runners
type Runner interface {
	CreateResources(ctx context.Context, runnerName string, jitConfig string) error
	WaitForResourceGraph(ctx context.Context) error
	DeleteResources(ctx context.Context) error
}

// KRORunner manages runner lifecycle using KRO ResourceGraph instances
type KRORunner struct {
	dynamicClient dynamic.Interface
	kubeClient    kubernetes.Interface
	namespace     string
	scaleSetName  string
}

var _ Runner = (*KRORunner)(nil)

// NewKRORunner creates a new KRO-based runner
func NewKRORunner(namespace string, dynamicClient dynamic.Interface, kubeClient kubernetes.Interface, scaleSetName string) *KRORunner {
	return &KRORunner{
		namespace:     namespace,
		dynamicClient: dynamicClient,
		kubeClient:    kubeClient,
		scaleSetName:  scaleSetName,
	}
}

// findRGDByLabel discovers an RGD by matching the actions.github.com/scale-set-name label
func (r *KRORunner) findRGDByLabel(ctx context.Context) (*RGDInfo, error) {
	log.Printf("Discovering RGD with label %s=%s", rgdLabelKey, r.scaleSetName)

	rgdGVR := schema.GroupVersionResource{
		Group:    "kro.run",
		Version:  "v1alpha1",
		Resource: "resourcegraphdefinitions",
	}

	// List all RGDs with matching label
	rgdList, err := r.dynamicClient.Resource(rgdGVR).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", rgdLabelKey, r.scaleSetName),
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list RGDs")
	}

	if len(rgdList.Items) == 0 {
		return nil, fmt.Errorf("no RGD found with label %s=%s", rgdLabelKey, r.scaleSetName)
	}

	if len(rgdList.Items) > 1 {
		return nil, fmt.Errorf("multiple RGDs found with label %s=%s, expected exactly one", rgdLabelKey, r.scaleSetName)
	}

	rgd := &rgdList.Items[0]

	// Extract the Kind from RGD schema
	kind, found, err := unstructured.NestedString(rgd.Object, "spec", "schema", "kind")
	if err != nil || !found {
		return nil, fmt.Errorf("RGD %s missing spec.schema.kind", rgd.GetName())
	}

	info := &RGDInfo{
		Name:      rgd.GetName(),
		Namespace: rgd.GetNamespace(),
		Kind:      kind,
	}

	log.Printf("Discovered RGD: name=%s, namespace=%s, kind=%s", info.Name, info.Namespace, info.Kind)
	return info, nil
}

// CreateResources creates a ResourceGraph instance for the runner
func (r *KRORunner) CreateResources(ctx context.Context, runnerName string, jitConfig string) error {
	if len(runnerName) == 0 {
		return ErrEmptyRunnerName
	}

	if len(jitConfig) == 0 {
		return ErrEmptyJitConfig
	}

	// Get the orchestrator pod to set as owner reference
	orchestratorPod, err := r.kubeClient.CoreV1().Pods(r.namespace).Get(ctx, runnerName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to get orchestrator pod for owner reference")
	}

	// Discover the RGD
	rgdInfo, err := r.findRGDByLabel(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to discover RGD")
	}

	// Note: We don't create a JIT secret - ARC already created one with the runner name
	// The RGD will reference the ARC-created secret directly
	log.Printf("Using ARC-created secret: %s", runnerName)

	// Create ResourceGraph instance
	rgInstance := &unstructured.Unstructured{}
	rgInstance.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "kro.run",
		Version: "v1alpha1",
		Kind:    rgdInfo.Kind,
	})
	rgInstance.SetName(runnerName)
	rgInstance.SetNamespace(r.namespace)

	// Set metadata annotation with runner info
	metadata := map[string]interface{}{
		"runnerName":       runnerName,
		"scaleSetName":     r.scaleSetName,
		"jitConfigSecret":  runnerName, // ARC creates secret with same name as runner
		"createdTimestamp": time.Now().Format(time.RFC3339),
	}
	metadataJSON, _ := json.Marshal(metadata)

	annotations := map[string]string{
		runnerMetadataAnnotation: string(metadataJSON),
	}
	rgInstance.SetAnnotations(annotations)

	// Set labels for tracking
	labels := map[string]string{
		"actions.github.com/scale-set-name": r.scaleSetName,
		"kro.run/runner-name":               runnerName,
	}
	rgInstance.SetLabels(labels)

	// Set owner reference to orchestrator pod for garbage collection
	rgInstance.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion: "v1",
			Kind:       "Pod",
			Name:       orchestratorPod.Name,
			UID:        orchestratorPod.UID,
			Controller: ptr.To(false),
		},
	})

	// Build the spec - just pass the runner name
	// The RGD will use this to reference the ARC-created secret
	spec := map[string]interface{}{
		"runnerName": runnerName,
	}

	rgInstance.Object["spec"] = spec

	log.Printf("Creating ResourceGraph instance: kind=%s, name=%s", rgdInfo.Kind, runnerName)

	// Create the RG instance
	rgGVR := schema.GroupVersionResource{
		Group:    "kro.run",
		Version:  "v1alpha1",
		Resource: toResourceName(rgdInfo.Kind), // PodRunner -> podrunners
	}

	_, err = r.dynamicClient.Resource(rgGVR).Namespace(r.namespace).Create(ctx, rgInstance, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to create ResourceGraph instance")
	}

	log.Printf("ResourceGraph instance created successfully: %s", runnerName)

	// Store in app context for cleanup
	// Note: No separate secret to track - ARC manages the secret lifecycle
	NewAppContext(runnerName, "")

	return nil
}

// WaitForResourceGraph watches the ResourceGraph instance until completion
func (r *KRORunner) WaitForResourceGraph(ctx context.Context) error {
	appCtx := GetAppContext()
	runnerName := appCtx.GetVMIName() // Reusing VMI name field for runner name

	log.Printf("Watching ResourceGraph instance: %s", runnerName)

	// First, discover the RGD to get the Kind
	rgdInfo, err := r.findRGDByLabel(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to discover RGD for watching")
	}

	rgGVR := schema.GroupVersionResource{
		Group:    "kro.run",
		Version:  "v1alpha1",
		Resource: toResourceName(rgdInfo.Kind),
	}

	// Watch the RG instance
	watcher, err := r.dynamicClient.Resource(rgGVR).Namespace(r.namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", runnerName),
	})
	if err != nil {
		return errors.Wrap(err, "failed to watch ResourceGraph instance")
	}
	defer watcher.Stop()

	for {
		select {
		case event := <-watcher.ResultChan():
			if event.Type == watch.Error {
				return fmt.Errorf("watch error: %v", event.Object)
			}

			rg, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				continue
			}

			// Get the state from status
			state, found, err := unstructured.NestedString(rg.Object, "status", "state")
			if err != nil || !found {
				log.Printf("ResourceGraph %s status not yet available", runnerName)
				continue
			}

			log.Printf("ResourceGraph %s state: %s", runnerName, state)

			switch state {
			case "ACTIVE":
				// Check if resources are ready (which means Pod completed due to readyWhen)
				conditions, found, err := unstructured.NestedSlice(rg.Object, "status", "conditions")
				if err == nil && found {
					for _, cond := range conditions {
						condMap, ok := cond.(map[string]interface{})
						if !ok {
							continue
						}
						condType, _ := condMap["type"].(string)
						condStatus, _ := condMap["status"].(string)

						// ResourcesReady means all readyWhen conditions are met (Pod completed)
						if condType == "ResourcesReady" && condStatus == "True" {
							log.Printf("ResourceGraph %s resources ready - runner completed", runnerName)

							// Check if it was success or failure by looking at pod status
							podStatus, found, err := unstructured.NestedMap(rg.Object, "status", "resources", "runnerPod", "status")
							if err == nil && found {
								phase, _ := podStatus["phase"].(string)
								if phase == "Succeeded" {
									log.Printf("Runner pod completed successfully")
									return nil
								} else if phase == "Failed" {
									log.Printf("Runner pod failed")
									return ErrRunnerFailed
								}
							}

							// Fallback: if we can't get pod status, assume success since ResourcesReady is true
							log.Printf("Runner completed (unable to determine pod phase, assuming success)")
							return nil
						}
					}
				}

			case "FAILED":
				log.Printf("ResourceGraph %s failed", runnerName)
				return ErrRunnerFailed

			case "DELETED":
				log.Printf("ResourceGraph %s deleted", runnerName)
				return nil
			}

		case <-ctx.Done():
			log.Printf("Context cancelled, stopping watch")
			return ctx.Err()
		}
	}
}

// DeleteResources cleans up the ResourceGraph instance and secret
func (r *KRORunner) DeleteResources(ctx context.Context) error {
	appCtx := GetAppContext()
	runnerName := appCtx.GetVMIName()
	secretName := appCtx.GetDataVolumeName() // Reusing DataVolume name field for secret name

	log.Printf("Cleaning up ResourceGraph resources for runner: %s", runnerName)

	// Discover the RGD to get the Kind
	rgdInfo, err := r.findRGDByLabel(ctx)
	if err != nil {
		log.Printf("Warning: failed to discover RGD for cleanup: %v", err)
		// Continue with cleanup anyway
	}

	if rgdInfo != nil {
		// Delete the ResourceGraph instance
		rgGVR := schema.GroupVersionResource{
			Group:    "kro.run",
			Version:  "v1alpha1",
			Resource: toResourceName(rgdInfo.Kind),
		}

		if err := r.dynamicClient.Resource(rgGVR).Namespace(r.namespace).Delete(
			ctx, runnerName, metav1.DeleteOptions{}); err != nil {
			if !k8serrors.IsNotFound(err) {
				log.Printf("Failed to delete ResourceGraph instance %s: %v", runnerName, err)
			}
		} else {
			log.Printf("Deleted ResourceGraph instance: %s", runnerName)
		}
	}

	// Delete the JIT secret
	if len(secretName) > 0 {
		if err := r.kubeClient.CoreV1().Secrets(r.namespace).Delete(
			ctx, secretName, metav1.DeleteOptions{}); err != nil {
			if !k8serrors.IsNotFound(err) {
				log.Printf("Failed to delete JIT secret %s: %v", secretName, err)
			}
		} else {
			log.Printf("Deleted JIT secret: %s", secretName)
		}
	}

	return nil
}

// createJitSecret creates a Kubernetes Secret with the JIT config
func (r *KRORunner) createJitSecret(ctx context.Context, secretName, runnerName, jitConfig string) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: r.namespace,
			Labels: map[string]string{
				"actions.github.com/scale-set-name": r.scaleSetName,
				"kro.run/runner-name":               runnerName,
			},
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			".jitconfig": jitConfig,
		},
	}

	_, err := r.kubeClient.CoreV1().Secrets(r.namespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to create secret")
	}

	return nil
}

// toResourceName converts Kind to resource name (PodRunner -> podrunners)
func toResourceName(kind string) string {
	// Simple lowercase + s pluralization
	// For more complex cases, consider using inflection library
	lower := strings.ToLower(kind)
	return lower + "s"
}
