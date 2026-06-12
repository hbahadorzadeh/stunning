# CI/CD

Stunning uses GitHub Actions. Three workflows:

| Workflow | File | Trigger | Purpose |
| -------- | ---- | ------- | ------- |
| **CI** | `workflows/ci.yml` | push to `main`/`master`/`develop`, all PRs | lint, security, build, test, benchmarks, DPI e2e |
| **Release** | `workflows/release.yml` | version tag push | build + publish all platform artifacts |
| **CLA Assistant** | `workflows/cla.yml` | PR open/sync, issue comment | enforce [CLA](../CLA.md) signing on PRs |

Requires **Go 1.25+**.

---

## CI pipeline (`ci.yml`)

| Job | What it does | Blocks merge |
| --- | ------------ | :----------: |
| `lint` | `gofmt`, `go vet`, `golangci-lint` | ✅ |
| `sast` | `gosec` security scan (fails on HIGH/CRITICAL) | ✅ |
| `build` | build core packages + CLI binary (Docker) | ✅ |
| `unit-tests` | unit tests + coverage → Codecov | ✅ |
| `code-quality` | `staticcheck`, `revive` | ✅ |
| `e2e-dpi` | 3-node docker DPI harness asserts chains evade a simulated firewall (`test/dpi/scenarios.sh`) | ✅ |
| `integration-tests` | protocol + library tests in Docker | ℹ️ |
| `race-detector` | `go test -race` | ℹ️ |
| `dependencies` | `nancy` vuln scan + `go mod verify` | ℹ️ |
| `benchmarks` | `core/plugin` micro-benchmarks (artifact) | ℹ️ |
| `summary` | aggregates results | ℹ️ |

### Artifacts

| Artifact | Retention |
| -------- | --------- |
| CLI binary | 5 days |
| gosec security report | 30 days |
| plugin benchmark results | 14 days |
| DPI verdict log | 14 days |
| coverage | uploaded to Codecov |

---

## Release pipeline (`release.yml`)

Triggered by a version tag. Builds and publishes per-platform artifacts:

| Job | Output |
| --- | ------ |
| `prepare` | resolve version, set up matrix |
| `release-cli` | `stunning-cli-<os>-<arch>` (linux/macos/windows, amd64/arm64) |
| `release-library` | `libstunning-<os>-<arch>.tar.gz` (shared lib + header) |
| `release-desktop` | `stunning-desktop-<os>-<arch>` (Fyne GUI) |
| `release-android` | `stunning-android.apk`, `libstunning.aar` |
| `release-ios` | `stunning-ios-xcframework.zip` |
| `create-github-release` | GitHub Release + `SHA256SUMS.txt` |
| `publish-google-play` | upload to Google Play (when configured) |
| `publish-app-store` | upload to App Store (when configured) |

---

## Running checks locally

```bash
# everything CI runs, the short way
go build ./...
go test -race -coverprofile=coverage.out ./...
golangci-lint run --timeout=5m
gosec ./...
```

Individual checks:

```bash
gofmt -s -l .                       # format check (should print nothing)
go vet ./...                        # vet
staticcheck ./...                   # static analysis
revive -set_exit_status ./...       # lint rules
go mod verify                       # module integrity
go list -json -m all | nancy sleuth # dependency vulnerabilities
```

Plugin benchmarks and the DPI e2e harness:

```bash
go test -run='^$' -bench=. -benchmem ./core/plugin/
test/dpi/build.sh && test/dpi/scenarios.sh
```

Tool installs:

```bash
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install github.com/sonatype-nexus-community/nancy@latest
```

---

## Configuration

| File | Role |
| ---- | ---- |
| `.github/workflows/ci.yml` | CI jobs |
| `.github/workflows/release.yml` | release/publish jobs |
| `.github/workflows/cla.yml` | CLA enforcement |
| `.github/dependabot.yml` | weekly Go + Actions updates, assigned to `hbahadorzadeh` |
| `.golangci.yml` | golangci-lint rules |
| `.revive.toml` | revive rules |

---

## Troubleshooting

**Build passes locally, fails on CI** — commit `go.mod`/`go.sum`; run `go mod tidy`.

**Lint failures** — `gofmt -s -w .` then `golangci-lint run --fix`.

**Tests flaky** — run with `-race`; check for concurrent access.

**Dependency vulnerability** — `go get -u <module>`, `go mod tidy`, open a PR.

For CI/CD issues, open a GitHub issue with the failing job, error message, and
local reproduction steps.
