# AGENTS.md

## Project

You are working on:

**PMOS Control Panel**

A lightweight web-based Linux management panel for a postmarketOS device.

Read `PLAN.md` before making architectural or implementation decisions.

---

## Current Status (2026-09-18)

**Phase 1-10: DONE** — All core features implemented and deployed on real hardware.

| Phase | Feature | Status |
|-------|---------|--------|
| 0 | Discovery | DONE |
| 1 | Foundation (Go + Svelte + Auth) | DONE |
| 2 | Monitoring (CPU/RAM/Storage/Temp) | DONE |
| 3 | Terminal (WebSocket PTY) | DONE |
| 4 | Systemd Service Manager | DONE |
| 5 | Process Manager | DONE |
| 6 | File Manager | DONE |
| 7 | Logs + Live Streaming | DONE |
| 8 | Network Monitor | DONE |
| 9 | Power (Reboot/Shutdown) | DONE |
| 10 | Production Deployment | DONE |
| 11 | Security Audit | TODO |
| 12 | Optimization | TODO |

### Recent Fixes Applied
- `process.Kill()` fixed (was using os.Remove, now uses syscall.Kill)
- Secret key generation uses crypto/rand (was deterministic)
- WebSocket CheckOrigin validates origin header
- Network stats delta calculation fixed
- File upload added (backend + frontend)
- Live log streaming via WebSocket added
- Mobile hamburger menu added

### Deployment (2026-09-18)
- Device: postmarketOS v26.06 on qcom msm8953
- Init: **OpenRC** (not systemd)
- Binary: `/usr/local/bin/pmos-panel`
- Config: `/home/fw/config.json`
- Service: `/etc/init.d/pmos-panel` (OpenRC)
- Firewall: nftables `inet filter input tcp dport 8080`
- Build: Cross-compiled from Windows laptop (GOOS=linux GOARCH=arm64)
- No Go/Node.js needed on device (embedded frontend)
- Auto-start on boot: `rc-update add pmos-panel default`

### Known Issues / Future Work
- Phase 11: Security audit (CSRF tokens, CSP headers, rate limiting on all endpoints)
- Phase 12: Performance optimization and benchmarking
- Settings page is a placeholder

---

# 1. Core Rules

You are an implementation agent.

Your job is to build the project described in `PLAN.md`.

Do not redesign the entire project without a technical reason.

Do not implement the entire project in one step.

Work phase by phase.

---

# 2. Target Environment

The target device is:

```text
OS: postmarketOS
Channel: stable
Vendor: qcom
Init: systemd
UI: fbkeyboard
Architecture: ARM/ARM64
```

Important:

**The target uses systemd.**

Do not design the application around OpenRC.

Do not assume a full desktop environment exists.

The web panel must work as a background service.

---

# 3. Backend Rules

Backend language:

```text
Go
```

Prefer Go standard library where practical.

Prefer:

```text
net/http
```

Avoid unnecessary frameworks.

Keep dependencies minimal.

The final backend should be deployable as a single binary.

---

# 4. Frontend Rules

Frontend:

```text
Svelte
TypeScript
CSS
```

Keep the frontend lightweight.

Do not introduce large UI libraries unless there is a strong reason.

Prioritize:

- fast loading
- low JavaScript overhead
- responsive UI
- mobile support
- desktop support
- dark mode

---

# 5. Architecture Rules

Use this architecture:

```text
Frontend
   │
REST / WebSocket
   │
Go Backend
   │
Service Layer
   │
Linux / systemd / filesystem
```

Frontend must never directly execute Linux commands.

Do not couple frontend components directly to system commands.

---

# 6. System Access

Prefer native Linux interfaces when practical.

Examples:

```text
/proc
/sys
statfs
systemd
journald
```

Do not parse command output when a reliable system interface/library is available.

When commands are necessary, isolate them inside backend service packages.

---

# 7. systemd

The target uses systemd.

Use:

```text
systemctl
journalctl
```

only through backend service abstractions.

Never expose arbitrary systemctl arguments directly to the browser.

Bad:

```text
POST /api/command
{
  "command": "systemctl ..."
}
```

Good:

```text
POST /api/services/sshd/restart
```

Backend validates the service name and operation.

---

# 8. Terminal

The web terminal must use a real Linux PTY.

Do not implement the terminal as simple command execution.

Architecture:

```text
Browser
   │
WebSocket
   │
PTY
   │
Shell
```

Terminal must support:

- stdin
- stdout
- stderr
- resize
- interactive applications
- disconnect cleanup

Terminal access requires authentication.

---

# 9. Security

Security is a first-class requirement.

Always consider:

```text
Authentication
Authorization
CSRF
XSS
Command injection
Path traversal
Symlink escape
Session security
WebSocket authentication
File upload security
Privilege escalation
Rate limiting
```

Never trust browser input.

Validate all API parameters on the backend.

---

# 10. Authentication

Never store plaintext passwords.

Use secure password hashing.

Use secure sessions.

Cookies should use appropriate:

```text
HttpOnly
SameSite
Secure
```

attributes where applicable.

Authentication must apply to:

- API
- terminal WebSocket
- file operations
- service control
- power control

---

# 11. Privileges

Do not run the entire backend as root unless absolutely necessary.

Prefer:

```text
normal Linux user
```

For privileged actions, use a narrowly scoped mechanism.

Never create:

```text
sudo ALL=(ALL) NOPASSWD: ALL
```

Do not give the web application unrestricted root access.

---

# 12. Filesystem Security

The file manager must have a configurable root directory.

Default:

```text
/home/<panel-user>
```

Every path must be canonicalized.

Reject:

```text
../
../../
absolute paths outside allowed root
```

Be careful with symbolic links.

Do not expose the entire filesystem in the MVP.

---

# 13. Network

The initial network feature must be read-only.

Do not automatically modify:

```text
Wi-Fi
IP
routes
DNS
NetworkManager
```

Changing network configuration can disconnect the device from the panel.

---

# 14. Destructive Operations

The following operations require explicit authorization and confirmation:

```text
Reboot
Shutdown
Kill process
Delete file
Stop service
```

Do not perform destructive operations through GET requests.

Prefer:

```text
POST
```

---

# 15. Development Process

Always work in phases.

Correct workflow:

```text
Read plan
   ↓
Inspect
   ↓
Plan
   ↓
Implement
   ↓
Build
   ↓
Test
   ↓
Review
   ↓
Report
```

Do not skip testing.

---

# 16. Phase Discipline

When asked to implement Phase N:

Implement **only Phase N** unless a dependency from an earlier phase must be fixed.

Do not start future phases automatically.

Example:

If asked:

```text
Implement Phase 2
```

do not also implement:

```text
Terminal
File Manager
Process Manager
Power
```

unless explicitly requested.

---

# 17. Before Coding

Before the first implementation:

1. Read `PLAN.md`.
2. Inspect repository.
3. Inspect development environment.
4. Determine Go version.
5. Determine frontend tooling.
6. Verify target assumptions where possible.
7. Produce a short implementation plan.

Do not blindly assume the environment.

---

# 18. Testing

After implementation:

Run appropriate tests.

At minimum:

```text
go test ./...
go vet ./...
go build ./...
```

For frontend, run its appropriate:

```text
lint
typecheck
build
test
```

commands.

Fix compilation and test errors before reporting completion.

---

# 19. Code Quality

Prefer:

- small packages
- clear interfaces
- dependency injection where useful
- explicit error handling
- structured logging
- context cancellation
- graceful shutdown
- readable code

Avoid:

- huge files
- global mutable state
- unnecessary abstractions
- duplicated logic
- magic constants
- hardcoded usernames
- hardcoded IP addresses
- hardcoded paths

---

# 20. Resource Constraints

Remember that the target is a phone being used as a server.

Optimize for:

```text
RAM
CPU
Storage
Battery
Startup time
```

Avoid unnecessary:

- background workers
- polling
- dependencies
- databases
- caches
- logging volume

Do not poll system metrics excessively.

---

# 21. Logging

Use structured application logging.

Do not continuously spam logs.

Production logs should be useful when viewed through:

```text
journalctl -u pmos-panel.service
```

---

# 22. Configuration

Do not hardcode:

```text
password
secret
username
IP
port
home directory
```

unless it is a safe default.

Configuration should be explicit and documented.

---

# 23. Error Handling

Errors must be handled deliberately.

Do not silently ignore errors.

Do not expose internal stack traces or sensitive system information to the browser.

Backend logs may contain diagnostic details, but API responses should provide safe errors.

---

# 24. API Design

Keep API endpoints predictable.

Example:

```text
GET  /api/system
GET  /api/stats
GET  /api/processes
GET  /api/services
GET  /api/logs
GET  /api/files

POST /api/services/{name}/start
POST /api/services/{name}/stop
POST /api/services/{name}/restart

POST /api/processes/{pid}/kill

POST /api/system/reboot
POST /api/system/shutdown

WS /api/ws/stats
WS /api/ws/terminal
WS /api/ws/logs
```

The exact API can evolve, but maintain consistency.

---

# 25. Git

Make logical commits.

Example:

```text
feat: add system information API
feat: add dashboard metrics
feat: add websocket monitoring
feat: add systemd service manager
fix: prevent file path traversal
test: add authentication tests
```

Do not create meaningless commits for every tiny change.

---

# 26. Documentation

Update documentation when behavior changes.

Important documentation:

```text
README.md
Installation
Development
Configuration
Deployment
Security
API
Troubleshooting
```

---

# 27. Do Not Do This

Never:

- install random packages without checking
- use OpenRC assumptions
- run everything as root
- expose arbitrary shell execution
- expose `/` through the file manager
- hardcode credentials
- hardcode device-specific paths
- modify network configuration automatically
- reboot the target automatically
- delete user data
- add Docker unnecessarily
- add PostgreSQL/MySQL unnecessarily
- implement future phases without instruction
- claim a feature works without testing it

---

# 28. When Something Fails

If something fails:

1. Read the error.
2. Determine the root cause.
3. Inspect relevant files/environment.
4. Make the smallest reasonable fix.
5. Run the test again.
6. Report the cause and fix.

Do not randomly change multiple unrelated components.

---

# 29. When Unsure

Prefer:

```text
Inspect → reason → propose → implement
```

rather than:

```text
Guess → modify system → hope
```

If an architectural decision could significantly affect the project, explain the tradeoff before making the change.

---

# 30. Phase Completion Report

After completing each phase, report:

```text
## Phase Completed

### Implemented
- ...

### Files Changed
- ...

### Tests
- ...

### Build
- ...

### Security Considerations
- ...

### Known Issues
- ...

### Next Phase
- ...
```

Do not automatically continue to the next phase.

---

# 31. First Task

The first task is **Phase 0 only**.

Do not write the application yet.

First:

```text
1. Read PLAN.md
2. Inspect repository
3. Inspect development environment
4. Verify Go
5. Verify frontend tooling
6. Analyze target requirements
7. Produce architecture report
8. Produce Phase 0 report
```

The first response should contain:

```text
=== PROJECT UNDERSTANDING ===

=== TARGET ENVIRONMENT ===

=== ARCHITECTURE ===

=== TECHNOLOGY STACK ===

=== SECURITY MODEL ===

=== RESOURCE CONSTRAINTS ===

=== PHASE 0 FINDINGS ===

=== RISKS ===

=== NEXT STEP ===
```

Stop after Phase 0.

Wait for further instruction before implementing Phase 1.