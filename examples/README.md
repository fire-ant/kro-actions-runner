# KRO Actions Runner Examples

This directory contains example configurations for using KRO Actions Runner with different compute backends.

## Quick Start

Choose your compute backend:

### Pod Runners (Simplest)
Run GitHub Actions workflows in Kubernetes pods:
```bash
# Apply infrastructure
kubectl apply -f ../manifests/kro-runner-rbac.yaml

# Apply RGD
kubectl apply -f pod-runner/pod-runner-rgd.yaml

# Create test instance
kubectl apply -f pod-runner/test-runner-secret.yaml
```

### EC2 Runners (LocalStack Testing)
Run workflows in EC2 instances (local testing with LocalStack):
```bash
# Setup infrastructure (once)
mise run dev:setup

# Deploy VPC Network
kubectl apply -f ec2-runner/vpc-network-rgd.yaml
kubectl apply -f ec2-runner/test-vpc-network.yaml

# Deploy EC2 Runner
kubectl apply -f ec2-runner/ec2-runner-rgd.yaml
kubectl apply -f ec2-runner/test-ec2-runner-secret.yaml
kubectl apply -f ec2-runner/test-ec2-runner-instance.yaml
```

## Directory Structure

```
examples/
├── README.md            # This file
├── pod-runner/          # Kubernetes pod-based runners
│   ├── pod-runner-rgd.yaml
│   └── test-runner-secret.yaml
└── ec2-runner/          # EC2-based runners (LocalStack/AWS)
    ├── vpc-network-rgd.yaml
    ├── ec2-runner-rgd.yaml
    ├── test-vpc-network.yaml
    ├── test-ec2-runner-secret.yaml
    └── test-ec2-runner-instance.yaml

../manifests/            # Infrastructure & configuration (see manifests/)
├── localstack.yaml
├── ack-ec2-values.yaml
├── arc-scale-set-values.yaml
└── kro-runner-rbac.yaml
```

## Production Use

For production deployments:

1. **Pod Runners**: Work as-is in any Kubernetes cluster
2. **EC2 Runners**:
   - Replace LocalStack endpoint with real AWS
   - Use AWS Secrets Manager instead of userData for JIT config
   - Configure proper IAM roles and permissions

See main [README.md](../README.md) for full documentation.
