# AGENTS.md

Guidance for AI coding agents working in this repository. Human contributors
should read [CONTRIBUTING.md](CONTRIBUTING.md); this file is the machine-facing
counterpart.

## What this project is

`gtoc` is a single-binary CLI, written in Go, that generates and maintains a
table of contents in Markdown files. It reads the headings of a file, builds a
hierarchical index with GitHub-compatible anchors, and writes it back between
HTML markers — idempotently.

Module path: `github.com/lpsm-dev/gtoc`.

## Layout

| Path | Responsibility |
| ------------------------ | ------------------------------------------------------------- |
| `main.go` | Entry point; delegates to `cmd.Execute`. |
| `cmd/` | Cobra commands: `generate`, `analyze`, `upgrade`, `version`. |
| `internal/generator/` | Heading extraction, anchor slugging, TOC assembly and file update. Pure, well-tested core. |
| `internal/logger/` | Thin wrapper over `charm.land/log/v2`. |
| `.github/workflows/` | CI (`ci.yaml`), release (`goreleaser.yaml`, `_release.yaml`). |

The `generate` command enumerates headings into a TOC; the `analyze` command
adds README best-practice markers and a "back to top" link at the end of each
top-level (`#`) section. The `upgrade` command self-updates from GitHub
releases.

## Build, test, lint

The project uses [Task](https://taskfile.dev) and [Devbox](https://www.jetify.com/devbox),
but plain `go` works everywhere. Prefer the raw commands in automation:

```bash
go build ./...            # or: task go:build
go test ./...             # or: task go:test
go vet ./...              # or: task go:vet
gofmt -l cmd internal     # must print nothing
golangci-lint run ./...   # or: task go:lint
```

`task go:all` runs fmt, vet, lint, test and build in sequence.

## Non-negotiable rules

- **Cyclomatic complexity ≤ 10 per function.** Enforced by `golangci-lint`
  (`gocyclo`, `min-complexity: 10` in `.golangci.yml`). New code must comply;
  do not raise the threshold.
- **Functions ≤ 100 lines**, prefer small single-purpose helpers.
- **All code, comments and identifiers in English.**
- **`gofmt`-clean, `go vet`-clean, zero linter warnings.**
- **Tests must stay green** (`go test ./...`) and new behavior needs new tests.
  The `internal/generator` package is the core — keep its coverage high.
- **No unnecessary dependencies.** The core uses only the standard library.
- **Anchors follow GitHub's github-slugger algorithm**: preserve Unicode
  letters/accents, do not collapse consecutive hyphens, and suffix duplicate
  slugs with `-1`, `-2`, … Changing this breaks existing anchors — add a test
  first.
- **Go module major bumps (v2+) are not routine.** Charm modules relocate their
  import path on major versions (e.g. `charm.land/glamour/v2`), so a bare
  `go.mod` bump breaks the build. Update the import sites and verify a full
  build before merging.

## Commits and releases

- **Conventional Commits**: `type(scope): description`, lowercase, imperative,
  no trailing period. Types: `feat`, `fix`, `refactor`, `docs`, `test`,
  `chore`, `ci`, `perf`. Breaking changes use `feat!:` or a `BREAKING CHANGE:`
  footer.
- Keep commits small and single-concern.
- Releases are driven by `semantic-release` and GoReleaser off `main`; never
  hand-edit the changelog.

## Definition of done

Before proposing a change as complete, confirm locally:

1. `go build ./...` succeeds.
2. `go test ./...` passes.
3. `go vet ./...` and `gofmt -l cmd internal` are clean.
4. `golangci-lint run ./...` reports no issues (complexity included).
5. If the change affects TOC output, regenerate this repo's `README.md` with
   the built binary to confirm behavior end to end.
