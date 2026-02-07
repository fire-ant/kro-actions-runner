#!/bin/bash
# Test EC2 Runner with LocalStack
# This script tests the complete EC2 runner workflow

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "=== Testing EC2 Runner with LocalStack ==="
echo ""

echo "Step 1: Applying VPC Network RGD..."
kubectl apply -f test/fixtures/rgds/vpc-network-rgd.yaml

echo "Step 2: Applying EC2 Runner with VPC reference RGD..."
kubectl apply -f test/fixtures/rgds/ec2-runner-with-vpc-ref-rgd.yaml

echo "Step 3: Creating VPC Network ResourceGraph..."
kubectl apply -f test/fixtures/instances/test-vpc-network.yaml

echo "Step 4: Waiting for VPC network to be ready (this may take a minute)..."
kubectl wait --for=condition=ReconciliationSucceeded --timeout=300s \
    resourcegraph/test-vpc-network || {
    echo "VPC Network not ready, checking status..."
    kubectl get resourcegraph test-vpc-network -o yaml
    exit 1
}

echo "Step 5: Creating mock JIT secret..."
kubectl apply -f test/fixtures/instances/test-ec2-runner-secret.yaml

echo "Step 6: Creating EC2 Runner instance..."
kubectl apply -f test/fixtures/instances/test-ec2-runner-instance.yaml

echo ""
echo "=== Test resources created! ==="
echo ""
echo "Watch ResourceGraph:"
echo "  kubectl get resourcegraph test-ec2-runner -w"
echo ""
echo "Watch ACK Instance:"
echo "  kubectl get instances -n default -w"
echo ""
echo "Check LocalStack EC2 instances:"
echo "  kubectl exec -n localstack deploy/localstack -- awslocal ec2 describe-instances"
echo ""
echo "View ACK controller logs:"
echo "  mise run ack:logs"
echo ""
