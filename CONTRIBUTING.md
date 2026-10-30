# Contributing to musex

Use English for commit subjects and bodies, pull request titles and descriptions,
pull request comments, and comments in project files. Material under `docs/` is
exempt from this requirement.

Use the Go version in `go.mod`, golangci-lint **v2.15.0** (also recorded in
`.golangci-lint-version`), and govulncheck **v1.9.0** to match CI.

Install golangci-lint using its [official release instructions](https://golangci-lint.run/docs/welcome/install/local/).
For a project-local installation, place the binary in `.cache/bin/` and install
govulncheck there as well:

```sh
mkdir -p .cache/bin
GOBIN="$PWD/.cache/bin" go install golang.org/x/vuln/cmd/govulncheck@v1.9.0
export PATH="$PWD/.cache/bin:$PATH"
```

Format the code and run the checks from the repository root:

```sh
golangci-lint fmt ./...
golangci-lint run ./...
go test -race -count=1 -covermode=atomic -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
govulncheck ./...
```

Imports are grouped as standard library, `github.com/go-muse/musex` packages,
then third-party packages, separated by blank lines and sorted alphabetically
within each group. Both formatting and linting use `.golangci.yml`.

GitHub Actions runs lint, tests, and vulnerability checks independently on pull
requests and pushes to `main`. Coverage is shown in the workflow summary and
saved as an artifact containing text, HTML, and the raw profile. There is no
minimum coverage threshold.

Until the first Go package is added, CI validates the lint configuration and
explicitly skips package analysis, tests, and coverage. Once packages exist,
these checks run automatically. Docker is not required.
