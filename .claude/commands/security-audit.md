# Security Audit

Perform comprehensive security audit using the security-engineer agent.

## Steps

1. Use the security-engineer agent to analyze the codebase

2. Check for common vulnerabilities:
   - **Authentication**:
     - Password storage (currently plaintext ⚠️)
     - Session management
     - Brute force protection

   - **Input Validation**:
     - Command injection risks
     - SQL injection (if database added)
     - Path traversal
     - Buffer overflows

   - **Data Security**:
     - Sensitive data exposure
     - Insecure file permissions
     - Logging sensitive information

   - **Network Security**:
     - Unencrypted connections
     - Rate limiting
     - DDoS protection

   - **Code Security**:
     - Race conditions
     - Use of unsafe functions
     - Error information disclosure
     - Hardcoded secrets

3. Run automated security scanners:
   ```bash
   # Run gosec
   gosec ./...

   # Check dependencies for vulnerabilities
   go list -json -m all | nancy sleuth
   ```

4. Review authentication system in detail:
   - Analyze `authenticate()` function in main.go
   - Check password storage in data/users.json
   - Identify security risks

5. Generate security report with:
   - **Critical Issues**: Must fix immediately
   - **High Priority**: Fix before production
   - **Medium Priority**: Fix soon
   - **Low Priority**: Nice to have
   - **Informational**: Be aware of

6. For each issue, provide:
   - Description of vulnerability
   - Potential impact
   - Code location
   - Recommended fix with example
   - Priority level

7. Create action plan:
   - Immediate fixes (critical/high)
   - Short-term improvements (medium)
   - Long-term enhancements (low)

## Example Report

```
# Security Audit Report - Matrix MUD

## Critical Issues

### 1. Plaintext Password Storage
**Location**: main.go:25-29, data/users.json
**Risk**: User credentials compromised if data/ directory accessed
**Impact**: All user accounts vulnerable

**Current Code**:
\```go
if pass == storedPass { return true }
\```

**Recommended Fix**:
\```go
import "golang.org/x/crypto/bcrypt"

// Hash password on creation
hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

// Verify password
err := bcrypt.CompareHashAndPassword(storedHash, []byte(pass))
\```

**Priority**: CRITICAL - Fix immediately

## High Priority

### 2. No Rate Limiting
**Location**: main.go:handleConnection
**Risk**: Brute force attacks, spam, DoS
**Impact**: Server overload, resource exhaustion

**Recommended Fix**: Implement token bucket rate limiter

...
```

## Notes

- This audit is most effective on actual production code
- Run regularly (monthly recommended)
- Track remediation progress
- Retest after fixes
- Consider professional penetration testing for production
