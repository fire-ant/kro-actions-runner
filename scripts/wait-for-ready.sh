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

# Generic script to wait for Kubernetes resources to be ready
# Usage: wait-for-ready.sh <resource-type> <resource-name> <namespace> [timeout]

RESOURCE_TYPE="${1:-}"
RESOURCE_NAME="${2:-}"
NAMESPACE="${3:-default}"
TIMEOUT="${4:-300s}"

if [[ -z ${RESOURCE_TYPE} ]] || [[ -z ${RESOURCE_NAME} ]]; then
    echo "Usage: $0 <resource-type> <resource-name> <namespace> [timeout]"
    echo "Example: $0 deployment kro kro-system 300s"
    exit 1
fi

echo "Waiting for ${RESOURCE_TYPE}/${RESOURCE_NAME} in namespace ${NAMESPACE} to be ready (timeout: ${TIMEOUT})..."

kubectl wait --for=condition=available \
    --timeout="${TIMEOUT}" \
    "${RESOURCE_TYPE}/${RESOURCE_NAME}" \
    -n "${NAMESPACE}"

echo "${RESOURCE_TYPE}/${RESOURCE_NAME} is ready"
