# Phase 4: Release - Research

**Researched:** 2026-03-06
**Domain:** Git branching, GitHub Actions CI/CD, Go cross-compilation, GitHub Releases
**Confidence:** HIGH

## Summary

Phase 4 is purely release infrastructure — no new feature code. The goal is to land all Phase 1-3 work via a PR to main, validate it with CI, then publish a tagged GitHub release (v0.3.0) with binaries for 6 platform targets.

All release infrastructure already exists: `ci.yml` runs format/vet/test/build on every PR, and `release.yml` builds 6 cross-compiled binaries and publishes them to a GitHub Release on any `v*.*.*` tag push. The Makefile already has a `release` target that tags and pushes. The `make release TAG=v0.3.0` command is the full release procedure.

The work for this phase is operational, not technical: verify the branch is clean, open the PR, wait for CI, merge, tag, and confirm the release workflow publishes assets.

**Primary recommendation:** Open PR from `feat/v0.3.0` to `main`, let CI pass, merge, then run `make release TAG=v0.3.0` from main.

## Current State Audit

### Branch state
- Current branch: `feat/v0.3.0`
- Commits ahead of main: **53 commits** (60 files changed, 8,863 insertions, 101 deletions)
- Remote tracking: `remotes/origin/feat/v0.3.0` exists
- Dirty files: `cmd/task.go` (uncommitted fix), `.planning/config.json` (minor change)

### Test suite
- `go test ./...` — all packages **pass** (cached)
- `gofmt -l .` — **no unformatted files**
- `go vet ./...` — **no issues**

### Pending dirty state
`cmd/task.go` has a staged change: early-exit when no completed tasks match bulk delete (empty slice guard before the confirmation prompt). This is a correctness fix that must be committed before opening the PR.

### Existing CI workflows
| File | Trigger | What it does |
|------|---------|--------------|
| `.github/workflows/ci.yml` | push to main/master + all PRs | gofmt check, go vet, go test, go build |
| `.github/workflows/release.yml` | push tag `v*.*.*` + workflow_dispatch | Cross-compile 6 targets, upload artifacts, create GH release with checksums |

### Existing infrastructure
| Component | Detail |
|-----------|--------|
| `make release TAG=vX.Y.Z` | Tags HEAD + pushes tag to origin |
| `main.go` | `var version = "dev"` — set at build time via `-ldflags "-X main.version=<tag>"` |
| Release binary names | `dtasks-{macos,linux,windows}-{amd64,arm64}[.exe]` — 6 files |
| Checksums | `sha256sum * > checksums.txt` — included in release assets |
| GH release notes | `--generate-notes` (auto from commit messages) |
| `workflow_dispatch` | Manual release trigger with explicit tag input |

### Previous release history
| Tag | Notes |
|-----|-------|
| v0.1.0 | First release |
| v0.2.0 | Recurrence + CLI output improvement |
| v0.3.0 | Target for this phase |

## Architecture Patterns

### Release workflow (existing pattern)

```
feat/v0.3.0 ──► PR to main ──► CI passes ──► merge ──► make release TAG=v0.3.0
                                                                │
                                                    git tag v0.3.0 + git push origin v0.3.0
                                                                │
                                                    release.yml fires ──► 6 binaries ──► GH Release
```

### CI workflow checks (from ci.yml)
1. `gofmt -l .` — fails if any file not formatted
2. `go vet ./...` — static analysis
3. `go test ./...` — full test suite
4. `go build ./...` — compile check

### Release workflow steps (from release.yml)
1. Checkout with `fetch-depth: 0` (needed for `git describe`)
2. Matrix build: 6 combinations (darwin/linux/windows × amd64/arm64)
3. `CGO_ENABLED=0` — pure Go, no cgo
4. `-ldflags="-s -w -X main.version=${TAG}"` — strips debug, embeds version
5. Upload each binary as GitHub artifact (retention: 1 day)
6. Download all artifacts, generate `checksums.txt`, create GH release

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Cross-compilation | Custom build scripts | Existing `release.yml` matrix | Already handles 6 targets with correct ldflags |
| Release creation | Manual `gh release create` | `make release TAG=v0.3.0` | Existing Makefile target |
| Checksums | Manual sha256 | `release.yml` generates `checksums.txt` | Already in workflow |
| Version embedding | Runtime version detection | `-X main.version=<tag>` ldflags | Already wired in Makefile and release.yml |

## Common Pitfalls

### Pitfall 1: Committing dirty files before PR
**What goes wrong:** Opening a PR with uncommitted changes to `cmd/task.go` means the fix is invisible to reviewers and not in the commit history.
**Why it happens:** The file was modified as part of an in-progress fix (empty-slice guard in rmCmd).
**How to avoid:** Commit `cmd/task.go` fix before pushing/opening PR.
**Warning signs:** `git status` shows `M cmd/task.go`.

### Pitfall 2: Tagging before merging to main
**What goes wrong:** `make release TAG=v0.3.0` run from `feat/v0.3.0` branch tags the feature branch HEAD, not main. The release.yml fires but the main branch doesn't contain the release commit.
**Why it happens:** `git tag` + `git push origin TAG` works from any branch.
**How to avoid:** Always merge to main first, checkout main, pull, then tag.

### Pitfall 3: CI triggers on PR but release.yml doesn't
**What goes wrong:** Assuming CI and release workflows are the same — CI runs on PRs, release.yml only fires on tag push.
**How to avoid:** Explicit sequence: PR (CI) → merge → tag push (release.yml). They are separate triggers.

### Pitfall 4: gofmt check fails on CI
**What goes wrong:** CI fails with "Unformatted files" if any `.go` file was edited without running `gofmt`.
**How to avoid:** Run `gofmt -l .` locally before pushing. Currently passes on this branch.

### Pitfall 5: `version` shows "dev" in release binary
**What goes wrong:** If tag is not passed via ldflags, `dtasks update` reports version "dev" which breaks version comparison logic.
**Why it happens:** `release.yml` uses `github.ref_name` for the tag — only works when triggered by a tag push, not by `workflow_dispatch` with empty input.
**How to avoid:** Use `make release TAG=v0.3.0` (tag push) rather than workflow_dispatch. The workflow correctly uses `github.event.inputs.tag || github.ref_name`.

## Code Examples

### Commit the pending fix
```bash
# Source: git status shows M cmd/task.go
git add cmd/task.go
git commit -m "fix(rm): exit early when no completed tasks match bulk delete"
```

### Push branch and open PR
```bash
git push origin feat/v0.3.0
gh pr create \
  --title "feat: v0.3.0 — querying, richness, tooling" \
  --base main \
  --body "..."
```

### After merge: tag and release
```bash
git checkout main
git pull origin main
make release TAG=v0.3.0
```

### Verify release
```bash
gh release view v0.3.0
# Confirm 7 assets: 6 binaries + checksums.txt
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Manual binary upload | `release.yml` matrix + `gh release create --generate-notes` | Already in place (v0.1.0) | No manual steps needed |
| `workflow_dispatch` only | Tag push + `workflow_dispatch` | Added in PR #2 | `make release TAG=` is the canonical path |

## Open Questions

1. **PR description content**
   - What we know: 53 commits, 60 files, all 32 requirements implemented
   - What's unclear: Level of detail desired in PR body
   - Recommendation: Planner should include a PR description task that summarizes requirements by phase (FILT/SORT/SRCH/PRIO/MAINT/STAT/UPDT/COMP/SKIL)

2. **`.planning/config.json` dirty file**
   - What we know: `git status` shows it modified
   - What's unclear: Whether this should be committed or the change discarded
   - Recommendation: Include as a task — either commit with docs or `git checkout -- .planning/config.json`

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing package (stdlib) |
| Config file | none (go test ./... is self-discovering) |
| Quick run command | `go test ./... -count=1` |
| Full suite command | `go test ./... -count=1 -race` |

### Phase Requirements → Test Map

Phase 4 has no REQUIREMENTS.md entries. Validation is operational: CI must pass on the PR, release assets must appear after tagging.

| Gate | Behavior | Type | Command |
|------|----------|------|---------|
| Pre-PR | All tests pass locally | automated | `go test ./...` |
| Pre-PR | No unformatted files | automated | `gofmt -l .` |
| Pre-PR | No vet issues | automated | `go vet ./...` |
| CI gate | Same checks on PR | CI (automated) | triggered by `git push` |
| Release gate | 6 binaries + checksums.txt appear | manual verification | `gh release view v0.3.0` |

### Wave 0 Gaps
None — existing test infrastructure covers all phase requirements. Release phase has no new code to test.

## Sources

### Primary (HIGH confidence)
- `.github/workflows/ci.yml` — exact CI steps executed on PRs
- `.github/workflows/release.yml` — exact release pipeline, matrix targets, asset names
- `Makefile` — `release` target behavior, ldflags, binary names
- `main.go` — version variable wiring
- `git log`, `git status`, `git diff` — verified current branch state

### Secondary (MEDIUM confidence)
- `go.mod` — confirmed Go 1.24.0, module path `github.com/danielmrdev/dtasks-cli`
- `gh pr list` — confirmed previous merge pattern (feature branch → PR → merge)

## Metadata

**Confidence breakdown:**
- Current branch state: HIGH — verified via git commands
- CI workflow behavior: HIGH — read actual workflow files
- Release workflow behavior: HIGH — read actual workflow files
- Operational sequence: HIGH — derived from existing Makefile + workflow combination

**Research date:** 2026-03-06
**Valid until:** Stable — workflow files don't change between now and tagging
