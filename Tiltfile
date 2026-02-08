load('ext://helm_remote', 'helm_remote')

config.define_bool('enable-localstack', args=False, usage='Enable LocalStack for EC2 simulation')
cfg = config.parse()

ENABLE_LOCALSTACK = cfg.get('enable-localstack', True)
NAMESPACE = 'arc-runners'
CONTROLLER_NS = 'arc-systems'

allow_k8s_contexts(['kind-gha-arc-kro-runner'])

default_registry(
    'localhost:5005',
    host_from_cluster='kro-test-registry:5000'
)

def with_mise(cmd):
    return 'export PATH="$HOME/.local/share/mise/shims:$PATH" && ' + cmd

# Build and load Docker image into kind cluster
# Using custom_build to load directly into kind (faster than registry push)
custom_build(
    'localhost:5005/kro-actions-runner',
    'docker build -t $EXPECTED_REF . && kind load docker-image $EXPECTED_REF --name gha-arc-kro-runner',
    deps=['cmd/', 'internal/', 'go.mod', 'go.sum', 'Dockerfile'],
)

# Dependencies: LocalStack → ACK → KRO → ARC
if ENABLE_LOCALSTACK:
    k8s_yaml('manifests/localstack.yaml')
    k8s_resource('localstack', port_forwards='4566:4566', labels=['dependencies', '1-localstack'])

    k8s_yaml(blob('apiVersion: v1\nkind: Namespace\nmetadata:\n  name: ack-system'))
    k8s_yaml(blob('apiVersion: v1\nkind: Secret\nmetadata:\n  name: ack-ec2-user-secrets\n  namespace: ack-system\ntype: Opaque\nstringData:\n  credentials: |\n    [default]\n    aws_access_key_id = test\n    aws_secret_access_key = test'))

    helm_remote('ec2-chart', release_name='ack-ec2-controller', repo_url='oci://public.ecr.aws/aws-controllers-k8s', version='1.9.2', namespace='ack-system', values=['manifests/ack-ec2-values.yaml'])
    k8s_resource('ack-ec2-controller-ec2-chart', resource_deps=['localstack'], labels=['dependencies', '2-ack'])

    KRO_DEPS = ['ack-ec2-controller-ec2-chart']
else:
    KRO_DEPS = []

helm_remote('kro', repo_url='oci://registry.k8s.io/kro/charts', version='0.8.4', namespace='kro-system', create_namespace=True)
k8s_resource('kro', resource_deps=KRO_DEPS, labels=['dependencies', '3-kro'])

helm_remote('gha-runner-scale-set-controller', release_name='arc', repo_url='oci://ghcr.io/actions/actions-runner-controller-charts', version='0.13.1', namespace=CONTROLLER_NS, create_namespace=True)
k8s_resource('arc-gha-rs-controller', resource_deps=['kro'], labels=['dependencies', '4-arc'])

# Manual triggers
local_resource('test-unit', cmd=with_mise('go test -v ./...'), auto_init=False, trigger_mode=TRIGGER_MODE_MANUAL, labels=['tests'])
local_resource('test-ec2-integration', cmd=with_mise('kubectl kuttl test --config test/e2e/ec2-integration/kuttl/kuttl-test.yaml'), auto_init=False, trigger_mode=TRIGGER_MODE_MANUAL, resource_deps=['ack-ec2-controller-ec2-chart', 'kro-actions-runner'] if ENABLE_LOCALSTACK else ['kro', 'kro-actions-runner'], labels=['tests'])
local_resource('lint', cmd=with_mise('golangci-lint run --fast'), auto_init=False, trigger_mode=TRIGGER_MODE_MANUAL, deps=['cmd/', 'internal/'], labels=['dev-tools'])
local_resource('cluster-status', cmd=with_mise('kubectl get pods -A'), auto_init=False, trigger_mode=TRIGGER_MODE_MANUAL, labels=['dev-tools'])

if ENABLE_LOCALSTACK:
    local_resource('view-ec2-instances', cmd=with_mise('kubectl exec -n localstack deploy/localstack -- awslocal ec2 describe-instances --query "Reservations[*].Instances[*].[InstanceId,State.Name,Tags[?Key==`Name`].Value|[0]]" --output table'), auto_init=False, trigger_mode=TRIGGER_MODE_MANUAL, labels=['dev-tools'])

update_settings(max_parallel_updates=3, k8s_upsert_timeout_secs=300)
