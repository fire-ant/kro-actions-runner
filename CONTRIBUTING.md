# Contributing to KRO Actions Runner

Thank you for your interest in contributing to KRO Actions Runner! This document provides guidelines for development workflow, testing, and submitting contributions.

## Development Environment Setup

### Prerequisites

- [mise](https://mise.jdx.dev) - Tool version manager and task runner
- [Docker](https://www.docker.com/) or [Podman](https://podman.io/) - For building container images
- A Kubernetes cluster for testing (optional but recommended)

### Initial Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/fire-ant/kro-actions-runner.git
   cd kro-actions-runner
   ```

2. **Install mise** (if not already installed):
   ```bash
   # macOS
   brew install mise

   # Linux/macOS
   curl https://mise.run | sh
   ```

3. **Install all development tools:**
   ```bash
   mise install
   ```
   This installs all required tools at the correct versions (Go, Node.js, Python, Ruby, linters, formatters, etc.).

4. **Set up the development environment:**
   ```bash
   mise run setup
   ```
   This installs Node.js dependencies and pre-commit hooks.

5. **Verify your setup:**
   ```bash
   mise run check
   ```
   This runs formatting, linting, and tests to ensure everything is working.

## Development Workflow

### Running Tasks

We use `mise` for all development tasks. Here are the most common commands:

```bash
# Run tests
mise run test

# Format code
mise run fmt

# Run linters
mise run lint

# Build the binary
mise run build

# Run all checks (format, lint, test)
mise run check
```

**See all available tasks:**
```bash
mise tasks
```

### Making Changes

1. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the [Go Code Style](#go-code-style) guidelines.

3. **Format your code** before committing:
   ```bash
   mise run fmt
   ```

4. **Run linters** to catch issues:
   ```bash
   mise run lint
   ```

5. **Run tests** to ensure nothing breaks:
   ```bash
   mise run test
   ```

6. **Commit your changes** with a descriptive message:
   ```bash
   git add .
   git commit -m "Add feature: description of your changes"
   ```

   Pre-commit hooks will automatically format and lint your code.

7. **Push your branch** and create a pull request:
   ```bash
   git push origin feature/your-feature-name
   ```

### Pre-commit Hooks

Pre-commit hooks automatically run formatters and linters before each commit. They are installed by `mise run setup`.

**Hooks include:**
- Trailing whitespace removal
- YAML syntax checking
- Shell script linting (shellcheck, bashate)
- Markdown linting
- YAML formatting (yamlfmt)

**Manual hook installation:**
```bash
mise run install-hooks
```

## Go Code Style

### General Guidelines

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting (automatically applied by `mise run fmt`)
- Write clear, descriptive variable and function names
- Add comments for exported functions and types
- Keep functions small and focused on a single responsibility

### Linting

We use `golangci-lint` with the configuration in `.golangci.yml`. Run it with:

```bash
mise run lint
```

Common issues to avoid:
- Unused variables or imports
- Error returns that aren't checked
- Unnecessary type conversions
- Overly complex functions

## Testing

### Unit Tests

Write unit tests for new functionality in `*_test.go` files:

```go
func TestYourFunction(t *testing.T) {
    // Arrange
    input := "test"

    // Act
    result, err := YourFunction(input)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

**Run all tests:**
```bash
mise run test
```

**Run tests for a specific package:**
```bash
go test -v ./pkg/runner
```

**Run tests with coverage:**
```bash
go test -v -cover ./...
```

### Integration Tests

For integration tests that require Kubernetes:

1. Set up a local cluster (kind, minikube, or k3s)
2. Install KRO: `kubectl apply -f https://github.com/kubernetes-sigs/kro/releases/latest/download/kro.yaml`
3. Run tests with appropriate kubeconfig

## Building and Testing the Container Image

### Build the image

```bash
docker build -t kro-actions-runner:dev .
```

### Test the image locally

```bash
# Create a test secret
kubectl create secret generic test-jit --from-literal=.jitconfig='{"test":"config"}'

# Deploy a test RGD
kubectl apply -f examples/pod-runner-rgd.yaml

# Run the container
kubectl run test-runner \
  --image=kro-actions-runner:dev \
  --env="ACTIONS_RUNNER_SCALE_SET_NAME=default" \
  --env="RUNNER_NAME=test-runner-1" \
  --env="ACTIONS_RUNNER_INPUT_JITCONFIG=$(kubectl get secret test-jit -o jsonpath='{.data.\.jitconfig}' | base64 -d)" \
  --restart=Never

# Check logs
kubectl logs test-runner -f
```

## Pull Request Guidelines

### Before Submitting

- [ ] Code is formatted (`mise run fmt`)
- [ ] All linters pass (`mise run lint`)
- [ ] All tests pass (`mise run test`)
- [ ] New functionality has tests
- [ ] Documentation is updated (README.md, comments, etc.)
- [ ] Commit messages are clear and descriptive

### PR Description

Include in your PR description:
- **What**: Brief summary of changes
- **Why**: Motivation and context
- **How**: Technical approach (if complex)
- **Testing**: How you tested the changes
- **Related Issues**: Link to related issues or discussions

### Code Review Process

1. Automated CI checks must pass (linting, tests, builds)
2. At least one maintainer review is required
3. Address review feedback by pushing new commits
4. Once approved, a maintainer will merge your PR

## CI/CD

### GitHub Actions Workflows

The project uses GitHub Actions for CI/CD:

- **Linting**: Runs golangci-lint, yamllint, shellcheck
- **Testing**: Runs `go test -v ./...`
- **Building**: Builds the Docker image
- **Publishing**: Publishes to GHCR on releases

**Important:** CI workflows do NOT use mise (they use GitHub Actions directly for optimized caching). However, tool versions in CI should match those in `mise.toml`. See [.github/docs/CI-TOOL-VERSIONS.md](.github/docs/CI-TOOL-VERSIONS.md) for details.

## Updating Tool Versions

### Mise-Managed Tools

To update tool versions (Go, Node.js, linters, etc.):

1. **Update `mise.toml`:**
   ```toml
   [tools]
   go = "1.26.0"  # Update version
   ```

2. **Update corresponding CI workflow files** in `.github/workflows/*.yml`

3. **Update other version references:**
   - `go.mod` (Go version)
   - `.golangci.yml` (if Go version specific)
   - `Dockerfile` (Go version)
   - `.pre-commit-config.yaml` (tool versions)

4. **Test the changes:**
   ```bash
   mise install
   mise run check
   ```

See [.github/docs/CI-TOOL-VERSIONS.md](.github/docs/CI-TOOL-VERSIONS.md) for a complete version sync checklist.

## Getting Help

- **Questions**: Open a [GitHub Discussion](https://github.com/fire-ant/kro-actions-runner/discussions)
- **Bugs**: Open a [GitHub Issue](https://github.com/fire-ant/kro-actions-runner/issues)
- **Documentation**: See [README.md](README.md) and [.github/docs/](.github/docs/)

## License

By contributing to KRO Actions Runner, you agree that your contributions will be licensed under the Apache License 2.0.
