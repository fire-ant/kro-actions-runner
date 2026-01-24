# Integration Tests

This directory contains integration tests for kro-actions-runner using [kuttl](https://kuttl.dev/) and Kind clusters.

## Overview

The integration tests validate (in priority order):

1. **kar Binary Integration** ⭐ **PRIMARY** (`e2e/kar-integration/`) - Actual kar binary execution (simulates ARC)
2. **Pod Runner Flow** (`e2e/pod-runner/`) - End-to-end pod-runner provisioning
3. **RGD Discovery** (`e2e/rgd-discovery/`) - Label-based ResourceGraphDefinition selection
4. **RBAC Validation** (`e2e/rbac-validation/`) - Service account permissions

**kar-integration is the primary test** as it validates the actual kar binary logic without requiring GitHub API integration.

## Prerequisites

Install required tools using mise:

```bash
mise install
```

This installs:
- `kubectl` - Kubernetes CLI
- `kind` - Local Kubernetes clusters
- `helm` - Kubernetes package manager
- `kuttl` - Kubernetes test framework

## Quick Start

Run the main kar integration test (fastest validation):

```bash
mise run test:integration
```

Or run all integration tests:

```bash
mise run test:e2e
```

This will:
1. Build the `kar` binary
2. Build the Docker image
3. Create a Kind cluster
4. Install KRO and ARC
5. Load the image into the cluster
6. Run all kuttl tests (kar-integration first)
7. Clean up

## Running Specific Tests

Run only the kar integration test (primary test):

```bash
mise run test:e2e:kar
```

Run only the pod-runner test:

```bash
mise run test:e2e:pod-runner
```

Run kuttl directly with custom options (from project root):

```bash
# Run from project root directory
kubectl kuttl test --config test/e2e/kuttl-test.yaml --test pod-runner
```

**Note:** All kuttl commands must be run from the project root directory, as test paths are configured relative to the project root.

## Cluster Management

### Create a Test Cluster

```bash
mise run cluster:create
```

### Set Up Cluster (Install KRO/ARC)

```bash
mise run cluster:setup
```

### View Cluster Status

```bash
mise run cluster:status
```

### View Component Logs

```bash
mise run cluster:logs
```

### Reset Cluster

Delete and recreate from scratch:

```bash
mise run cluster:reset
```

### Delete Cluster

```bash
mise run cluster:delete
```

## Directory Structure

```
test/
├── e2e/                           # Kuttl end-to-end tests
│   ├── kuttl-test.yaml            # Main TestSuite configuration
│   ├── pod-runner/                # Pod runner test case
│   │   ├── 00-namespace.yaml      # Setup namespace
│   │   ├── 01-rbac.yaml           # Apply RBAC
│   │   ├── 01-assert.yaml         # Assert RBAC created
│   │   ├── 02-rgd.yaml            # Apply RGD
│   │   ├── 02-assert.yaml         # Assert RGD created
│   │   ├── 03-secret.yaml         # Create test secret
│   │   ├── 03-assert.yaml         # Assert secret created
│   │   ├── 04-resourcegraph.yaml  # Create ResourceGraph
│   │   ├── 04-assert.yaml         # Assert Pod created
│   │   ├── 05-cleanup.yaml        # Delete ResourceGraph
│   │   └── 05-assert.yaml         # Assert cleanup
│   ├── rgd-discovery/             # RGD label discovery test
│   │   ├── 00-namespace.yaml
│   │   ├── 01-rgds.yaml           # Apply multiple RGDs
│   │   ├── 01-assert.yaml         # Assert all exist
│   │   ├── 02-query-labels.yaml   # Query by labels
│   │   └── 02-assert.yaml
│   ├── rbac-validation/           # RBAC permissions test
│   │   ├── 00-namespace.yaml
│   │   ├── 01-rbac.yaml           # Apply RBAC
│   │   ├── 01-assert.yaml         # Assert resources
│   │   ├── 02-check-permissions.yaml  # Test permissions
│   │   └── 02-assert.yaml
│   └── kar-integration/           # kar binary integration test
│       ├── 00-namespace.yaml
│       ├── 01-rbac.yaml           # Apply RBAC
│       ├── 01-assert.yaml
│       ├── 02-rgd.yaml            # Apply RGD
│       ├── 02-assert.yaml
│       ├── 03-secret.yaml         # Create test secret
│       ├── 03-assert.yaml
│       ├── 04-kar-pod.yaml        # Run kar binary in pod
│       ├── 04-assert.yaml         # Assert kar pod succeeded
│       ├── 05-assert-resources.yaml  # Assert ResourceGraph and runner pod
│       ├── 06-cleanup.yaml        # Delete ResourceGraph
│       └── 06-assert.yaml         # Assert cleanup
├── fixtures/                      # Shared test fixtures
│   ├── rbac/
│   │   └── kro-runner-rbac.yaml
│   ├── rgds/
│   │   ├── pod-runner-rgd.yaml
│   │   └── test-rgd-multi.yaml
│   ├── secrets/
│   │   └── test-jit-config.yaml
│   └── kind/
│       └── kind-config.yaml
├── scripts/                       # Test helper scripts
│   ├── install-kro.sh             # Install KRO via Helm
│   ├── install-arc.sh             # Install ARC via Helm
│   ├── load-image.sh              # Load image into Kind
│   └── wait-for-ready.sh          # Wait for resource ready
└── README.md                      # This file
```

## Test Cases

### 1. Pod Runner Test

**Purpose:** Validates the complete pod-runner provisioning flow.

**Steps:**
1. Create namespace
2. Apply RBAC resources
3. Apply Pod Runner RGD
4. Create test JIT config secret
5. Create PodRunner ResourceGraph instance
6. Assert runner Pod is created and running
7. Clean up ResourceGraph
8. Assert resources are deleted

**What it tests:**
- KRO ResourceGraph creation
- Pod provisioning via RGD
- Secret reference resolution
- Resource lifecycle management

### 2. RGD Discovery Test

**Purpose:** Validates label-based RGD selection.

**Steps:**
1. Apply multiple RGDs with different labels
2. Verify all RGDs exist
3. Query RGDs by label selector
4. Verify correct RGD is selected

**What it tests:**
- Multiple RGDs can coexist
- Label-based querying works
- Correct RGD selection by scale-set-name

### 3. RBAC Validation Test

**Purpose:** Validates service account permissions.

**Steps:**
1. Apply RBAC resources
2. Verify ServiceAccount, Role, ClusterRole exist
3. Verify RoleBinding and ClusterRoleBinding
4. Test permissions using `kubectl auth can-i`

**What it tests:**
- RBAC resources are created correctly
- ServiceAccount has required permissions
- Namespace-scoped and cluster-scoped access

### 4. kar Binary Integration Test

**Purpose:** Validates the actual kar binary by running it in a pod (simulates ARC behavior).

**Steps:**
1. Apply RBAC resources
2. Apply Pod Runner RGD
3. Create mock JIT config secret
4. Run kar binary in a pod with environment variables (simulates ARC)
5. Assert kar pod completes successfully
6. Assert ResourceGraph was created by kar
7. Assert runner Pod was provisioned by KRO
8. Clean up ResourceGraph

**What it tests:**
- Actual kar binary execution (not just KRO)
- RGD discovery logic by label matching
- ResourceGraph creation by kar
- Pod provisioning through the full stack
- End-to-end integration without GitHub API

**Key difference:** This tests the actual kar binary code path, whereas pod-runner test directly creates ResourceGraphs. This is the most realistic test without needing GitHub API access.

## Configuration

Environment variables (set in `mise.toml`):

- `KIND_CLUSTER_NAME` - Kind cluster name (default: `kro-test`)
- `KUTTL_NAMESPACE` - Default namespace (default: `arc-runners`)
- `CONTROLLER_NS` - ARC controller namespace (default: `arc-systems`)

## Debugging

### View Test Resources

```bash
# List all RGDs
kubectl get rgd

# List ResourceGraphs
kubectl get resourcegraphs -A

# List pods
kubectl get pods -n arc-runners

# Describe a specific resource
kubectl describe rgd pod-runner
```

### Access Test Logs

```bash
# KRO logs
kubectl logs -n kro-system -l app=kro --tail=100

# ARC logs
kubectl logs -n arc-systems -l app.kubernetes.io/name=gha-runner-scale-set-controller --tail=100

# Test pod logs
kubectl logs -n arc-runners test-runner-jit-job
```

### Keep Cluster After Test Failure

Edit `test/e2e/kuttl-test.yaml`:

```yaml
skipDelete: true
```

This keeps resources after tests for inspection.

### Run Single Test Step

```bash
# Apply a specific test step manually
kubectl apply -f test/e2e/pod-runner/04-resourcegraph.yaml

# Check the results
kubectl get pods -n arc-runners
```

## Common Issues

### Image Not Found

If you see "image not found" errors:

```bash
# Rebuild and load image
mise run build
docker build -t kro-actions-runner:latest .
mise run cluster:delete
mise run cluster:setup
```

### KRO/ARC Installation Fails

Check Helm installations:

```bash
helm list -A
kubectl get pods -n kro-system
kubectl get pods -n arc-systems
```

Reinstall if needed:

```bash
mise run cluster:reset
```

### Permission Denied

Ensure RBAC is applied:

```bash
kubectl get serviceaccount kro-runner -n arc-runners
kubectl get clusterrole kro-runner-cluster
```

### Tests Timeout

Increase timeout in `test/e2e/kuttl-test.yaml`:

```yaml
timeout: 900  # 15 minutes
```

## CI Integration

To run tests in CI, add to your GitHub Actions workflow:

```yaml
- name: Run integration tests
  run: |
    mise install
    mise run test:e2e
```

## Adding New Tests

1. Create a new directory under `test/e2e/`
2. Add test steps as numbered YAML files (00-*, 01-*, etc.)
3. Add corresponding assert files (*-assert.yaml)
4. Update `test/e2e/kuttl-test.yaml` to include the new test:

```yaml
testDirs:
  - ./pod-runner/
  - ./rgd-discovery/
  - ./rbac-validation/
  - ./your-new-test/
```

## References

- [Kuttl Documentation](https://kuttl.dev/)
- [Kind Documentation](https://kind.sigs.k8s.io/)
- [KRO Documentation](https://github.com/awslabs/kro)
- [Actions Runner Controller](https://github.com/actions/actions-runner-controller)
