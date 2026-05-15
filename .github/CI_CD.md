# CI/CD Pipeline Documentation

## Overview

The Stunning project uses GitHub Actions for continuous integration and continuous deployment. The CI pipeline runs on every push to main/master/develop branches and on all pull requests.

## Pipeline Stages

### 1. Lint & Format Check
- **Purpose**: Ensures code style consistency
- **Tools**:
  - `gofmt` - Code formatting
  - `go vet` - Correctness analysis
  - `golangci-lint` - Comprehensive linting
- **Failure Condition**: Code not properly formatted or vet issues found
- **Local Execution**: `gofmt -s -l . && go vet ./... && golangci-lint run`

### 2. SAST Security Scan
- **Purpose**: Identifies security vulnerabilities and weaknesses
- **Tools**:
  - `gosec` - Go security scanner
  - Searches for: SQL injection, hardcoded secrets, weak crypto, etc.
- **Failure Condition**: HIGH or CRITICAL severity issues found
- **Local Execution**: `gosec ./...`

### 3. Build
- **Purpose**: Verifies code compiles correctly
- **Steps**:
  1. Build all packages: `go build ./...`
  2. Build main binary: `go build -v -o stunning`
- **Failure Condition**: Build errors
- **Local Execution**: `go build ./...`

### 4. Unit Tests
- **Purpose**: Runs all unit tests with race detection and coverage
- **Coverage**: Reported to Codecov
- **Failure Condition**: Test failures
- **Local Execution**: `go test -v -race -coverprofile=coverage.out ./...`

### 5. Functional Tests
- **Purpose**: Tests individual component functionality
- **Tests**:
  - Tunnel factory pattern tests
  - Interface implementation tests
  - Connection recovery tests
- **Local Execution**: `go test -v -timeout=30s -run 'TestTunnel|TestInterface|TestRecovery' ./...`

### 6. E2E Tests
- **Purpose**: Tests complete tunnel workflows end-to-end
- **Tests**:
  - TCP tunnel with data transmission
  - SOCKS proxy functionality
  - Concurrent connections
  - Connection recovery
- **Local Execution**: `go test -v -timeout=60s -run 'TestE2E' ./...`

### 7. Race Detector
- **Purpose**: Detects data race conditions
- **Coverage**: All code paths
- **Failure Condition**: Any race condition detected
- **Local Execution**: `go test -race ./...`

### 8. Dependency Check
- **Purpose**: Identifies vulnerable and outdated dependencies
- **Tools**:
  - `nancy` - Vulnerable dependency scanner
  - `go mod verify` - Module integrity check
  - Go outdated check
- **Local Execution**: `go mod verify && nancy sleuth`

### 9. Code Quality Metrics
- **Purpose**: Comprehensive code quality analysis
- **Tools**:
  - `staticcheck` - Advanced static analysis
  - `revive` - Go linter with rules
- **Local Execution**: `staticcheck ./... && revive -set_exit_status ./...`

## Running Tests Locally

### Prerequisites
```bash
go version # 1.21+
git clone https://github.com/hbahadorzadeh/stunning.git
cd stunning
go mod tidy
```

### Run All Tests
```bash
go test -v -race -coverprofile=coverage.out ./...
```

### Run Specific Test Categories

**Unit Tests Only**:
```bash
go test -v ./... -skip E2E
```

**Functional Tests**:
```bash
go test -v -timeout=30s -run 'TestTunnel|TestInterface|TestRecovery' ./...
```

**E2E Tests**:
```bash
go test -v -timeout=60s -run 'TestE2E' ./...
```

**With Race Detection**:
```bash
go test -race -timeout=60s ./...
```

### Run Linters Locally

**Format Check**:
```bash
gofmt -s -l .
go vet ./...
```

**Install golangci-lint**:
```bash
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

**Run golangci-lint**:
```bash
golangci-lint run --timeout=5m
```

**Security Scan**:
```bash
go install github.com/securego/gosec/v2/cmd/gosec@latest
gosec ./...
```

**Dependency Check**:
```bash
go mod verify
go install github.com/sonatype-nexus-community/nancy@latest
go list -json -m all | nancy sleuth
```

## Pipeline Configuration

### `.github/workflows/ci.yml`
Main CI workflow defining all jobs and steps.

### `.github/dependabot.yml`
Dependabot configuration for automated dependency updates:
- Go dependencies: Weekly updates
- GitHub Actions: Weekly updates

### `.golangci.yml`
Linting rules and configuration for golangci-lint.

## Dependency Management

### Automatic Updates
Dependabot automatically creates pull requests for:
- **Go dependencies**: Check for updates weekly
- **GitHub Actions**: Check for updates weekly
- **Security patches**: Immediate updates for security vulnerabilities

### Manual Updates
```bash
go get -u ./...
go mod tidy
```

## Secrets & Security

### No Sensitive Data
- No API keys, tokens, or credentials in code
- Use GitHub Secrets for any CI/CD secrets
- Security reports are uploaded as artifacts

### Artifact Retention
- Binaries: 5 days
- Security reports: 30 days (default)

## Troubleshooting

### Build Fails on New Branch
- Ensure `go.mod` and `go.sum` are committed
- Run `go mod tidy` locally and commit

### Tests Fail Locally But Pass on CI
- Use `go test -race` to check for race conditions
- Ensure all dependencies are installed: `go mod download`
- Check Go version: `go version` (must be 1.21+)

### Lint Issues
- Run `gofmt -s -w .` to auto-format
- Run `golangci-lint run --fix` to auto-fix many issues
- Some issues require manual fixes (see golangci-lint output)

### Dependency Vulnerability
- Check the advisory: `go list -u -m all`
- Update: `go get -u <module>`
- Create PR for the update

## Performance Optimization

- **Caching**: Go modules cached between runs
- **Parallel Jobs**: Most jobs run in parallel
- **Timeout**: 5 minutes per job (60 minutes for E2E)

## Status Badges

Add to README.md:
```markdown
[![CI](https://github.com/hbahadorzadeh/stunning/workflows/CI/badge.svg)](https://github.com/hbahadorzadeh/stunning/actions)
[![codecov](https://codecov.io/gh/hbahadorzadeh/stunning/branch/master/graph/badge.svg)](https://codecov.io/gh/hbahadorzadeh/stunning)
```

## Contributing

When contributing:
1. Ensure all tests pass locally: `go test -race ./...`
2. Run linters: `golangci-lint run`
3. Check security: `gosec ./...`
4. Ensure code is formatted: `gofmt -s -w .`

The CI pipeline will verify all checks before merging.

## Questions?

For issues with CI/CD, please open a GitHub issue with:
- CI job that failed
- Error message
- Steps to reproduce locally
