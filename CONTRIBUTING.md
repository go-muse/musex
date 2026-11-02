# Contributing to musex

Use English for commit subjects and bodies, pull request titles and descriptions,
pull request comments, and comments in project files. Material under `docs/` is
exempt from this requirement.

Use the Go version in `go.mod`, golangci-lint **v2.15.0** (also recorded in
`.golangci-lint-version`), and govulncheck **v1.9.0** (also recorded in
`.govulncheck-version`) to match CI. `make tools` installs both tools; see
[Make targets](#make-targets).

Install golangci-lint using its [official release instructions](https://golangci-lint.run/docs/welcome/install/local/).
For a project-local installation, place the binary in `.cache/bin/` and install
govulncheck there as well:

```sh
mkdir -p .cache/bin
GOBIN="$PWD/.cache/bin" go install "golang.org/x/vuln/cmd/govulncheck@$(cat .govulncheck-version)"
export PATH="$PWD/.cache/bin:$PATH"
```

Format the code and run the checks from the repository root:

```sh
golangci-lint fmt ./...
golangci-lint run ./...
go mod tidy -diff
go test -race -count=1 -covermode=atomic -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
govulncheck ./...
```

Imports are grouped as standard library, `github.com/go-muse/musex` packages,
then third-party packages, separated by blank lines and sorted alphabetically
within each group. Both formatting and linting use `.golangci.yml`.
The depguard rule of each package in `.golangci.yml` sets which musex packages
it may import; when the package plan changes, update the rule in the same change.

GitHub Actions runs lint, tests, and vulnerability checks independently on pull
requests and pushes to `main`. The lint check also verifies that `go.mod` and
`go.sum` are tidy with `go mod tidy -diff`. Coverage is shown in the workflow
summary and saved as an artifact containing text, HTML, the raw profile, and
an SVG badge.
The tests check also counts how many package checkboxes in `ROADMAP.md` are
checked, shows the result in the workflow summary, and adds a roadmap progress
badge, `roadmap.svg` (for example `1/30 packages`), to the same artifact.
Checkboxes in bullet and numbered lists are counted; those in fenced code blocks
and HTML comments are ignored. If there are no checkboxes, a checkbox is
malformed or inside a block quote, or a code block or comment is not closed, the
tests check fails instead of publishing a misleading badge.
After all checks pass on `main`, both badges and the HTML report are published to
[GitHub Pages](https://go-muse.github.io/musex/). Pull requests generate artifacts
but never deploy. There is no minimum coverage threshold or external coverage
service. The coverage badge uses a neutral chart icon and shows `n/a` when there
are no executable statements; otherwise it shows Go's total statement coverage.

To generate the same report and badges locally after running the tests above
(`make cover` runs the tests and both commands):

```sh
python3 scripts/coverage.py --profile coverage.out
python3 scripts/roadmap.py
```

Open `coverage/index.html` to inspect the report and `coverage/roadmap.svg` to
see the roadmap badge. For a repository without Go packages, use
`python3 scripts/coverage.py --no-packages`. Generated files stay in the ignored
`coverage/` directory. GitHub Pages must use GitHub Actions as its publishing
source in the repository settings.

Until the first Go package is added, CI validates the lint configuration and
module tidiness and explicitly skips package analysis, tests, and coverage.
Once packages exist, these checks run automatically. Docker is not required.

## Make targets

The `Makefile` runs the same checks as CI. It needs Go, GNU make 3.81 or later,
and `python3` for `make cover` and `make check`, and works on macOS and Linux
but not Windows. Run it from the repository root:

| Target | What it does |
|---|---|
| `make help` | Lists the targets. |
| `make tools` | Installs golangci-lint and govulncheck of the required versions into `.cache/bin/`. |
| `make fmt` | Fixes formatting and import order in place: `golangci-lint fmt ./...`. |
| `make lint` | Runs `golangci-lint config verify`, then `golangci-lint run ./...`. Unformatted code fails. |
| `make test` | Runs `go test -race -count=1 -covermode=atomic -coverprofile=coverage.out ./...`. |
| `make cover` | Runs the tests and `python3 scripts/roadmap.py`, then `python3 scripts/coverage.py --profile coverage.out`. |
| `make vuln` | Runs `govulncheck ./...`. |
| `make tidy` | Runs `go mod tidy`, which updates `go.mod` and `go.sum`. |
| `make check` | Runs `lint`, the module tidiness check (`go mod tidy -diff`), `test`, the roadmap badge (`python3 scripts/roadmap.py`), and `vuln`. Changes no source files. |
| `make clean` | Removes `coverage.out` and `coverage/`. |

- `make` without a target runs `make check`, which stops at the first failure.
  Use `make -k check` to run every check and see all failures.
- `fmt`, `lint`, and `vuln` use only the tools in `.cache/bin/`, never a
  `golangci-lint` or `govulncheck` from `PATH`, and check the tool version
  first. If a tool is missing or has another version, they stop and ask you to
  run `make tools`.
- To use another golangci-lint binary, pass its path, for example
  `make lint GOLANGCI_LINT=/path/to/golangci-lint`. It must have the required
  version.
- `make tools`, `make vuln`, and `make check` need network access: the Go module
  proxy to install the tools and the Go vulnerability database for govulncheck.
- Go commands run with `GOTOOLCHAIN=local`, as in CI.

CI does not call make. When you change a check command or flag, change it in
both the `Makefile` and `.github/workflows/ci.yml` in the same pull request.
Tool versions are kept only in `.golangci-lint-version` and
`.govulncheck-version`, which both read.
