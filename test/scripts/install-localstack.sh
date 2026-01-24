#!/bin/bash
# Install LocalStack with HTTPS support for testing
# This script sets up LocalStack with self-signed TLS certificates

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURES_DIR="$SCRIPT_DIR/../fixtures/localstack"

echo "Installing LocalStack..."

# Add helm repo
echo "Adding LocalStack helm repository..."
helm repo add localstack https://localstack.github.io/helm-charts || true
helm repo update

# Create namespace
echo "Creating localstack namespace..."
kubectl create namespace localstack --dry-run=client -o yaml | kubectl apply -f -

# Generate self-signed TLS certificate using openssl directly
echo "Generating self-signed TLS certificate..."
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

cat >"$TMP_DIR/openssl.cnf" <<EOF
[req]
distinguished_name = req_distinguished_name
x509_extensions = v3_req
prompt = no

[req_distinguished_name]
CN = localstack.localstack.svc.cluster.local

[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = localstack
DNS.3 = localstack.localstack
DNS.4 = localstack.localstack.svc
DNS.5 = localstack.localstack.svc.cluster.local
IP.1 = 127.0.0.1
EOF

# Generate certificate and key
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout "$TMP_DIR/server.key" \
    -out "$TMP_DIR/server.crt" \
    -config "$TMP_DIR/openssl.cnf" 2>/dev/null

# Combine cert and key into a single PEM file (LocalStack expects combined format)
cat "$TMP_DIR/server.crt" "$TMP_DIR/server.key" > "$TMP_DIR/server.pem"

echo "TLS certificate generated successfully"

# Create Kubernetes secret with TLS certificate
echo "Creating TLS certificate secret..."
kubectl create secret generic localstack-tls-certs \
    --namespace localstack \
    --from-file=server.pem="$TMP_DIR/server.pem" \
    --from-file=server.crt="$TMP_DIR/server.crt" \
    --from-file=server.key="$TMP_DIR/server.key" \
    --dry-run=client -o yaml | kubectl apply -f -

# Install LocalStack
echo "Installing LocalStack helm chart..."
helm upgrade --install localstack localstack/localstack \
    --namespace localstack \
    --values "$FIXTURES_DIR/localstack-values.yaml" \
    --wait \
    --timeout 5m

echo "Waiting for LocalStack to be ready..."
kubectl wait --for=condition=ready --timeout=300s pod -l app.kubernetes.io/name=localstack -n localstack || {
    echo "Warning: LocalStack pod not ready within timeout, checking logs..."
    kubectl logs -n localstack -l app.kubernetes.io/name=localstack --tail=50 || true
}

echo "LocalStack installed successfully"
echo "HTTPS endpoint: https://localstack.localstack.svc.cluster.local:4566"
