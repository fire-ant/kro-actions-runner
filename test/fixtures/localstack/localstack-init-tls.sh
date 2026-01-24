#!/bin/bash
# LocalStack init hook to configure custom TLS certificates
# This script generates a self-signed certificate for HTTPS support

set -e

echo "Generating self-signed TLS certificate for LocalStack..."

# Create directory for certificates
mkdir -p /tmp/localstack/tls

# Generate self-signed certificate
# Subject Alternative Names include common LocalStack hostnames
cat >/tmp/localstack/tls/openssl.cnf <<EOF
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

# Generate private key and certificate
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /tmp/localstack/tls/server.key \
    -out /tmp/localstack/tls/server.crt \
    -config /tmp/localstack/tls/openssl.cnf

# Set permissions
chmod 644 /tmp/localstack/tls/server.crt
chmod 600 /tmp/localstack/tls/server.key

echo "TLS certificate generated successfully"
echo "Certificate: /tmp/localstack/tls/server.crt"
echo "Private key: /tmp/localstack/tls/server.key"
