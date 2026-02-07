#!/bin/bash
# Install ACK EC2 Controller configured for LocalStack
# Simplified configuration without TLS complexity - uses plain HTTP endpoint

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

ACK_SYSTEM_NAMESPACE="ack-system"
SERVICE="ec2"
AWS_REGION="us-east-1"
RELEASE_VERSION="${ACK_EC2_VERSION:-1.9.2}"

echo "Installing ACK EC2 Controller..."

# Create namespace for ACK controllers
echo "Creating $ACK_SYSTEM_NAMESPACE namespace..."
kubectl create namespace "$ACK_SYSTEM_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# Create dummy AWS credentials secret (required by ACK chart, but not validated by LocalStack)
# The ACK controller expects a credentials file in AWS CLI format
echo "Creating AWS credentials secret (dummy values for LocalStack)..."
cat >/tmp/ack-credentials <<EOF
[default]
aws_access_key_id = test
aws_secret_access_key = test
EOF

kubectl create secret generic ack-ec2-user-secrets \
    --namespace "$ACK_SYSTEM_NAMESPACE" \
    --from-file=credentials=/tmp/ack-credentials \
    --dry-run=client -o yaml | kubectl apply -f -

rm -f /tmp/ack-credentials

# Install ACK EC2 controller from official OCI registry using anonymous access
echo "Installing ACK EC2 controller version $RELEASE_VERSION (using anonymous access to ECR public)..."

# Unset AWS credentials to force anonymous access to public ECR
unset AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY AWS_SESSION_TOKEN

# Logout from any existing helm registry authentication
helm registry logout public.ecr.aws 2>/dev/null || true

helm upgrade --install ack-$SERVICE-controller \
    oci://public.ecr.aws/aws-controllers-k8s/$SERVICE-chart \
    --version="$RELEASE_VERSION" \
    --namespace "$ACK_SYSTEM_NAMESPACE" \
    --create-namespace \
    --set=aws.region="$AWS_REGION" \
    --values "$PROJECT_ROOT/manifests/ack-ec2-values.yaml" \
    --wait \
    --timeout 5m

echo "Waiting for ACK EC2 controller to be ready..."
kubectl wait --for=condition=available --timeout=300s \
    deployment -l app.kubernetes.io/name=ec2-chart -n "$ACK_SYSTEM_NAMESPACE" || {
    echo "Warning: ACK EC2 controller not ready within timeout, checking logs..."
    kubectl get deployments -n "$ACK_SYSTEM_NAMESPACE"
    kubectl logs -n "$ACK_SYSTEM_NAMESPACE" -l app.kubernetes.io/name=ec2-chart --tail=50 || true
    exit 1
}

echo "ACK EC2 controller installed successfully!"
echo "Configured to use LocalStack endpoint: http://localstack.localstack.svc.cluster.local:4566"
