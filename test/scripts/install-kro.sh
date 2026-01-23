#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2025
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -euo pipefail

echo "Installing KRO..."
helm install kro oci://registry.k8s.io/kro/charts/kro \
    --namespace kro-system \
    --create-namespace \
    --wait \
    --timeout=5m

echo "Waiting for KRO deployment to be available..."
kubectl wait --for=condition=available \
    --timeout=300s \
    deployment.apps/kro \
    -n kro-system

echo "KRO installed successfully"
