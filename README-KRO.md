# KRO Actions Runner

A custom GitHub Actions runner image that uses **KRO (Kubernetes Resource Orchestrator)** to provision compute resources dynamically. This provides the security of the kubevirt-actions-runner approach while supporting multiple compute backends (Pods, VMs, cloud instances) through RGD (ResourceGraphDefinition) templates.

## Overview

Instead of modifying ARC controllers or CRDs, this runner image:
1. Runs inside a standard ARC Pod (unchanged upstream ARC)
2. Reads JIT config from standard ARC environment variables
3. Discovers the appropriate RGD by label matching
4. Creates a KRO ResourceGraph instance for the actual compute
5. Waits for the runner to complete
6. Cleans up resources

## Architecture

```
GitHub Job → ARC (upstream) → Pod with kro-actions-runner image
                                    ↓
                   Discovers RGD by scale-set-name label
                                    ↓
              Creates ResourceGraph instance (PodRunner/VMRunner/etc)
                                    ↓
                        KRO provisions the resources
                                    ↓
                Pod/VM runs GitHub Actions runner with JIT config
                                    ↓
                        Job completes, resources cleaned up
```

## Security Model

### Secrets Never Travel Through RG Specs

The JIT config secret is handled securely:

1. **ARC creates the JIT secret** with the GitHub runner token
2. **kro-actions-runner reads it from environment**: `ACTIONS_RUNNER_INPUT_JITCONFIG`
3. **Creates a Kubernetes Secret** with the JIT config
4. **ResourceGraph spec contains only the SECRET NAME** (not the token)
5. **RGD template uses `secretRef`** to reference the secret
6. **Kubernetes injects the secret** into the actual runner Pod/VM

**No tokens are exposed in:**
- ResourceGraph specs
- ResourceGraph status
- Logs or events
- etcd (beyond standard Secret encryption)

## Prerequisites

- Kubernetes cluster (1.25+)
- [KRO installed](https://github.com/kubernetes-sigs/kro)
- [Actions Runner Controller (ARC)](https://github.com/actions/actions-runner-controller) installed
- Optional: KubeVirt (for VM runners)

## Quick Start

### 1. Install KRO

```bash
kubectl apply -f https://github.com/kubernetes-sigs/kro/releases/latest/download/kro.yaml
```

### 2. Deploy a ResourceGraphDefinition

Choose a compute backend:

**Option A: Pod-based runners** (simplest, no additional dependencies)

```bash
kubectl apply -f examples/pod-runner-rgd.yaml
```

**Option B: VM-based runners** (requires KubeVirt)

```bash
# Install KubeVirt first
kubectl apply -f https://github.com/kubevirt/kubevirt/releases/latest/download/kubevirt-operator.yaml
kubectl apply -f https://github.com/kubevirt/kubevirt/releases/latest/download/kubevirt-cr.yaml

# Deploy the VM runner RGD
kubectl apply -f examples/vm-runner-rgd.yaml
```

### 3. Deploy ARC Runner Scale Set

Create a `values.yaml`:

```yaml
githubConfigUrl: https://github.com/<your_org>
githubConfigSecret: <your_github_secret>

# Use the kro-actions-runner image
template:
  spec:
    containers:
      - name: runner
        image: kro-actions-runner:latest
        command: []
        env:
          # Scale set name for RGD discovery
          - name: ACTIONS_RUNNER_SCALE_SET_NAME
            value: "default"  # Must match RGD label
          # Runner name from Pod name
          - name: RUNNER_NAME
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          # JIT config (provided by ARC automatically)
          - name: ACTIONS_RUNNER_INPUT_JITCONFIG
            valueFrom:
              secretKeyRef:
                name: $(RUNNER_NAME)
                key: .jitconfig
```

Install the scale set:

```bash
helm upgrade --install --namespace arc-runners --create-namespace \
  --values values.yaml \
  my-runners \
  oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set
```

## How It Works

### Label-Based RGD Discovery

The kro-actions-runner discovers which RGD to use via label matching:

1. **Scale set name** is passed via `ACTIONS_RUNNER_SCALE_SET_NAME` environment variable
2. **Runner searches** for RGDs with label: `actions.github.com/scale-set-name: <scale-set-name>`
3. **RGD Kind** determines the resource type (PodRunner, VMRunner, EC2Runner, etc.)
4. **ResourceGraph instance** is created with that Kind

### Example Flow

```yaml
# Runner scale set values.yaml
env:
  - name: ACTIONS_RUNNER_SCALE_SET_NAME
    value: "vm-runners"

---
# RGD with matching label
apiVersion: kro.run/v1alpha1
kind: ResourceGraphDefinition
metadata:
  name: vm-runner
  labels:
    actions.github.com/scale-set-name: "vm-runners"  # <-- Matches!
spec:
  schema:
    kind: VMRunner  # <-- This Kind will be created
```

When a job arrives:
1. ARC creates Pod with kro-actions-runner image
2. Runner discovers the `vm-runner` RGD (label match)
3. Creates a `VMRunner` ResourceGraph instance
4. KRO provisions the VM
5. VM runs the GitHub Actions job
6. Resources are cleaned up

## Creating Custom RGDs

### RGD Requirements

Your RGD must:

1. **Have the label** for discovery:
   ```yaml
   metadata:
     labels:
       actions.github.com/scale-set-name: "<your-scale-set-name>"
   ```

2. **Define these spec fields**:
   ```yaml
   spec:
     schema:
       spec:
         jitConfigSecretName: string  # Secret reference
         runnerName: string            # Runner name
   ```

3. **Reference the JIT secret** (not inline):
   ```yaml
   env:
     - name: ACTIONS_RUNNER_INPUT_JITCONFIG
       valueFrom:
         secretKeyRef:
           name: ${schema.spec.jitConfigSecretName}
           key: .jitconfig
   ```

### Example: EC2 Runner RGD

```yaml
apiVersion: kro.run/v1alpha1
kind: ResourceGraphDefinition
metadata:
  name: ec2-runner
  labels:
    actions.github.com/scale-set-name: "aws-runners"
spec:
  schema:
    apiVersion: v1alpha1
    kind: EC2Runner
    spec:
      jitConfigSecretName: string
      runnerName: string

  resources:
    - id: ec2-instance
      template:
        apiVersion: ec2.services.k8s.aws/v1alpha1
        kind: Instance
        metadata:
          name: ${schema.spec.runnerName}
        spec:
          imageID: ami-12345678
          instanceType: t3.medium
          userData: |
            #!/bin/bash
            # Fetch JIT config from AWS Secrets Manager or Parameter Store
            # Install and configure GitHub Actions runner
            # Run: ./run.sh --jitconfig "$JITCONFIG"
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ACTIONS_RUNNER_INPUT_JITCONFIG` | Yes | - | JIT config from ARC (automatic) |
| `RUNNER_NAME` | Yes | - | Runner name (use Pod name) |
| `ACTIONS_RUNNER_SCALE_SET_NAME` | Yes | - | Scale set name for RGD discovery |
| `USE_KRO` | No | `true` | Use KRO mode (vs legacy KubeVirt) |
| `KAR_CLEANUP_TIMEOUT` | No | `5m` | Cleanup timeout duration |

## Command-Line Flags

```bash
# KRO mode (default)
kar --use-kro \
    --scale-set-name "default" \
    --runner-name "runner-abc123" \
    --actions-runner-input-jitconfig "<jit-config>"

# Legacy KubeVirt mode
kar --use-kro=false \
    --kubevirt-vm-template "ubuntu-jammy-vm" \
    --runner-name "runner-abc123" \
    --actions-runner-input-jitconfig "<jit-config>"
```

## Building

```bash
# Build the Docker image
docker build -t kro-actions-runner:latest .

# Or use the provided Makefile
make docker-build
```

## Testing

### Test RGD Discovery

```bash
# Create a test RGD
kubectl apply -f examples/pod-runner-rgd.yaml

# Create a test secret
kubectl create secret generic test-jit --from-literal=.jitconfig='{"test":"config"}'

# Run the runner
kubectl run test-runner \
  --image=kro-actions-runner:latest \
  --env="ACTIONS_RUNNER_SCALE_SET_NAME=default" \
  --env="RUNNER_NAME=test-runner-1" \
  --env="ACTIONS_RUNNER_INPUT_JITCONFIG=$(kubectl get secret test-jit -o jsonpath='{.data.\.jitconfig}' | base64 -d)" \
  --restart=Never

# Check logs
kubectl logs test-runner -f
```

## Troubleshooting

### Runner can't find RGD

```
Error: no RGD found with label actions.github.com/scale-set-name=default
```

**Solution**: Ensure the RGD has the correct label:

```bash
kubectl get rgd -o custom-columns=NAME:.metadata.name,LABELS:.metadata.labels
```

### ResourceGraph stays in PENDING

```
ResourceGraph test-runner-1 state: PENDING
```

**Solution**: Check KRO controller logs and RG status:

```bash
kubectl logs -n kro-system deployment/kro-controller
kubectl describe podrunner test-runner-1
```

### Secret not found

```
Error: secret "runner-xyz-jit" not found
```

**Solution**: Ensure the runner has permission to create secrets:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
rules:
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["create", "get", "delete"]
```

## Security Best Practices

1. **Enable etcd encryption at rest** for Secrets
2. **Use RBAC** to limit Secret access per runner
3. **Enable audit logging** to track Secret access
4. **Validate RGD specs** don't contain inline credentials
5. **Use separate namespaces** for different runner types
6. **Consider external-secrets operator** for secret rotation

## Comparison with Other Approaches

| Approach | Pros | Cons |
|----------|------|------|
| **kro-actions-runner (this project)** | ✅ No controller modifications<br>✅ Supports multiple compute types<br>✅ Secure secret handling<br>✅ Uses upstream ARC | ⚠️ Requires KRO installation |
| **Controller modification** | ✅ Direct integration | ❌ Must maintain fork<br>❌ Complex upgrades<br>❌ Conflicts with upstream |
| **kubevirt-actions-runner** | ✅ No controller modifications<br>✅ Proven approach | ❌ KubeVirt-only<br>❌ No abstraction for other backends |

## Contributing

Contributions welcome! Please open issues or PRs.

## License

Apache 2.0 - See LICENSE file
