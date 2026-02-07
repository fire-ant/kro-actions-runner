# KRO Actions Runner

A GitHub Actions runner that uses **KRO (Kubernetes Resource Orchestrator)** to dynamically provision compute resources. Supports Pods, VMs, and cloud instances through ResourceGraphDefinition templates.

## Prerequisites

- Kubernetes cluster (1.25+)
- [KRO](https://github.com/kubernetes-sigs/kro) installed
- [Actions Runner Controller (ARC)](https://github.com/actions/actions-runner-controller) installed

## Installation

### 1. Install KRO

```bash
kubectl apply -f https://github.com/kubernetes-sigs/kro/releases/latest/download/kro.yaml
```

### 2. Deploy a ResourceGraphDefinition

```bash
kubectl apply -f examples/pod-runner-rgd.yaml
```

### 3. Deploy ARC Runner Scale Set

Create `values.yaml`:

```yaml
githubConfigUrl: https://github.com/<your_org>
githubConfigSecret: <your_github_secret>

template:
  spec:
    containers:
      - name: runner
        image: kro-actions-runner:latest
        command: []
        env:
          - name: ACTIONS_RUNNER_SCALE_SET_NAME
            value: "default"  # Must match RGD label
          - name: RUNNER_NAME
            valueFrom:
              fieldRef:
                fieldPath: metadata.name
          - name: ACTIONS_RUNNER_INPUT_JITCONFIG
            valueFrom:
              secretKeyRef:
                name: $(RUNNER_NAME)
                key: .jitconfig
```

Install:

```bash
helm upgrade --install --namespace arc-runners --create-namespace \
  --values values.yaml \
  my-runners \
  oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set
```

## How It Works

1. GitHub job triggers → ARC creates Pod with kro-actions-runner image
2. Runner discovers RGD by matching `ACTIONS_RUNNER_SCALE_SET_NAME` to RGD label
3. Runner creates ResourceGraph instance
4. KRO provisions the actual compute (Pod/VM/instance)
5. Compute runs GitHub Actions job
6. Resources are cleaned up

## Creating Custom RGDs

Your RGD must have:

1. Label for discovery:
```yaml
metadata:
  labels:
    actions.github.com/scale-set-name: "your-scale-set-name"
```

2. Required spec fields:
```yaml
spec:
  schema:
    spec:
      jitConfigSecretName: string
      runnerName: string
```

3. Secret reference (not inline):
```yaml
env:
  - name: ACTIONS_RUNNER_INPUT_JITCONFIG
    valueFrom:
      secretKeyRef:
        name: ${schema.spec.jitConfigSecretName}
        key: .jitconfig
```

See `examples/` for complete examples.

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `ACTIONS_RUNNER_INPUT_JITCONFIG` | Yes | JIT config from ARC |
| `RUNNER_NAME` | Yes | Runner name (use Pod name) |
| `ACTIONS_RUNNER_SCALE_SET_NAME` | Yes | Scale set name for RGD discovery |
| `KAR_CLEANUP_TIMEOUT` | No | Cleanup timeout (default: 5m) |

## EC2 Runners with LocalStack

For testing EC2-based runners locally without AWS costs, we support LocalStack + ACK EC2 integration.

### Architecture

Two-tier ResourceGraph approach:

1. **Infrastructure Layer** (deploy once): VPC, Subnet, Security Group
2. **Application Layer** (ephemeral): EC2 runner instances

```
GitHub Workflow → ARC creates JIT Secret
    ↓
EC2RunnerWithVPC ResourceGraph
    ↓ (references secret via ${string(jitSecret.data[".jitconfig"])})
    ↓
KRO creates ACK EC2 Instance
    ↓ (JIT config injected into userData)
    ↓
LocalStack EC2 Service (or real AWS)
    ↓
GitHub Actions runner executes workflow
```

### Quick Start with LocalStack

```bash
# Complete development setup (creates cluster, installs everything)
mise run dev:setup

# Deploy VPC infrastructure (once)
kubectl apply -f examples/ec2-runner/test-vpc-network.yaml

# Create test runner with JIT config
kubectl apply -f examples/ec2-runner/test-ec2-runner-secret.yaml
kubectl apply -f examples/ec2-runner/test-ec2-runner-instance.yaml

# Watch instance creation
mise run ec2:watch-instances
```

### Key Features

- **JIT Config Injection**: Kubernetes secrets automatically injected into EC2 userData via CEL expression
- **VPC Reuse**: One VPC Network serves many ephemeral runners
- **LocalStack Testing**: Test EC2 workflows without AWS costs
- **Production Ready**: Same RGDs work with real AWS (switch to Secrets Manager for production)

See [examples/README.md](examples/README.md) for more examples and detailed guides.

## Development

This project uses [mise](https://mise.jdx.dev) for tool management.

```bash
# Install tools
mise install

# Run tests
mise run test

# Build
mise run build

# Development cluster setup
mise run dev:setup
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## Testing

```bash
# Integration tests
mise run test:e2e

# EC2 LocalStack tests
mise run ec2:test
```

See [test/README.md](test/README.md) for comprehensive testing documentation.

## License

Apache 2.0
