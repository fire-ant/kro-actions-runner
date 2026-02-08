# Development Guide

This guide covers local development using Tilt for a streamlined Kubernetes development experience.

## Prerequisites

Install tools using mise:
```bash
mise install
```

This installs: Go, kubectl, kind, helm, kuttl, ctlptl, tilt, and other dev tools.

## Quick Start (Recommended)

### 1. One-Time Setup

For initial cluster setup without staying attached:

```bash
mise up
```

This will:
1. Create kind cluster with local registry
2. Build Docker image
3. Install KRO, ARC, LocalStack, and ACK EC2 Controller
4. Exit when complete

**Perfect for:** Setting up the cluster infrastructure once.

### 2. Development with Live Reload

For active development with automatic rebuilds:

```bash
mise dev
```

This will:
1. Create kind cluster (if needed)
2. Start Tilt in watch mode
3. Auto-rebuild on code changes
4. Stream logs to terminal

**Perfect for:** Iterating on code with fast feedback loop.

Press Ctrl+C to stop when done.

## Alternative: Direct Tilt Usage

If you prefer using Tilt directly:

**Option A: Stream logs to terminal**
```bash
tilt up --stream
```

All logs appear in your terminal. Press Ctrl+C to stop.

**Option B: Use web UI**
```bash
tilt up
```

Then open [http://localhost:10350](http://localhost:10350)

**Option C: Build once and exit (CI mode)**
```bash
tilt ci
```

The Tilt UI shows:
- **Build status** for kar binary and image
- **Dependency installation** (KRO, ARC, LocalStack, ACK)
- **Test resources** status
- **Logs** for each component (clickable)
- **Manual triggers** for tests and tools

## Verify Setup

Check that all resources are running:

```bash
mise run cluster:verify
# or
kubectl get pods -A
```

## Running Tests

### Install Test Resources

Before running tests, install the RGDs and test setup:

```bash
mise run test:ec2:install-rgds
```

This creates:
- `arc-runners` namespace
- RBAC for kar testing
- VPC Network RGD
- EC2 Runner RGD
- Test VPC setup

### Run Tests

In the Tilt UI:
- Click **test-unit** → Press trigger button (or `t`)
- Click **test-ec2-integration** → Press trigger button

Or from terminal:
```bash
# Unit tests
go test -v ./...

# EC2 integration tests
kubectl kuttl test --config test/e2e/ec2-integration/kuttl/kuttl-test.yaml
```

### 6. View LocalStack EC2 instances

In Tilt UI, trigger **view-ec2-instances** to see simulated EC2 instances.

Or manually:
```bash
kubectl exec -n localstack deploy/localstack -- awslocal ec2 describe-instances
```

### 7. Manual test runners

Create test runner pods with kar:
```bash
kubectl apply -f test/e2e/ec2-integration/kuttl/tests/01-full-lifecycle/01-simulate-jobs.yaml
```

Watch kar create EC2Runners:
```bash
kubectl get ec2runner -n arc-runners -w
```

Check logs:
```bash
kubectl logs -n arc-runners runner-0 -f
```

Delete pod to test garbage collection:
```bash
kubectl delete pod runner-0 -n arc-runners
# EC2Runner should be auto-deleted via owner reference
kubectl get ec2runner -n arc-runners
```

### 8. Stop Tilt

Press `Ctrl+C` in the terminal running `tilt up`, or:
```bash
tilt down
```

This **does not** delete the cluster. Resources remain for faster restarts.

### 9. Clean up cluster

When done:
```bash
ctlptl delete -f test/kind/ctlptl-kind-config.yaml
```

## Configuration

### Tilt Options

```bash
# Disable LocalStack (faster startup if not testing EC2)
tilt up -- --enable-localstack=false

# Custom registry
tilt up -- --registry-host=myregistry:5000
```

### Environment Variables

See `mise.toml` for environment variables:
- `DOCKER_CMD`: Docker command (default: `docker`, can use `podman`)
- `LOCALSTACK_ENDPOINT`: LocalStack endpoint
- `AWS_REGION`: AWS region for ACK

## Alternative: Manual Setup (mise tasks)

If you prefer not to use Tilt:

```bash
# Build
mise run build

# Build image
mise run image:build

# Setup cluster
ctlptl apply -f test/kind/ctlptl-kind-config.yaml

# Install dependencies manually
mise run test:e2e  # Runs full e2e test suite with setup
```

See `mise.toml` for all available tasks.

## Troubleshooting

### Tilt fails to start

Check cluster is running:
```bash
kubectl cluster-info
```

If not, recreate:
```bash
ctlptl delete -f test/kind/ctlptl-kind-config.yaml
ctlptl apply -f test/kind/ctlptl-kind-config.yaml
```

### Helm chart installation fails

Check namespace exists:
```bash
kubectl get ns kro-system arc-systems localstack
```

Manually install if needed:
```bash
helm upgrade --install kro oci://registry.k8s.io/kro/charts/kro --version 0.8.4 --namespace kro-system --create-namespace
```

### Image not loading

Verify image in kind:
```bash
docker exec gha-arc-kro-runner-control-plane crictl images | grep kro-actions-runner
```

Manually load:
```bash
kind load docker-image localhost:5005/kro-actions-runner:latest --name gha-arc-kro-runner
```

### LocalStack not working

Check LocalStack pods:
```bash
kubectl get pods -n localstack
kubectl logs -n localstack deploy/localstack
```

Test endpoint:
```bash
kubectl exec -n localstack deploy/localstack -- awslocal ec2 describe-instances
```

## IDE Integration

### VSCode + Tilt

Install [Tilt extension](https://marketplace.visualstudio.com/items?itemName=tilt-dev.Tiltfile):
- Syntax highlighting for Tiltfile
- Quick navigation to resources

### IntelliJ + Tilt

Install [Tilt plugin](https://plugins.jetbrains.com/plugin/15869-tilt):
- View Tilt resources in IDE
- One-click logs

## Next Steps

- Read [README.md](README.md) for project overview
- Check [Warm Pool Plan](/Users/clavery/.claude/plans/stateful-beaming-adleman.md) for upcoming features
- Run tests: `mise run test`
- Contribute: See pull request guidelines in README
