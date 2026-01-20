# KRO Actions Runner - Implementation Summary

## ✅ Completed: Secure KRO-Based GitHub Actions Runner

**kro-actions-runner** - a secure, flexible GitHub Actions runner that uses KRO (Kubernetes Resource Orchestrator) to provision compute resources dynamically. Inspired by kubevirt-actions-runner.

---

## 🎯 What Was Achieved

### 1. **Core KRO Integration** ✅
- **File**: [internal/kro_runner.go](internal/kro_runner.go)
- **Features**:
  - Label-based RGD (ResourceGraphDefinition) discovery
  - Dynamic ResourceGraph instance creation and lifecycle management
  - Secure secret handling (JIT configs passed by reference, not inline)
  - Status synchronization and cleanup

### 2. **Security Model** ✅
- **JIT Config Protection**:
  - ✅ Secrets stored in Kubernetes Secrets only
  - ✅ Only secret **names** pass through RG specs
  - ✅ RGD templates use `secretRef` for injection
  - ✅ No tokens in status, logs, or events
  - ✅ Kubernetes Secret encryption at rest (when enabled)

### 3. **Removed Dependencies** ✅
- **Removed**: KubeVirt client libraries (incompatible with k8s.io v0.34)
- **Reason**: KRO manages VMs through RGDs, so direct KubeVirt integration not needed
- **Benefit**: Cleaner dependencies, smaller image size, faster builds

### 4. **Documentation & Examples** ✅
Created comprehensive examples and documentation:
- [README-KRO.md](README-KRO.md) - Complete usage guide
- [examples/pod-runner-rgd.yaml](examples/pod-runner-rgd.yaml) - Pod-based runners
- [examples/vm-runner-rgd.yaml](examples/vm-runner-rgd.yaml) - KubeVirt VM runners
- [examples/arc-scale-set-values.yaml](examples/arc-scale-set-values.yaml) - ARC Helm values
- [examples/rbac.yaml](examples/rbac.yaml) - Required RBAC permissions

### 5. **Docker Image** ✅
Successfully built and verified:
```text
IMAGE                       DISK USAGE   CONTENT SIZE
kro-actions-runner:latest   54.3MB       16.4MB
```

---

## 🏗️ Architecture

### How It Works

```text
GitHub Job → ARC (upstream) → Pod with kro-actions-runner
                                    ↓
                  Discovers RGD by scale-set-name label
                                    ↓
         Creates ResourceGraph instance (PodRunner/VMRunner)
                                    ↓
                      KRO provisions resources
                                    ↓
              Pod/VM runs GitHub Actions job
                                    ↓
            Resources cleaned up automatically
```

### Label-Based Discovery

The runner discovers which RGD to use via label matching:

```yaml
# Runner Pod environment
env:
  - name: ACTIONS_RUNNER_SCALE_SET_NAME
    value: "default"

---
# RGD with matching label
apiVersion: kro.run/v1alpha1
kind: ResourceGraphDefinition
metadata:
  labels:
    actions.github.com/scale-set-name: "default"  # ← Matches!
spec:
  schema:
    kind: PodRunner  # ← This Kind will be created
```

---

## 🔒 Security Implementation

### Secret Flow

1. **ARC creates JIT secret** containing GitHub runner token
2. **kro-actions-runner reads** from environment: `ACTIONS_RUNNER_INPUT_JITCONFIG`
3. **Creates Kubernetes Secret** with JIT config data
4. **ResourceGraph spec** contains only the **secret name** (not token)
5. **RGD template** uses `secretRef` to reference secret
6. **Kubernetes injects** secret into runner Pod/VM

### What's NOT Exposed

- ❌ JIT tokens in ResourceGraph specs
- ❌ Tokens in ResourceGraph status
- ❌ Tokens in logs or events
- ❌ Tokens in etcd (beyond encrypted Secrets)

---

## 📦 Key Files

### Implementation
- **[internal/kro_runner.go](internal/kro_runner.go)** - Main KRO integration logic
  - `NewKRORunner()` - Constructor
  - `findRGDByLabel()` - RGD discovery (lines 88-133)
  - `CreateResources()` - RG instance creation (lines 135-203)
  - `WaitForResourceGraph()` - Status monitoring (lines 205-263)
  - `DeleteResources()` - Cleanup (lines 265-312)
  - `createJitSecret()` - Secure secret creation (lines 314-336)

### Entry Points
- **[cmd/kar/main.go](cmd/kar/main.go)** - Main entry point
- **[cmd/kar/app/root.go](cmd/kar/app/root.go)** - CLI command structure
- **[cmd/kar/app/opts.go](cmd/kar/app/opts.go)** - Configuration options
- **[cmd/kar/app/flag.go](cmd/kar/app/flag.go)** - CLI flags

### Docker
- **[Dockerfile](Dockerfile)** - Multi-stage build (Go → scratch)

---

## 🚀 Quick Start

### 1. Install Prerequisites

```bash
# Install KRO
kubectl apply -f https://github.com/kubernetes-sigs/kro/releases/latest/download/kro.yaml

# Install ARC (if not already installed)
kubectl apply -f https://github.com/actions/actions-runner-controller/releases/latest/download/actions-runner-controller.yaml
```

### 2. Deploy RBAC

```bash
kubectl apply -f examples/rbac.yaml -n arc-runners
```

### 3. Deploy RGD

```bash
# For Pod-based runners
kubectl apply -f examples/pod-runner-rgd.yaml

# For VM-based runners (requires KubeVirt)
kubectl apply -f examples/vm-runner-rgd.yaml
```

### 4. Deploy Runner Scale Set

```bash
helm upgrade --install --namespace arc-runners --create-namespace \
  --values examples/arc-scale-set-values.yaml \
  my-runners \
  oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set
```

---

## 🔄 Key Advantages

- ❌ No controller modifications required
- ✅ Excellent secret security via Kubernetes Secrets
- ✅ Multi-backend support via RGD (Pods, VMs, cloud instances)
- ✅ Full upstream ARC compatibility

---

## 🛠️ Build Instructions

### Build Binary

```bash
go build -o bin/kar ./cmd/kar
```

### Build Docker Image

```bash
docker build -t kro-actions-runner:latest .
```

### Load into Kind (for testing)

```bash
kind load docker-image kro-actions-runner:latest --name <cluster-name>
```

---

## 🧪 Testing

### Test RGD Discovery

```bash
# Create test RGD
kubectl apply -f examples/pod-runner-rgd.yaml

# Create test secret
kubectl create secret generic test-jit \
  --from-literal=.jitconfig='{"test":"config"}'

# Run test
kubectl run test-runner \
  --image=kro-actions-runner:latest \
  --env="ACTIONS_RUNNER_SCALE_SET_NAME=default" \
  --env="RUNNER_NAME=test-runner-1" \
  --env="ACTIONS_RUNNER_INPUT_JITCONFIG=test-jit-value" \
  --restart=Never

# Check logs
kubectl logs test-runner -f
```

---

## 🔍 Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ACTIONS_RUNNER_INPUT_JITCONFIG` | ✅ Yes | - | JIT config from ARC |
| `RUNNER_NAME` | ✅ Yes | - | Runner name (use Pod name) |
| `ACTIONS_RUNNER_SCALE_SET_NAME` | ✅ Yes | - | Scale set name for RGD discovery |
| `KAR_CLEANUP_TIMEOUT` | ❌ No | `5m` | Resource cleanup timeout |

---

## 📋 Next Steps

### Immediate

1. **Test with real GitHub repository**: Configure GitHub App/PAT and test end-to-end
2. **Create additional RGDs**:
   - EC2Runner for AWS instances
   - GCERunner for Google Cloud VMs
   - FirecrackerRunner for microVMs

### Future Enhancements

1. **Status Reporting**: Better detection of runner completion (currently monitors RG state)
2. **Error Handling**: More granular error messages and retry logic
3. **Metrics**: Prometheus metrics for runner lifecycle events
4. **Multi-Cluster**: Support for runners across multiple Kubernetes clusters

---

## 🎉 Summary

Successfully created a **secure, flexible GitHub Actions runner** that:

✅ Uses **upstream ARC unchanged** (no controller modifications)
✅ **Securely handles JIT configs** (reference-based, no inline secrets)
✅ Supports **multiple compute types** via RGD (Pods, VMs, Cloud instances)
✅ **Clean dependencies** (removed KubeVirt client conflicts)
✅ **Small image size** (54MB) with minimal attack surface
✅ **Comprehensive documentation** and examples

**Image**: `kro-actions-runner:latest` (16.4MB compressed, 54.3MB on disk)

---

## 📚 Resources

- [KRO Documentation](https://github.com/kubernetes-sigs/kro)
- [ARC Documentation](https://github.com/actions/actions-runner-controller)
- [GitHub Actions Runner](https://github.com/actions/runner)
- [KubeVirt](https://kubevirt.io) (for VM runners)

---

**Built with ❤️ for secure, flexible GitHub Actions in Kubernetes**
