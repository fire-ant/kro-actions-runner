#!/bin/bash
# Watch EC2 instances with detailed status information

NAMESPACE="${1:-default}"

echo "Watching EC2 instances in namespace: $NAMESPACE"
echo ""

# Use custom columns to show instance state
kubectl get instances -n "$NAMESPACE" \
    -o custom-columns=NAME:.metadata.name,INSTANCE_ID:.status.instanceID,STATE:.status.state.name,TYPE:.spec.instanceType,AMI:.spec.imageID,AGE:.metadata.creationTimestamp \
    --watch
