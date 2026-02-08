#!/bin/bash
# Clean up EC2 integration test resources

set -e

NAMESPACE="${1:-arc-runners}"

echo "=== Cleaning EC2 Integration Test Resources ==="
echo ""

echo "Deleting runner pods..."
kubectl delete pod -l actions.github.com/scale-set-name -n "$NAMESPACE" --ignore-not-found --force --grace-period=0 || true

echo ""
echo "Deleting EC2Runner ResourceGraphs..."
kubectl delete ec2runner --all -n "$NAMESPACE" --ignore-not-found --timeout=30s || true

echo ""
echo "Force deleting stuck EC2Runners..."
kubectl get ec2runner -n "$NAMESPACE" -o name 2>/dev/null | while read -r rg; do
    kubectl patch "$rg" -n "$NAMESPACE" -p '{"metadata":{"finalizers":[]}}' --type=merge 2>/dev/null || true
    kubectl delete "$rg" -n "$NAMESPACE" --force --grace-period=0 2>/dev/null || true
done

echo ""
echo "Deleting EC2 Instances..."
kubectl delete instances --all -n "$NAMESPACE" --ignore-not-found --timeout=30s || true

echo ""
echo "Force deleting stuck Instances..."
kubectl get instances -n "$NAMESPACE" -o name 2>/dev/null | while read -r inst; do
    kubectl patch "$inst" -n "$NAMESPACE" -p '{"metadata":{"finalizers":[]}}' --type=merge 2>/dev/null || true
    kubectl delete "$inst" -n "$NAMESPACE" --force --grace-period=0 2>/dev/null || true
done

echo ""
echo "Deleting EC2RunnerInfra..."
kubectl delete ec2runnerinfra --all -n "$NAMESPACE" --ignore-not-found --timeout=30s || true

echo ""
echo "Force deleting stuck EC2RunnerInfra..."
kubectl get ec2runnerinfra -n "$NAMESPACE" -o name 2>/dev/null | while read -r vpc; do
    kubectl patch "$vpc" -n "$NAMESPACE" -p '{"metadata":{"finalizers":[]}}' --type=merge 2>/dev/null || true
    kubectl delete "$vpc" -n "$NAMESPACE" --force --grace-period=0 2>/dev/null || true
done

echo ""
echo "Deleting ACK EC2 resources (Subnet, SecurityGroup, VPC)..."
kubectl delete subnet --all -n "$NAMESPACE" --ignore-not-found --timeout=30s || true
kubectl delete securitygroup --all -n "$NAMESPACE" --ignore-not-found --timeout=30s || true
kubectl delete vpc --all -n "$NAMESPACE" --ignore-not-found --timeout=30s || true

echo ""
echo "Force deleting stuck ACK resources..."
for resource in subnet securitygroup vpc; do
    kubectl get "$resource" -n "$NAMESPACE" -o name 2>/dev/null | while read -r res; do
        kubectl patch "$res" -n "$NAMESPACE" -p '{"metadata":{"finalizers":[]}}' --type=merge 2>/dev/null || true
        kubectl delete "$res" -n "$NAMESPACE" --force --grace-period=0 2>/dev/null || true
    done
done

echo ""
echo ""
echo "Waiting for cleanup to complete..."
sleep 5

echo ""
echo "=== Remaining Resources ==="
echo "EC2Runners:"
kubectl get ec2runner -n "$NAMESPACE" 2>/dev/null || echo "  None"
echo ""
echo "Instances:"
kubectl get instances -n "$NAMESPACE" 2>/dev/null || echo "  None"
echo ""
echo "EC2RunnerInfra:"
kubectl get ec2runnerinfra -n "$NAMESPACE" 2>/dev/null || echo "  None"
echo ""
echo "Pods:"
kubectl get pods -n "$NAMESPACE" 2>/dev/null || echo "  None"
echo ""

echo "Cleaning up LocalStack (restarting to clear in-memory state)..."
kubectl rollout restart deployment -n localstack localstack
kubectl rollout status deployment -n localstack localstack --timeout=60s
echo "  ✅ LocalStack restarted"

echo ""
echo "✅ Cleanup complete!"
echo ""
echo "Note: If resources are still stuck, you may need to restart the controllers:"
echo "  kubectl rollout restart deployment -n kro-system kro"
echo "  kubectl rollout restart deployment -n ack-system ack-ec2-controller-ec2-chart"
