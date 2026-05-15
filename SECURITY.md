# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

**Do not** file public issues for security vulnerabilities. Instead, please report security issues responsibly by emailing [h.bahadorzadeh@gmail.com](mailto:h.bahadorzadeh@gmail.com) with:

1. **Description**: A clear description of the vulnerability
2. **Location**: The file and line numbers affected
3. **Impact**: The potential impact of the vulnerability
4. **Proof of Concept**: Steps to reproduce or a minimal proof of concept
5. **Suggested Fix**: If you have a suggested fix, please include it

## Security Considerations

### TLS Configuration
- TLS certificate verification is **enabled by default** (`InsecureSkipVerify: false`)
- For development/testing, custom TLS configurations can be provided via `GetTlsDialerWithConfig()`
- Never disable certificate verification in production

### Authentication
- This library does NOT provide authentication mechanisms
- Implement proper authentication at the application level
- Use TLS for transport security

### Network Isolation
- Ensure tunnels are properly isolated and not exposed to untrusted networks
- Use firewall rules to restrict access to tunnel endpoints
- Monitor tunnel connections for suspicious activity

### Resource Management
- Monitor connection counts to prevent resource exhaustion
- Implement timeouts on long-running connections
- Use connection pooling where appropriate

## Security Tools

The CI/CD pipeline includes:

1. **gosec** - Go static security analyzer
   - Detects common security issues
   - Scans for vulnerable patterns
   - Fails on HIGH/CRITICAL severity issues

2. **Nancy** - Vulnerable dependency checker
   - Identifies vulnerable third-party packages
   - Runs as part of dependency checks

3. **Staticcheck** - Advanced static analysis
   - Detects code issues
   - Identifies potential problems

4. **Race Detector** - Concurrency issue detection
   - Identifies data races
   - Helps prevent concurrent access bugs

## Security Updates

Security updates are released as patch versions (e.g., 1.0.1, 1.0.2).

Subscribe to release notifications to stay informed of security updates.

## Best Practices

1. **Keep dependencies updated**: Use Dependabot automated updates
2. **Monitor security advisories**: Check GitHub Security Advisories regularly
3. **Use TLS in production**: Always use TLS/HTTPS for tunnel connections
4. **Restrict network access**: Use firewalls to limit tunnel endpoint access
5. **Implement logging**: Log tunnel connections for security monitoring
6. **Regular testing**: Run security tests regularly (gosec, nancy)
7. **Code review**: Have security-sensitive code reviewed by team members

## Vulnerability Disclosure Timeline

Once a security vulnerability is reported:

1. **Acknowledgment**: We acknowledge receipt within 24-48 hours
2. **Assessment**: We assess the vulnerability (3-5 business days)
3. **Fix Development**: We develop and test a fix (varies by severity)
4. **Release**: We release a patch version with the fix
5. **Public Disclosure**: We publicly disclose the vulnerability after a fix is released

## Security Headers & Practices

When using Stunning in production:

1. Use HTTPS/TLS for all tunnel connections
2. Implement rate limiting on tunnel endpoints
3. Monitor and log all tunnel activity
4. Use strong TLS cipher suites
5. Keep Go and dependencies updated
6. Run the application with least privilege

## Third-Party Dependencies

All third-party dependencies are:
- Vetted for security
- Scanned with Nancy and gosec
- Automatically updated with Dependabot
- Monitored for vulnerabilities

Current dependencies:
- `github.com/getlantern/go-socks5` - SOCKS5 protocol support
- `github.com/jacobsa/go-serial` - Serial communication
- `github.com/rainycape/dl` - Dynamic library loading
- `github.com/songgao/water` - TUN/TAP interface
- `github.com/yuin/gopher-lua` - Lua scripting support
- `golang.org/x/net` - Extended network utilities

## Questions?

For security questions (non-vulnerability related), please email [h.bahadorzadeh@gmail.com](mailto:h.bahadorzadeh@gmail.com).
