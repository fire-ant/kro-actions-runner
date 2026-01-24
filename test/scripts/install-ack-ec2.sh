#!/bin/bash
# Install ACK EC2 Controller configured for LocalStack
# This controller allows managing EC2 resources via Kubernetes CRDs

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURES_DIR="$SCRIPT_DIR/../fixtures/localstack"

ACK_SYSTEM_NAMESPACE="ack-system"
SERVICE="ec2"
AWS_REGION="us-east-1"
RELEASE_VERSION="${ACK_EC2_VERSION:-1.9.2}"

echo "Installing ACK EC2 Controller..."

# Create namespace for ACK controllers
echo "Creating $ACK_SYSTEM_NAMESPACE namespace..."
kubectl create namespace "$ACK_SYSTEM_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# Create dummy AWS credentials secret (required by ACK, but not used by LocalStack)
# The ACK controller expects a credentials file in AWS CLI format
echo "Creating AWS credentials secret (dummy values for LocalStack)..."
cat > /tmp/ack-credentials <<EOF
[default]
aws_access_key_id = test
aws_secret_access_key = test
EOF

kubectl create secret generic ack-ec2-user-secrets \
    --namespace "$ACK_SYSTEM_NAMESPACE" \
    --from-file=credentials=/tmp/ack-credentials \
    --dry-run=client -o yaml | kubectl apply -f -

rm -f /tmp/ack-credentials

# Install ACK EC2 controller from official OCI registry
echo "Installing ACK EC2 controller version $RELEASE_VERSION..."
helm upgrade --install ack-$SERVICE-controller \
    oci://public.ecr.aws/aws-controllers-k8s/$SERVICE-chart \
    --version="$RELEASE_VERSION" \
    --namespace "$ACK_SYSTEM_NAMESPACE" \
    --create-namespace \
    --set=aws.region="$AWS_REGION" \
    --values "$FIXTURES_DIR/ack-ec2-values.yaml" \
    --wait \
    --timeout 5m

echo "Waiting for ACK EC2 controller to be ready..."
kubectl wait --for=condition=available --timeout=300s \
    deployment/ack-ec2-controller -n "$ACK_SYSTEM_NAMESPACE" || {
    echo "Warning: ACK EC2 controller not ready within timeout, checking logs..."
    kubectl logs -n "$ACK_SYSTEM_NAMESPACE" -l app.kubernetes.io/name=ec2-chart --tail=50 || true
}

echo "ACK EC2 controller installed successfully"
echo "Configured to use LocalStack endpoint: https://localstack.localstack.svc.cluster.local:4566"
