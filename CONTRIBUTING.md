# Contributing to dtasks

Thanks for taking the time to contribute.

## Getting started

```bash
git clone https://github.com/danielmrdev/dtasks-cli
cd dtasks
go mod tidy
go build ./...
go test ./...
```

## Workflow

1. [Open an issue](https://github.com/danielmrdev/dtasks-cli/issues) before starting non-trivial work so we can discuss the approach.
2. Fork the repository and create a branch from `main`:
   ```bash
   git checkout -b fix/short-description
   git checkout -b feat/short-description
   ```
3. Make your changes. Keep commits focused and atomic.
4. Run the checks locally before pushing:
   ```bash
   go vet ./...
   go test ./...
   ```
5. Open a pull request against `main`. Fill in the PR template.

## Code style

- Standard Go formatting — run `gofmt` or `goimports` before committing.
- Keep functions small and focused. Prefer clarity over cleverness.
- All new behaviour should have a test in the same package (`_test.go` next to the file).
- No CGO. The project must build with `CGO_ENABLED=0` on all target platforms.

## Project layout

```
cmd/         CLI commands (cobra). No business logic here.
internal/
  config/    Config file loading and first-run wizard.
  db/        SQLite open + PRAGMAs + migration.
  models/    Plain structs. No methods, no dependencies.
  repo/      Database queries. Pure functions, no global state.
  output/    Table and JSON rendering.
```

New packages under `internal/` are welcome for new features. Avoid adding dependencies to the module root (`cmd/`) beyond cobra.

## Reporting bugs

Use the [bug report template](https://github.com/danielmrdev/dtasks-cli/issues/new?template=bug_report.yml). Include your platform, dtasks version (`dtasks --version`), and the exact command that failed.

## Suggesting features

Use the [feature request template](https://github.com/danielmrdev/dtasks-cli/issues/new?template=feature_request.yml).

## Versioning

This project follows [Semantic Versioning](https://semver.org). Changes are documented in [CHANGELOG.md](CHANGELOG.md).

To publish a release:
```bash
make release TAG=v1.2.3
```

This creates a git tag and pushes it, which triggers the GitHub Actions release workflow.
