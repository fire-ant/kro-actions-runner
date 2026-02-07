#!/bin/bash
# Install LocalStack using plain Kubernetes manifest
# Simplified setup without TLS complexity - uses plain HTTP

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "Installing LocalStack..."

# Apply the manifest
kubectl apply -f "$PROJECT_ROOT/manifests/localstack.yaml"

echo "Waiting for LocalStack to be ready..."
kubectl wait --for=condition=ready --timeout=300s pod -l app=localstack -n localstack || {
    echo "Warning: LocalStack pod not ready within timeout, checking logs..."
    kubectl logs -n localstack -l app=localstack --tail=50 || true
    exit 1
}

echo "LocalStack installed successfully!"
echo "HTTP endpoint: http://localstack.localstack.svc.cluster.local:4566"
echo "External access (NodePort): http://localhost:30566"
echo ""

# Wait a bit for init script to complete
echo "Waiting for initialization script to complete..."
sleep 10

# Display VPC and Subnet information
echo "Checking initialized EC2 resources..."
echo ""
echo "VPCs:"
# shellcheck disable=SC2016
kubectl exec -n localstack deploy/localstack -- awslocal ec2 describe-vpcs \
    --query 'Vpcs[*].[VpcId,CidrBlock,Tags[?Key==`Name`].Value | [0]]' \
    --output table 2>/dev/null || echo "  (VPC creation may still be in progress)"

echo ""
echo "Subnets:"
kubectl exec -n localstack deploy/localstack -- awslocal ec2 describe-subnets \
    --query 'Subnets[*].[SubnetId,VpcId,CidrBlock]' \
    --output table 2>/dev/null || echo "  (Subnet creation may still be in progress)"

echo ""
echo "To get the subnet ID for EC2Runner instances:"
echo "  SUBNET_ID=\$(kubectl exec -n localstack deploy/localstack -- awslocal ec2 describe-subnets --query 'Subnets[0].SubnetId' --output text)"
