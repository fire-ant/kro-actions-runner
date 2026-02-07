#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2025
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

# Build and push kro-actions-runner image to local registry
# This is needed so the runner pods can pull the image from the registry

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

IMAGE_NAME="${IMAGE_NAME:-kro-actions-runner}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
REGISTRY_HOST="${REGISTRY_HOST:-localhost:5005}"

FULL_IMAGE="${REGISTRY_HOST}/${IMAGE_NAME}:${IMAGE_TAG}"

echo "Building and pushing image ${FULL_IMAGE}..."

# Build the image
echo "Building image..."
cd "$PROJECT_ROOT"
docker build -t "${FULL_IMAGE}" .

# Push to local registry
echo "Pushing image to local registry..."
docker push "${FULL_IMAGE}"

echo "Image ${FULL_IMAGE} successfully built and pushed to local registry"
echo "You can now use image: ${FULL_IMAGE} in your pod specs"
