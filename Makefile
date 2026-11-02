# Development checks for musex.
#
# Run targets from the repository root; 'make help' lists them. Requires Go,
# GNU make 3.81 or later, python3 (for 'cover' and 'check') and common POSIX
# shell utilities. macOS and Linux are supported; Windows is not.
#
# CI (.github/workflows/ci.yml) does not call make: it repeats the same check
# commands. Change a check command or flag here and in ci.yml in the same
# change. Tool versions are kept only in .golangci-lint-version and
# .govulncheck-version, which both make and CI read.

SHELL := /bin/sh

# Always use the local Go toolchain, as CI does (GOTOOLCHAIN: local). The
# override ignores values from the environment and the command line.
override GOTOOLCHAIN := local
export GOTOOLCHAIN

# Tools directory inside the repository. The override ignores values from the
# environment and the command line, so 'make tools' never writes elsewhere.
override TOOLS_BIN := $(CURDIR)/.cache/bin
# golangci-lint to use; override with 'make <target> GOLANGCI_LINT=<path>'.
GOLANGCI_LINT = $(TOOLS_BIN)/golangci-lint
GOVULNCHECK = $(TOOLS_BIN)/govulncheck

GOLANGCI_LINT_VERSION_FILE = .golangci-lint-version
GOVULNCHECK_VERSION_FILE = .govulncheck-version

GOLANGCI_LINT_PKG = github.com/golangci/golangci-lint/v2/cmd/golangci-lint
GOVULNCHECK_PKG = golang.org/x/vuln/cmd/govulncheck

ifeq ($(origin GOLANGCI_LINT),file)
golangci_lint_hint = Run 'make tools'.
else
golangci_lint_hint = Set GOLANGCI_LINT to golangci-lint $$v, or omit it and run 'make tools'.
endif

# Shell fragments for the tool checks. Each expands inside one recipe line and
# exits the recipe shell with status 1 on failure. Version files are read in
# recipes, not while make parses this file. Do not use the hash character in
# these definitions: make would treat it as the start of a comment.

# $(call read_version,FILE): set v to the version in FILE (vMAJOR.MINOR.PATCH)
# or fail. The case rejects empty values, extra lines, spaces and CR.
read_version = v=$$(cat "$(1)" 2>/dev/null); \
	case "$$v" in ''|*[!v0-9.]*) v=;; esac; \
	if [ -z "$$v" ] || ! printf '%s\n' "$$v" | grep -Eqx 'v[0-9]+\.[0-9]+\.[0-9]+'; then \
		printf '%s\n' "$(1) is missing or invalid (expected vMAJOR.MINOR.PATCH)." >&2; \
		exit 1; \
	fi

# $(call golangci_lint_version,PATH): set found to the version of the
# golangci-lint at PATH: vX.Y.Z, 'missing' or 'unknown'.
golangci_lint_version = if [ -f "$(1)" ] && [ -x "$(1)" ]; then \
		found=$$("$(1)" version --short 2>/dev/null) || found=; \
		if [ -n "$$found" ]; then found="v$$found"; else found=unknown; fi; \
	else \
		found=missing; \
	fi

# $(call govulncheck_version,PATH): set found to the golang.org/x/vuln module
# version recorded in the Go binary at PATH: vX.Y.Z, 'missing' or 'unknown'.
# Reads build information only; govulncheck itself is not run.
govulncheck_version = if [ -f "$(1)" ] && [ -x "$(1)" ]; then \
		found=$$(go version -m "$(1)" 2>/dev/null | awk '$$1=="mod" && $$2=="golang.org/x/vuln" {print $$3}'); \
		[ -n "$$found" ] || found=unknown; \
	else \
		found=missing; \
	fi

# $(call require_version,NAME,PATH,HINT): fail unless found equals v.
require_version = if [ "$$found" = missing ]; then \
		printf '%s\n' "$(1) not found at $(2) (required $$v). $(3)" >&2; \
		exit 1; \
	elif [ "$$found" != "$$v" ]; then \
		printf '%s\n' "$(1) at $(2) is version $$found, required $$v. $(3)" >&2; \
		exit 1; \
	fi

.NOTPARALLEL:

.PHONY: check help fmt lint test cover vuln tidy tools clean \
	require-golangci-lint require-govulncheck mod-check roadmap-badge

# The first rule is the default goal: 'make' runs 'make check'. Both tools
# are checked first, so a missing tool is reported before any check runs.
# Use 'make -k check' to run every check and see all failures.
check: require-golangci-lint require-govulncheck lint mod-check test roadmap-badge vuln ## Run lint, module tidiness, test, roadmap badge and vuln checks; changes no source files (default)
	@printf '%s\n' 'All checks passed.'

help: ## List the targets
	@printf '%s\n\n' 'Usage: make [target] ...'
	@awk -F':.*## ' '/^[a-z][a-z-]*:.*## / {printf "  %-8s %s\n", $$1, $$2}' "$(firstword $(MAKEFILE_LIST))"

fmt: require-golangci-lint ## Fix formatting and import order in place (golangci-lint fmt)
	"$(GOLANGCI_LINT)" fmt ./...

# Keep in sync with CI: lint / 'Validate lint configuration' and
# lint / 'Check code, formatting and import order'.
lint: require-golangci-lint ## Verify the linter config, then lint; unformatted code fails
	"$(GOLANGCI_LINT)" config verify
	"$(GOLANGCI_LINT)" run ./...

# Keep in sync with CI: tests / 'Test with race detection and coverage'.
test: ## Run the tests with the race detector; writes coverage.out
	go test -race -count=1 -covermode=atomic -coverprofile=coverage.out ./...

# Keep in sync with CI: tests / 'Generate coverage badge and reports'.
cover: test roadmap-badge ## Run the tests, then build the coverage report and roadmap badge in coverage/
	python3 scripts/coverage.py --profile coverage.out

# Keep in sync with CI: vulnerabilities / 'Check known vulnerabilities'.
vuln: require-govulncheck ## Check for known vulnerabilities with govulncheck (network)
	"$(GOVULNCHECK)" ./...

tidy: ## Run 'go mod tidy'; updates go.mod and go.sum
	go mod tidy

# Install each tool into TOOLS_BIN unless the required version is already
# there, then check the result. A file of another version is removed first,
# because go install refuses to overwrite a file that is not a Go binary.
# Always writes to TOOLS_BIN and ignores GOLANGCI_LINT. GOBIN is set
# explicitly, so nothing is installed into GOPATH/bin or a GOBIN from the
# environment.
# Keep the govulncheck installation in sync with CI: vulnerabilities /
# 'Install govulncheck'.
tools: ## Install golangci-lint and govulncheck of the required versions into .cache/bin (network)
	@$(call read_version,$(GOLANGCI_LINT_VERSION_FILE)); \
	$(call golangci_lint_version,$(TOOLS_BIN)/golangci-lint); \
	if [ "$$found" = "$$v" ]; then \
		printf '%s\n' "golangci-lint $$v is already installed in $(TOOLS_BIN)."; \
	else \
		printf '%s\n' "GOBIN=\"$(TOOLS_BIN)\" go install $(GOLANGCI_LINT_PKG)@$$v"; \
		mkdir -p "$(TOOLS_BIN)" && rm -f "$(TOOLS_BIN)/golangci-lint" && \
		GOBIN="$(TOOLS_BIN)" go install "$(GOLANGCI_LINT_PKG)@$$v" || exit 1; \
		$(call golangci_lint_version,$(TOOLS_BIN)/golangci-lint); \
		$(call require_version,golangci-lint,$(TOOLS_BIN)/golangci-lint,Installation failed.); \
	fi
	@$(call read_version,$(GOVULNCHECK_VERSION_FILE)); \
	$(call govulncheck_version,$(TOOLS_BIN)/govulncheck); \
	if [ "$$found" = "$$v" ]; then \
		printf '%s\n' "govulncheck $$v is already installed in $(TOOLS_BIN)."; \
	else \
		printf '%s\n' "GOBIN=\"$(TOOLS_BIN)\" go install $(GOVULNCHECK_PKG)@$$v"; \
		mkdir -p "$(TOOLS_BIN)" && rm -f "$(TOOLS_BIN)/govulncheck" && \
		GOBIN="$(TOOLS_BIN)" go install "$(GOVULNCHECK_PKG)@$$v" || exit 1; \
		$(call govulncheck_version,$(TOOLS_BIN)/govulncheck); \
		$(call require_version,govulncheck,$(TOOLS_BIN)/govulncheck,Installation failed.); \
	fi

clean: ## Remove coverage.out and coverage/
	rm -rf coverage.out coverage

# Internal targets: not listed in 'make help'.

# Check that golangci-lint at GOLANGCI_LINT has the version in
# .golangci-lint-version. Never falls back to golangci-lint from PATH.
require-golangci-lint:
	@$(call read_version,$(GOLANGCI_LINT_VERSION_FILE)); \
	$(call golangci_lint_version,$(GOLANGCI_LINT)); \
	$(call require_version,golangci-lint,$(GOLANGCI_LINT),$(golangci_lint_hint))

# Check that govulncheck at GOVULNCHECK has the version in .govulncheck-version.
require-govulncheck:
	@$(call read_version,$(GOVULNCHECK_VERSION_FILE)); \
	$(call govulncheck_version,$(GOVULNCHECK)); \
	$(call require_version,govulncheck,$(GOVULNCHECK),Run 'make tools'.)

# Roadmap badge, part of 'cover' and 'check': counts the checkboxes in
# ROADMAP.md and writes coverage/roadmap.svg; fails on an unexpected layout.
# Keep in sync with CI: tests / 'Generate roadmap badge'.
roadmap-badge:
	python3 scripts/roadmap.py

# Module tidiness check, part of 'check'; changes no files.
# Keep in sync with CI: lint / 'Check module tidiness'.
mod-check:
	go mod tidy -diff || { printf '%s\n' "go.mod/go.sum are not tidy (or 'go mod tidy -diff' failed, see above). Run 'make tidy'." >&2; exit 1; }
