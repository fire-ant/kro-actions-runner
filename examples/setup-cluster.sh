#!/bin/bash
# Setup script for slapchop-test kind cluster with KRO, ARC, and LocalStack
# This script will recreate the cluster from scratch with all components

set -e

CLUSTER_NAME="${CLUSTER_NAME:-slapchop-test}"
SKIP_LOCALSTACK="${SKIP_LOCALSTACK:-false}"
KIND_CONFIG="${KIND_CONFIG:-kind-basic.yaml}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== Setting up $CLUSTER_NAME cluster ==="
echo "Skip LocalStack: $SKIP_LOCALSTACK"

# Step 1: Delete existing cluster if it exists
echo "Step 1: Cleaning up existing cluster..."
if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
    echo "Deleting existing cluster: $CLUSTER_NAME"
    kind delete cluster --name "$CLUSTER_NAME"
fi

# Step 2: Create new kind cluster
echo "Step 2: Creating kind cluster..."
kind create cluster --name "$CLUSTER_NAME" --config "$SCRIPT_DIR/$KIND_CONFIG"

# Step 3: Install KRO
echo "Step 3: Installing KRO..."
helm install kro oci://registry.k8s.io/kro/charts/kro --namespace kro-system --create-namespace

# Wait for KRO to be ready
echo "Waiting for KRO controller to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment.apps/kro -n kro-system

# Step 4: Install ARC
echo "Step 4: Installing Actions Runner Controller..."
CONTROLLER_NS="arc-systems"
helm install arc \
    --namespace "${CONTROLLER_NS}" \
    --create-namespace \
    oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set-controller

echo "Waiting for ARC controller to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/arc-gha-rs-controller -n "$CONTROLLER_NS"

# Step 5: Create runner namespace and RBAC
echo "Step 5: Setting up runner namespace and RBAC..."
RUNNER_NS="arc-runners"
kubectl create namespace "$RUNNER_NS" --dry-run=client -o yaml | kubectl apply -f -

# Apply RBAC from manifest (update namespace in ClusterRoleBinding)
sed "s/namespace: arc-runners/namespace: $RUNNER_NS/g" "$SCRIPT_DIR/kro-runner-rbac.yaml" | kubectl apply -n "$RUNNER_NS" -f -

# Step 6: Apply Pod Runner RGD
echo "Step 6: Applying Pod Runner ResourceGraphDefinition..."
kubectl apply -f "$SCRIPT_DIR/pod-runner-rgd.yaml"

# Step 7: Load kro-actions-runner image
echo "Step 7: Loading kro-actions-runner image..."
if docker images | grep -q "kro-actions-runner"; then
    kind load docker-image kro-actions-runner:latest --name "$CLUSTER_NAME"
else
    echo "Warning: kro-actions-runner:latest image not found locally. Please build it first."
fi

# Step 8: Install runner scale set
echo "Step 8: Installing ARC runner scale set..."
if [ ! -f "$SCRIPT_DIR/arc-scale-set-values-local.yaml" ]; then
    echo "Error: arc-scale-set-values-local.yaml not found"
    exit 1
fi

helm upgrade --install \
    --namespace "$RUNNER_NS" \
    --create-namespace \
    --values "$SCRIPT_DIR/arc-scale-set-values-local.yaml" \
    kro-runner-set \
    oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set

echo "Waiting for runner scale set to be ready..."
sleep 10

# Step 9: Install LocalStack (optional)
if [ "$SKIP_LOCALSTACK" != "true" ]; then
    echo "Step 9: Installing LocalStack..."
    helm repo add localstack https://localstack.github.io/helm-charts || true
    helm repo update

    kubectl create namespace localstack --dry-run=client -o yaml | kubectl apply -f -

    # Create ConfigMap with TLS init hook
    echo "Creating LocalStack TLS init hook ConfigMap..."
    kubectl create configmap localstack-init-hooks \
        --namespace localstack \
        --from-file=01-init-tls.sh="$SCRIPT_DIR/localstack-init-tls.sh" \
        --dry-run=client -o yaml | kubectl apply -f -

    # Note: The init hook will be mounted at /etc/localstack/init/ready.d/
    # For TLS to work, the certificate must be generated before LocalStack starts its server
    # We mount to ready.d but the actual certificate location is what matters

    helm install localstack localstack/localstack \
        --namespace localstack \
        --values "$SCRIPT_DIR/localstack-values.yaml"

    echo "Waiting for LocalStack to be ready..."
    kubectl wait --for=condition=ready --timeout=120s pod -l app.kubernetes.io/name=localstack -n localstack || true
else
    echo "Step 9: Skipping LocalStack installation"
fi

# Step 10: Install ACK EC2 Controller (optional, only with LocalStack)
if [ "$SKIP_LOCALSTACK" != "true" ]; then
    echo "Step 10: Installing ACK EC2 Controller..."

    ACK_SYSTEM_NAMESPACE="ack-system"
    SERVICE="ec2"
    AWS_REGION="us-east-1"
    RELEASE_VERSION="${ACK_EC2_VERSION:-v1.9.0}"

    # Create namespace for ACK controllers
    kubectl create namespace "$ACK_SYSTEM_NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

    # Create dummy AWS credentials secret (required by ACK, but not used by LocalStack)
    kubectl create secret generic ack-ec2-user-secrets \
        --namespace "$ACK_SYSTEM_NAMESPACE" \
        --from-literal=AWS_ACCESS_KEY_ID=test \
        --from-literal=AWS_SECRET_ACCESS_KEY=test \
        --dry-run=client -o yaml | kubectl apply -f -

    # Install ACK EC2 controller from official OCI registry
    echo "Installing ACK EC2 controller version $RELEASE_VERSION from OCI registry..."
    helm install --create-namespace \
        -n "$ACK_SYSTEM_NAMESPACE" \
        ack-$SERVICE-controller \
        oci://public.ecr.aws/aws-controllers-k8s/$SERVICE-chart \
        --version="$RELEASE_VERSION" \
        --set=aws.region="$AWS_REGION" \
        --values "$SCRIPT_DIR/ack-ec2-values.yaml"

    echo "Waiting for ACK EC2 controller to be ready..."
    kubectl wait --for=condition=available --timeout=120s deployment/ack-ec2-controller -n "$ACK_SYSTEM_NAMESPACE" || true
else
    echo "Step 10: Skipping ACK EC2 Controller installation (requires LocalStack)"
fi

echo ""
echo "=== Setup Complete! ==="
echo ""
echo "Cluster: $CLUSTER_NAME"
echo "KRO: kro-system namespace"
echo "ARC Controller: $CONTROLLER_NS namespace"
echo "ARC Runners: $RUNNER_NS namespace"
if [ "$SKIP_LOCALSTACK" != "true" ]; then
    echo "LocalStack: localstack namespace"
    echo "ACK EC2: ack-system namespace"
fi
echo ""
echo "Check status with:"
echo "  kubectl get pods -n kro-system"
echo "  kubectl get pods -n $CONTROLLER_NS"
echo "  kubectl get pods -n $RUNNER_NS"
if [ "$SKIP_LOCALSTACK" != "true" ]; then
    echo "  kubectl get pods -n localstack"
    echo "  kubectl get pods -n ack-system"
fi
echo "  kubectl get podrunner -n $RUNNER_NS"
