# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2025
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

# This Makefile is a backward-compatible wrapper around mise
# All tasks are now defined in mise.toml
#
# DEPRECATION NOTICE: Consider using `mise run <task>` directly
# - `make test` -> `mise run test`
# - `make fmt` -> `mise run fmt`
# - `make lint` -> `mise run lint`
#
# Setup: mise install && mise run setup
# See available tasks: mise tasks

# Check if mise is installed
MISE := $(shell command -v mise 2> /dev/null)

ifndef MISE
$(error mise is not installed. Install it from https://mise.jdx.dev/getting-started.html)
endif

.PHONY: test
test:
	@mise run test

.PHONY: lint
lint:
	@mise run lint

.PHONY: fmt
fmt:
	@mise run fmt

.PHONY: build
build:
	@mise run build

.PHONY: clean
clean:
	@mise run clean

.PHONY: check
check:
	@mise run check

.PHONY: setup
setup:
	@mise run setup

.PHONY: help
help:
	@echo "Makefile wrapper for mise tasks"
	@echo ""
	@echo "Available targets:"
	@echo "  make test    - Run Go tests"
	@echo "  make fmt     - Format all code files"
	@echo "  make lint    - Run all linters"
	@echo "  make build   - Build the kar binary"
	@echo "  make clean   - Clean build artifacts"
	@echo "  make check   - Run all checks (fmt, lint, test)"
	@echo "  make setup   - Set up development environment"
	@echo ""
	@echo "For more details, see: mise tasks"
	@echo ""
	@echo "DEPRECATION NOTICE: Consider using 'mise run <task>' directly"
