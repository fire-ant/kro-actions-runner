# LocalStack Integration for EC2 Runner Testing

This directory contains configuration files and scripts for integrating LocalStack into the test infrastructure to support EC2-based GitHub Actions runners.

## Overview

LocalStack provides a local AWS cloud environment that allows testing ACK EC2 controllers without needing real AWS resources. This integration enables:

1. **HTTPS-enabled LocalStack** with self-signed certificates
2. **ACK EC2 Controller** configured to use LocalStack endpoints
3. **EC2 Runner ResourceGraphDefinition** for creating GitHub Actions runners as EC2 instances
4. **End-to-end test cases** validating the complete EC2 runner workflow

## Architecture

```
GitHub Actions Workflow
    ↓
KRO Actions Runner (kar binary)
    ↓ (discovers EC2Runner RGD)
    ↓ (creates JIT secret)
    ↓
EC2Runner ResourceGraph (KRO)
    ↓ (creates ACK EC2 Instance resource)
    ↓
ACK EC2 Controller
    ↓ (talks to LocalStack HTTPS endpoint)
    ↓
LocalStack EC2 Service
    ↓ (creates Docker container as fake EC2 instance)
    ↓
GitHub Actions Runner
    ↓ (runs workflow)
    ↓ (terminates when complete)
```

## Files

### Configuration Files

- **localstack-values.yaml** - Helm values for LocalStack deployment
  - Enables HTTPS on ports 4566 and 443
  - Mounts Docker socket for EC2 instance simulation
  - Configures resource limits for test environment

- **ack-ec2-values.yaml** - Helm values for ACK EC2 Controller
  - Configures HTTPS endpoint to LocalStack
  - Skips SSL verification for self-signed certificates
  - Sets up dummy AWS credentials (not used by LocalStack)

- **localstack-init-tls.sh** - TLS certificate generation script
  - Generates self-signed certificates for HTTPS
  - Includes Subject Alternative Names for cluster DNS
  - Creates certificates valid for 365 days

## Installation Scripts

Located in `test/scripts/`:

### install-localstack.sh

Installs and configures LocalStack with HTTPS support:

1. Adds LocalStack Helm repository
2. Creates `localstack` namespace
3. Generates self-signed TLS certificates
4. Creates Kubernetes secret with certificates
5. Installs LocalStack Helm chart
6. Waits for LocalStack to be ready

**Usage:**
```bash
./test/scripts/install-localstack.sh
```

**Endpoint:** `https://localstack.localstack.svc.cluster.local:4566`

### install-ack-ec2.sh

Installs ACK EC2 Controller configured for LocalStack:

1. Creates `ack-system` namespace
2. Creates dummy AWS credentials secret
3. Installs ACK EC2 controller from OCI registry
4. Configures controller to use LocalStack HTTPS endpoint
5. Waits for controller to be ready

**Usage:**
```bash
./test/scripts/install-ack-ec2.sh
```

**Configuration:**
- Endpoint: `https://localstack.localstack.svc.cluster.local:4566`
- SSL Verification: Disabled (uses `AWS_EC2_SKIP_SSL_VERIFICATION=true`)
- Region: `us-east-1`

## HTTPS Configuration

LocalStack is configured with HTTPS endpoints to match production AWS behavior:

### Certificate Generation

Self-signed certificates are generated with the following SANs:
- `localhost`
- `localstack`
- `localstack.localstack`
- `localstack.localstack.svc`
- `localstack.localstack.svc.cluster.local`
- `127.0.0.1`

### SSL Verification

ACK EC2 Controller is configured to skip SSL verification because LocalStack uses self-signed certificates. This is acceptable for testing but would not be used in production.

## EC2 Runner ResourceGraphDefinition

The EC2 Runner RGD (`test/fixtures/rgds/ec2-runner-rgd.yaml`) defines how to create GitHub Actions runners as EC2 instances:

### Schema

```yaml
apiVersion: v1alpha1
kind: EC2Runner
spec:
  runnerName: string        # Name of the runner (from JIT secret)
  imageID: string           # AMI ID (default: ami-000001)
  instanceType: string      # Instance type (default: t3.medium)
  subnetID: string          # Required subnet ID
```

### Resources

1. **jitSecret** - External reference to ARC-created JIT configuration secret
2. **runnerInstance** - ACK EC2 Instance resource with:
   - User data script to run GitHub Actions runner
   - Automatic shutdown after workflow completion
   - Tags for identification

### User Data Script

The instance user data:
1. Extracts JIT config from the secret
2. Sets environment variables
3. Runs the GitHub Actions runner (`./run.sh`)
4. Shuts down the instance when complete

## Testing

### Running EC2 Runner Tests

The EC2 runner test case is located in `test/e2e/ec2-runner/basic/` and includes:

1. **00-namespace.yaml** - Create test namespace
2. **01-rbac.yaml** - Apply RBAC for KRO runner
3. **02-localstack-check.yaml** - Verify LocalStack and ACK are running
4. **03-rgd.yaml** - Apply EC2Runner RGD
5. **04-secret.yaml** - Create mock JIT config secret
6. **05-vpc-subnet.yaml** - Create VPC and subnet for instances
7. **06-resourcegraph.yaml** - Create EC2Runner instance
8. **07-cleanup.yaml** - Delete resources

### Using Mise Tasks

```bash
# Run only EC2 runner tests (includes LocalStack setup)
mise run test:e2e:ec2

# Set up cluster with LocalStack and ACK
mise run cluster:setup:localstack

# Run all tests except EC2 (skips LocalStack)
mise run test:e2e:skip-ec2

# Run all tests including EC2
mise run test:e2e
```

### Using Kuttl Directly

```bash
# Run all tests (includes EC2 by default)
kubectl kuttl test --config test/e2e/kuttl-test.yaml

# Run only EC2 runner test
kubectl kuttl test --config test/e2e/kuttl-test.yaml --test ec2-runner

# Skip EC2 tests
SKIP_EC2_TESTS=true kubectl kuttl test --config test/e2e/kuttl-test.yaml
```

## Environment Variables

- **SKIP_EC2_TESTS** - Set to `true` to skip LocalStack/ACK installation and EC2 tests
- **ACK_EC2_VERSION** - ACK EC2 controller version (default: `1.9.2`)

## Troubleshooting

### LocalStack Pod Not Starting

Check if Docker socket is mounted:
```bash
kubectl describe pod -n localstack -l app.kubernetes.io/name=localstack
```

Ensure the kind cluster was created with the updated config:
```bash
kind create cluster --name kro-test --config test/fixtures/kind/kind-config.yaml
```

### ACK Controller Can't Reach LocalStack

Check endpoint configuration:
```bash
kubectl get deployment ack-ec2-controller -n ack-system -o yaml | grep endpoint
```

Should show: `https://localstack.localstack.svc.cluster.local:4566`

Check LocalStack service:
```bash
kubectl get svc -n localstack
```

### SSL Certificate Issues

Check if TLS secret exists:
```bash
kubectl get secret localstack-tls-certs -n localstack
```

Check ACK controller logs:
```bash
kubectl logs -n ack-system -l app.kubernetes.io/name=ec2-chart --tail=50
```

### EC2 Instances Not Launching

Verify LocalStack is running:
```bash
kubectl exec -n localstack deploy/localstack -- awslocal ec2 describe-instances
```

Check VPC and subnet:
```bash
kubectl get vpc,subnet -n arc-runners
```

## Differences from Examples

The files in this directory are adapted from `examples/` for use in the automated test suite:

1. **Resource limits** - Adjusted for CI/CD environments
2. **Test-specific configuration** - Optimized for quick iteration
3. **Integration with Kuttl** - Configured for automated testing
4. **Conditional execution** - Can be skipped via environment variables

## References

- [LocalStack Documentation](https://docs.localstack.cloud/)
- [ACK EC2 Controller](https://aws-controllers-k8s.github.io/community/reference/ec2/v1alpha1/instance/)
- [Kuttl Testing](https://kuttl.dev/)
- [KRO (Kubernetes Resource Orchestrator)](https://kro.run/)
