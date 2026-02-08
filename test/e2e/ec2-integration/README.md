# EC2 Runner Integration Test

Complete end-to-end integration test for the unified EC2Runner approach.

## What This Tests

This test validates the complete lifecycle of EC2 instances created through the unified RGD:

1. **Infrastructure Setup**
   - Creates VPCNetwork (provides subnet and security group)
   - Creates ScaleSet configuration (minRunners=3, maxRunners=10)

2. **Job Simulation**
   - Creates 5 simulated ARC runner pods (integration-test-runners-0 through integration-test-runners-4)
   - Creates EC2Runner ResourceGraphs (simulating what kar would do)
   - Verifies instances are created with correct types:
     - `integration-test-runners-0, 1, 2` → **reusable** (index < minRunners)
     - `integration-test-runners-3, 4` → **ephemeral** (index >= minRunners)

3. **Policy Verification**
   - Reusable instances have `adoption-policy: adopt-or-create` and `deletion-policy: retain`
   - Ephemeral instances have no special annotations (default behavior)
   - All instances reach `running` state

4. **Job Completion**
   - Sleeps 5 seconds (simulates job runtime)
   - Deletes runner pods (simulates ARC cleaning up)
   - Deletes EC2Runner ResourceGraphs (simulates kar cleanup)

5. **Lifecycle Verification**
   - **Reusable instances**: Retained with same instance IDs
   - **Ephemeral instances**: Deleted or terminating
   - Validates in both Kubernetes and LocalStack

## Test Flow

```
VPC Setup → Job Arrival → Instances Running → Job Complete → Lifecycle Verified
            (5 pods)      (3 reusable,       (delete pods    (reusable stay,
                          2 ephemeral)        & RGs)          ephemeral go)
```

## Running the Test

### Prerequisites

```bash
# Setup cluster with all components
mise run cluster:setup:full

# Deploy unified RGD
kubectl apply -f examples/ec2-runner/ec2-runner-config-rgd.yaml
```

### Run Integration Test

```bash
# Run the full integration test
mise run test:ec2:integration

# Or run with kubectl kuttl directly
kubectl kuttl test --config test/e2e/ec2-integration/kuttl/kuttl-test.yaml
```

### Watch Instance Lifecycle

In another terminal, watch instances during the test:

```bash
mise run warmpool:watch:instances
```

## Expected Results

### Step 1: Instances Created

```
NAME                           ID                TYPE        INDEX  STATE
integration-test-runners-0     i-abc123...       reusable    0      running
integration-test-runners-1     i-def456...       reusable    1      running
integration-test-runners-2     i-ghi789...       reusable    2      running
integration-test-runners-3     i-jkl012...       ephemeral   3      running
integration-test-runners-4     i-mno345...       ephemeral   4      running
```

### Step 2: After Job Completion

```
NAME                           ID                TYPE        STATE       DELETION
integration-test-runners-0     i-abc123...       reusable    running     <none>
integration-test-runners-1     i-def456...       reusable    running     <none>
integration-test-runners-2     i-ghi789...       reusable    running     <none>
integration-test-runners-3     <deleted>
integration-test-runners-4     <deleted>
```

## Success Criteria

- ✅ VPC infrastructure created
- ✅ 5 instances created (3 reusable, 2 ephemeral)
- ✅ Reusable instances have adoption and retention annotations
- ✅ Ephemeral instances have no special annotations
- ✅ All instances reach running state
- ✅ After deletion:
  - Reusable instances retained with same IDs
  - Ephemeral instances deleted

## Troubleshooting

### Instances stuck in pending

Check ACK controller logs:
```bash
kubectl logs -n ack-system -l app.kubernetes.io/name=ec2-chart --tail=50
```

### VPC not created

Check KRO controller logs:
```bash
kubectl logs -n kro-system -l app=kro --tail=50
```

### LocalStack issues

Restart LocalStack:
```bash
kubectl rollout restart deployment -n localstack localstack
kubectl wait --for=condition=ready pod -n localstack -l app=localstack --timeout=60s
```

## Cleanup

The test automatically cleans up resources. To manually clean up:

```bash
kubectl delete namespace arc-runners --force --grace-period=0
kubectl create namespace arc-runners
```
