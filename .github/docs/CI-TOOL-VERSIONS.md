# CI and Tool Version Management

This document explains the project's approach to tool version management and the relationship between local development (mise) and CI workflows.

## Overview

The project uses a **hybrid approach** for tool management:

- **Local Development**: Uses [mise](https://mise.jdx.dev) (`mise.toml`) for tool version management and task running
- **CI Workflows**: Uses GitHub Actions directly with explicit tool versions in workflow files

## Why Not Use mise in CI?

We deliberately keep CI independent of mise for several reasons:

1. **Performance**: GitHub Actions has optimized caching for common tools (Go, Node.js, etc.)
2. **Reliability**: No dependency on external package manager in CI
3. **Simplicity**: CI workflows are self-contained and easier to debug
4. **Existing Investment**: Current CI setup is stable and performant

## Tool Version Sources

Tool versions are defined in multiple files:

| Tool | mise.toml | CI Workflows | go.mod | Dockerfile | .pre-commit-config.yaml |
|------|-----------|--------------|--------|------------|-------------------------|
| Go | ✅ 1.25.0 | ✅ | ✅ 1.25 | ✅ 1.25 | - |
| Node.js | ✅ 22 | ✅ | - | - | - |
| Python | ✅ 3.12 | ✅ | - | - | - |
| Ruby | ✅ 3.3 | - | - | - | - |
| golangci-lint | ✅ latest | ✅ v1.x | - | - | - |
| yamlfmt | ✅ 0.20.0 | - | - | - | ✅ v0.20.0 |
| shellcheck | ✅ 0.11.0.1 | ✅ | - | - | ✅ v0.11.0.1 |
| shfmt | ✅ latest | - | - | - | - |
| prettier | ✅ latest | - | - | - | - |

## Version Sync Strategy

### Single Source of Truth

**`mise.toml`** is the **primary source of truth** for tool versions used in local development.

### Syncing with CI

When updating tool versions, you must manually sync between `mise.toml` and CI workflows.

**Example: Updating Go version**

1. Update `mise.toml`:
   ```toml
   [tools]
   go = "1.26.0"
   ```

2. Update `.github/workflows/*.yml`:
   ```yaml
   - uses: actions/setup-go@v5
     with:
       go-version: '1.26.0'
   ```

3. Update `go.mod`:
   ```go
   go 1.26
   ```

4. Update `Dockerfile`:
   ```dockerfile
   FROM golang:1.26-alpine
   ```

5. Update `.golangci.yml` (if needed):
   ```yaml
   run:
     go: '1.26'
   ```

6. Test locally:
   ```bash
   mise install
   mise run check
   ```

7. Commit all changes together.

## Version Update Checklist

Use this checklist when updating tool versions to ensure consistency across all files.

### Updating Go Version

- [ ] `mise.toml` - Update `go = "X.Y.Z"`
- [ ] `.github/workflows/*.yml` - Update `go-version: 'X.Y'`
- [ ] `go.mod` - Update `go X.Y`
- [ ] `Dockerfile` - Update `FROM golang:X.Y-alpine`
- [ ] `.golangci.yml` - Update `run.go: 'X.Y'` (if exists)
- [ ] Test: `mise install && mise run test`
- [ ] Test: Build Docker image

### Updating Node.js Version

- [ ] `mise.toml` - Update `node = "X"`
- [ ] `.github/workflows/*.yml` - Update `node-version: 'X'`
- [ ] Test: `mise install && mise run fmt`

### Updating Python Version

- [ ] `mise.toml` - Update `python = "X.Y"`
- [ ] `.github/workflows/*.yml` - Update `python-version: 'X.Y'`
- [ ] Test: `mise install && pre-commit run --all-files`

### Updating Linters/Formatters

- [ ] `mise.toml` - Update tool version
- [ ] `.github/workflows/*.yml` - Update corresponding action/version
- [ ] `.pre-commit-config.yaml` - Update `rev:` field (if applicable)
- [ ] Test: `mise install && mise run lint`

## Detecting Version Drift

### Manual Verification

Compare versions across files:

```bash
# Check Go version
grep -r "go.*1.25" mise.toml go.mod Dockerfile .github/workflows/

# Check yamlfmt version
grep -r "yamlfmt.*0.20" mise.toml .pre-commit-config.yaml

# Check shellcheck version
grep -r "shellcheck.*0.11" mise.toml .pre-commit-config.yaml
```

### Automated Verification (Future)

Consider adding a CI check that validates version consistency:

```yaml
# .github/workflows/version-check.yml
name: Version Consistency Check
on: [pull_request]
jobs:
  check-versions:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Verify tool versions match
        run: |
          # Script to parse mise.toml and compare with CI workflows
          ./scripts/check-version-sync.sh
```

## Best Practices

### When Adding New Tools

1. **Add to `mise.toml` first** with a specific version
2. **Add to CI workflows** if the tool is used in CI
3. **Add to `.pre-commit-config.yaml`** if it's a pre-commit hook
4. **Document the version** in this file's table
5. **Test thoroughly** in both local dev and CI

### When Deprecating Tools

1. **Remove from `mise.toml`**
2. **Remove from CI workflows**
3. **Remove from `.pre-commit-config.yaml`**
4. **Update this document**
5. **Update `CONTRIBUTING.md`** if workflow changes

### Version Pinning Philosophy

- **Language Runtimes** (Go, Node.js, Python): Pin to minor version (e.g., `1.25.0`)
- **Linters/Formatters**: Pin to exact version for consistency (e.g., `0.20.0`)
- **Optional Tools**: Can use `latest` if breaking changes are rare

## FAQ

### Why don't we use mise in CI?

GitHub Actions provides optimized caching and setup actions for common tools. Using mise in CI would add latency and complexity without significant benefits. The hybrid approach gives us the best of both worlds: consistency in local development and performance in CI.

### How do I know if versions are out of sync?

Run `mise run check` locally and ensure all CI checks pass. If CI passes but local checks fail (or vice versa), versions may be out of sync. Use the manual verification commands above to check.

### What if CI and local development have different versions?

This is undesirable and should be fixed immediately. Version drift can lead to:
- Tests passing locally but failing in CI (or vice versa)
- Inconsistent linting results
- Unexpected behavior in production builds

Follow the [Version Update Checklist](#version-update-checklist) to sync versions.

### Can I use different tool versions locally?

While mise allows per-directory tool versions, we recommend sticking to the versions in `mise.toml` for consistency with CI and other developers. If you need a different version for experimentation, use a separate directory or branch.

### Who is responsible for keeping versions in sync?

Everyone contributing to the project. When you update a tool version, you must update all relevant files in the same commit. Reviewers should verify version consistency during code review.

## Resources

- [mise documentation](https://mise.jdx.dev)
- [GitHub Actions setup-go](https://github.com/actions/setup-go)
- [GitHub Actions setup-node](https://github.com/actions/setup-node)
- [GitHub Actions setup-python](https://github.com/actions/setup-python)
- [golangci-lint configuration](https://golangci-lint.run/usage/configuration/)
