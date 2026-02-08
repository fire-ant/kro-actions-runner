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
	"fmt"
	"log"
	"time"

	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

const (
	// ACK Secrets Manager Secret CRD
	ackSecretGroup    = "secretsmanager.services.k8s.aws"
	ackSecretVersion  = "v1alpha1"
	ackSecretResource = "secrets"
	ackSecretKind     = "Secret"

	// Secret name format for JIT configs
	// Secrets are created by RGD with instance name: warmpool-{scaleSetName}-{idx}-jit
	jitSecretSuffix = "-jit"
	jitConfigKey    = "jitconfig"
)

var (
	// ACK Secret GVR
	ackSecretGVR = schema.GroupVersionResource{
		Group:    ackSecretGroup,
		Version:  ackSecretVersion,
		Resource: ackSecretResource,
	}
)

// SecretsManager manages JIT config secrets for warm pool instances
// Uses cloud-agnostic approach with ACK Secrets Manager controller
type SecretsManager struct {
	kubeClient    kubernetes.Interface
	dynamicClient dynamic.Interface
	namespace     string
	region        string
}

// NewSecretsManager creates a new secrets manager
func NewSecretsManager(kubeClient kubernetes.Interface, dynamicClient dynamic.Interface, namespace, region string) *SecretsManager {
	return &SecretsManager{
		kubeClient:    kubeClient,
		dynamicClient: dynamicClient,
		namespace:     namespace,
		region:        region,
	}
}

// UpdateJITSecret updates an existing JIT config secret with new JIT config
// Secrets are pre-created by RGD, this function just updates the content
// Uses ACK Secrets Manager controller for cloud-agnostic secret management:
// 1. Updates K8s Secret with JIT config data
// 2. ACK controller detects change and syncs to AWS Secrets Manager
// 3. Instance fetches updated secret from AWS Secrets Manager on boot
func (sm *SecretsManager) UpdateJITSecret(ctx context.Context, instanceName, jitConfig string) error {
	secretName := instanceName + jitSecretSuffix // e.g., warmpool-test-runners-0-jit
	awsSecretName := fmt.Sprintf("/kro/runners/jit-config/%s", instanceName)

	log.Printf("Updating JIT secret for instance %s", instanceName)

	// Get existing K8s Secret (created by RGD)
	k8sSecret, err := sm.kubeClient.CoreV1().Secrets(sm.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to get K8s secret %s (should be created by RGD)", secretName)
	}

	// Update JIT config data
	if k8sSecret.StringData == nil {
		k8sSecret.StringData = make(map[string]string)
	}
	k8sSecret.StringData[jitConfigKey] = jitConfig

	// Update labels with timestamp
	if k8sSecret.Labels == nil {
		k8sSecret.Labels = make(map[string]string)
	}
	k8sSecret.Labels["kro.run/last-updated"] = time.Now().Format(time.RFC3339)

	// Update the secret
	_, err = sm.kubeClient.CoreV1().Secrets(sm.namespace).Update(ctx, k8sSecret, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to update K8s secret %s", secretName)
	}

	log.Printf("✓ Updated K8s Secret: %s", secretName)
	log.Printf("✓ ACK controller will sync to AWS Secrets Manager: %s", awsSecretName)
	log.Printf("✓ JIT secret updated successfully for instance: %s", instanceName)

	return nil
}

// ClearJITSecret clears the JIT config from a secret after job completion
// Secrets persist (created by RGD), this just clears the content for reuse
// ACK controller syncs the empty value to AWS Secrets Manager
func (sm *SecretsManager) ClearJITSecret(ctx context.Context, instanceName string) error {
	secretName := instanceName + jitSecretSuffix // e.g., warmpool-test-runners-0-jit
	awsSecretName := fmt.Sprintf("/kro/runners/jit-config/%s", instanceName)

	log.Printf("Clearing JIT secret for instance %s", instanceName)

	// Get existing K8s Secret
	k8sSecret, err := sm.kubeClient.CoreV1().Secrets(sm.namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("Secret %s not found - may have been deleted", secretName)
			return nil
		}
		return errors.Wrapf(err, "failed to get K8s secret %s", secretName)
	}

	// Clear JIT config data
	if k8sSecret.StringData == nil {
		k8sSecret.StringData = make(map[string]string)
	}
	k8sSecret.StringData[jitConfigKey] = "" // Empty string

	// Update labels with cleared timestamp
	if k8sSecret.Labels == nil {
		k8sSecret.Labels = make(map[string]string)
	}
	k8sSecret.Labels["kro.run/last-cleared"] = time.Now().Format(time.RFC3339)
	delete(k8sSecret.Labels, "kro.run/last-updated")

	// Update the secret
	_, err = sm.kubeClient.CoreV1().Secrets(sm.namespace).Update(ctx, k8sSecret, metav1.UpdateOptions{})
	if err != nil {
		return errors.Wrapf(err, "failed to clear K8s secret %s", secretName)
	}

	log.Printf("✓ Cleared K8s Secret: %s", secretName)
	log.Printf("✓ ACK controller will sync empty value to AWS: %s", awsSecretName)
	log.Printf("✓ JIT secret cleared successfully for instance: %s", instanceName)

	return nil
}

// GetJITSecretName returns the AWS Secrets Manager secret name for an instance
// This is the name the EC2 instance will use to fetch the JIT config
func (sm *SecretsManager) GetJITSecretName(instanceName string) string {
	return fmt.Sprintf("/kro/runners/jit-config/%s", instanceName)
}
