# GitHub Actions & CI/CD Setup

This directory contains the GitHub Actions CI/CD pipeline configuration for the Stunning project.

## Files Overview

### Workflows
- **`workflows/ci.yml`** - Main CI pipeline with 12 jobs covering linting, security, testing, and reporting

### Configuration
- **`dependabot.yml`** - Automated dependency and GitHub Actions updates
- **`CI_CD.md`** - Comprehensive documentation on the CI/CD pipeline

### Root Level
- **`.golangci.yml`** - Linting rules (50+ linters)
- **`SECURITY.md`** - Security policy and vulnerability reporting

## Quick Start

### Push to Repository
The CI pipeline automatically runs on:
- Push to `main`, `master`, or `develop` branches
- All pull requests

### View Results
1. Go to: https://github.com/hbahadorzadeh/stunning/actions
2. Select the workflow run
3. View detailed logs for each job

### Local Testing
```bash
# Run all checks locally
go test -race ./...
golangci-lint run
gosec ./...
```

See `.github/CI_CD.md` for detailed local testing commands.

## Pipeline Overview

### Critical Jobs (Blocks Merge)
1. **Lint & Format** - Code quality standards
2. **SAST Security** - Security vulnerability scanning
3. **Build** - Compilation verification
4. **Unit Tests** - Test coverage and race detection
5. **Code Quality** - Advanced static analysis

### Informational Jobs (Non-Blocking)
1. **Functional Tests** - Component tests
2. **E2E Tests** - End-to-end workflows
3. **Race Detector** - Concurrency verification
4. **Dependency Check** - Vulnerable packages

## Automated Dependency Updates

Dependabot automatically creates pull requests:
- **Go Dependencies**: Weekly (Mondays 3:00 AM UTC)
- **GitHub Actions**: Weekly (Mondays 4:00 AM UTC)

All updates are assigned to `hbahadorzadeh` for review.

## Security

The pipeline includes:
- **SAST Scanning** with gosec
- **Dependency Vulnerability** scanning with nancy
- **Race Condition** detection
- **Code Quality** enforcement

See `SECURITY.md` for vulnerability reporting procedures.

## Artifacts

Generated and stored:
- **Compiled binaries** (5 days retention)
- **Security reports** (30 days retention)
- **Coverage data** (uploaded to Codecov)

## Status Badges

Add to README.md:
```markdown
[![CI](https://github.com/hbahadorzadeh/outstanding/workflows/CI/badge.svg)](https://github.com/hbahadorzadeh/outstanding/actions)
```

## Configuration Highlights

- **Go Version**: 1.21+
- **Timeout**: 5 minutes (linting), 60 seconds (tests)
- **Coverage**: Codecov integration enabled
- **Caching**: Go modules cached for faster builds
- **Parallel**: Most jobs run in parallel

## For Contributors

Before creating a pull request:
1. Ensure code is formatted: `gofmt -s -w .`
2. Run all tests locally: `go test -race ./...`
3. Check security: `gosec ./...`
4. Run linters: `golangci-lint run`

The CI pipeline will automatically verify all checks.

## Troubleshooting

**Build fails on CI but passes locally?**
- Ensure `go.mod` and `go.sum` are committed
- Run `go mod tidy` and commit the results

**Lint failures?**
- Run `golangci-lint run --fix` to auto-fix many issues
- See `.golangci.yml` for complete rules

**Test failures?**
- Run with `-race` flag: `go test -race ./...`
- Check for concurrent access issues

## Documentation

Detailed documentation:
- Pipeline jobs: `.github/CI_CD.md`
- Security policy: `SECURITY.md`
- Linting rules: `.golangci.yml`

## Questions?

For CI/CD questions or issues:
1. Check `.github/CI_CD.md` for troubleshooting
2. Open a GitHub issue with details about the failure
3. Email: h.bahadorzadeh@gmail.com
