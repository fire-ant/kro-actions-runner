# LocalStack + ACK EC2 Setup for Testing

This guide explains how to set up and test GitHub Actions runners using LocalStack EC2 instances with ACK controllers.

## Prerequisites

- Docker running on your host machine
- kind CLI
- kubectl
- helm

## Setup

### 1. Create Cluster with LocalStack Support

```bash
cd examples
KIND_CONFIG=kind-localstack.yaml ./setup-cluster.sh
```

This will:
- Create a kind cluster with Docker socket mounted
- Install KRO, ARC, LocalStack, and ACK EC2 controller
- Configure everything to work together

### 2. Verify LocalStack is Running

```bash
kubectl get pods -n localstack
kubectl logs -n localstack -l app.kubernetes.io/name=localstack
```

LocalStack should be running and the logs should show EC2 service available.

### 3. Verify ACK EC2 Controller

```bash
kubectl get pods -n ack-system
kubectl logs -n ack-system -l app.kubernetes.io/name=ec2-chart
```

The ACK controller should be running and configured to use LocalStack endpoint.

## Testing EC2 Runners

### 1. Create a GitHub Runner AMI in LocalStack

LocalStack recognizes Docker images tagged with the pattern `localstack-ec2/<AmiName>:<AmiId>` as AMIs.

```bash
# Tag the GitHub Actions runner image as an AMI
docker exec slapchop-test-control-plane docker tag \
  ghcr.io/actions/actions-runner:latest \
  localstack-ec2/github-runner-ami:ami-000001
```

### 2. Create EC2Runner ResourceGraphDefinition

See `ec2-runner-rgd.yaml` (to be created) for the RGD that:
- References the ARC-created JIT secret
- Creates an ACK EC2 Instance
- Passes JIT config as user data
- Uses `readyWhen` to detect when the instance terminates

### 3. Deploy and Test

```bash
# Apply the EC2Runner RGD
kubectl apply -f ec2-runner-rgd.yaml

# Update the ARC scale set to use EC2Runner instead of PodRunner
# (modify arc-scale-set-values-local.yaml)

# Trigger a workflow in GitHub
# Watch the EC2 instances get created
kubectl get instances -n arc-runners
```

## Architecture

```
ARC Orchestrator Pod (kro-actions-runner)
    ↓ (creates JIT secret)
    ↓
EC2Runner ResourceGraph (KRO)
    ↓
ACK EC2 Instance Resource
    ↓ (ACK controller talks to LocalStack)
    ↓
LocalStack EC2 Service
    ↓
Docker Container (fake EC2 instance)
    ↓
GitHub Actions Runner (runs workflow)
```

## Key Differences from Pod Runners

### Pod Runners
- **Fast**: Pods start in seconds
- **Simple**: No additional controllers needed
- **Ephemeral**: Natural Kubernetes lifecycle

### EC2 Runners (via LocalStack)
- **Slower**: Containers-as-instances take longer to provision
- **Complex**: Requires LocalStack + ACK + Docker socket
- **Realistic**: Mimics real EC2 behavior for testing
- **Self-terminating**: User data script calls `shutdown -h now` after workflow

## Troubleshooting

### LocalStack Pod Not Starting

**Error**: `hostPath type check failed: /var/run/docker.sock is not a socket file`

**Solution**: Ensure you created the cluster with `kind-localstack.yaml`:
```bash
kind delete cluster --name slapchop-test
KIND_CONFIG=kind-localstack.yaml ./setup-cluster.sh
```

### ACK Controller Can't Reach LocalStack

**Check endpoint configuration**:
```bash
kubectl get configmap -n ack-system ack-ec2-controller-ec2-chart-config -o yaml
```

Should show: `AWS_ENDPOINT_URL: http://localstack.localstack.svc.cluster.local:4566`

### EC2 Instances Not Launching

**Check LocalStack EC2 service**:
```bash
kubectl exec -n localstack deploy/localstack -- awslocal ec2 describe-images
```

Should show your AMI with ID `ami-000001`.

### Runner Not Connecting to GitHub

**Check user data script**:
The user data must:
1. Extract JIT config
2. Run `/home/runner/run.sh`
3. Call `shutdown -h now` on completion

## Configuration Files

- **kind-localstack.yaml**: Kind cluster config with Docker socket mount
- **localstack-values.yaml**: LocalStack Helm values
- **ack-ec2-values.yaml**: ACK EC2 controller Helm values
- **ec2-runner-rgd.yaml**: EC2Runner ResourceGraphDefinition (to be created)
- **setup-cluster.sh**: Automated setup script

## Cleanup

```bash
kind delete cluster --name slapchop-test
```

This removes everything including all Docker containers created by LocalStack.
