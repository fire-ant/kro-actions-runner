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

CLUSTER_NAME="${KIND_CLUSTER_NAME:-kro-test}"
IMAGE="${IMAGE:-kro-actions-runner:latest}"

echo "Loading image ${IMAGE} into kind cluster ${CLUSTER_NAME}..."

# Check if image exists
if ! docker images --format "{{.Repository}}:{{.Tag}}" | grep -q "^${IMAGE}$"; then
    echo "Error: Image ${IMAGE} not found. Build it first with: docker build -t ${IMAGE} ."
    exit 1
fi

kind load docker-image "${IMAGE}" --name "${CLUSTER_NAME}"
echo "Image loaded successfully"
