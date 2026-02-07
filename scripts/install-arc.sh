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

CONTROLLER_NS="${CONTROLLER_NS:-arc-systems}"

echo "Installing Actions Runner Controller in namespace ${CONTROLLER_NS}..."
helm upgrade --install arc \
    --namespace "${CONTROLLER_NS}" \
    --create-namespace \
    --wait \
    --timeout=5m \
    oci://ghcr.io/actions/actions-runner-controller-charts/gha-runner-scale-set-controller

echo "Waiting for ARC deployment to be available..."
kubectl wait --for=condition=available \
    --timeout=300s \
    deployment/arc-gha-rs-controller \
    -n "${CONTROLLER_NS}"

echo "ARC installed successfully"
