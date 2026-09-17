# PMOS Control Panel — Project Plan

## 1. Project Overview

PMOS Control Panel adalah web-based management panel yang dibuat khusus untuk perangkat yang menjalankan postmarketOS.

Perangkat target adalah HP bekas yang digunakan sebagai mini Linux server.

Tujuan:

> Mengontrol, memonitor, dan mengelola HP postmarketOS melalui browser dari perangkat lain di jaringan.

Contoh:

```text
Laptop / PC
     │
     │ Browser
     ▼
http://192.168.x.x:8080
     │
     ▼
┌──────────────────────────┐
│     PMOS CONTROL PANEL   │
├──────────────────────────┤
│ Dashboard                │
│ Terminal                 │
│ Files                    │
│ Processes                │
│ Services                 │
│ Logs                     │
│ Network                  │
│ System                   │
│ Power                    │
│ Settings                 │
└──────────────────────────┘
     │
     ▼
 postmarketOS
```

Project ini dibuat sendiri dan **bukan fork Cockpit atau Webmin**.

---

# 2. Target Environment

Target utama:

- OS: postmarketOS
- Channel: stable
- Vendor: qcom
- Init system: systemd
- UI: fbkeyboard
- Architecture: ARM/ARM64, harus diverifikasi

Catatan:

`fbkeyboard` adalah UI/environment pada perangkat dan bukan dependency utama web panel.

Web panel harus dapat berjalan sebagai background service tanpa bergantung pada desktop environment.

Jangan mengasumsikan OpenRC.

Jangan mengasumsikan systemd desktop environment.

---

# 3. Project Goals

Panel harus dapat:

### Monitoring

- CPU
- RAM
- Storage
- Temperature
- Uptime
- Load average
- Network traffic
- Processes
- System information

### Management

- Systemd services
- Processes
- Files
- Logs
- Network information
- Terminal

### Control

- Reboot
- Shutdown

---

# 4. Technology Stack

## Backend

Gunakan:

```text
Go
```

Alasan:

- low runtime overhead
- cocok untuk ARM
- mudah cross-compile
- deployment sederhana
- single binary
- concurrency bagus
- cocok untuk HTTP/WebSocket

## HTTP

Prioritaskan standard library:

```text
net/http
```

Framework hanya digunakan jika memang diperlukan.

## API

REST API.

## Realtime

WebSocket untuk:

- system monitoring
- terminal
- live logs

## Terminal

Linux PTY.

## Frontend

Prioritas:

```text
Svelte
TypeScript
CSS
```

Frontend harus ringan.

## Database

Tidak diperlukan untuk MVP.

Jika persistence benar-benar diperlukan:

```text
SQLite
```

Jangan menambahkan database server seperti PostgreSQL/MySQL untuk MVP.

---

# 5. Final Architecture

```text
                         Browser
                            │
                    HTTP / WebSocket
                            │
                            ▼
              ┌────────────────────────┐
              │       Go Backend       │
              │                        │
              │ HTTP Server             │
              │ REST API                │
              │ WebSocket               │
              │ Authentication          │
              │ Authorization           │
              └───────────┬────────────┘
                          │
             ┌────────────┼─────────────┐
             │            │             │
             ▼            ▼             ▼
         Monitoring    systemd        Terminal
             │          Manager          │
             ▼            ▼              ▼
           /proc      systemctl         PTY
           /sys       journalctl        shell
             │            │
             └────────────┼─────────────┘
                          ▼
                    postmarketOS
```

---

# 6. Application Structure

Navigation:

```text
Dashboard
Terminal
Files
Processes
Services
Logs
Network
System
Power
Settings
```

---

# 7. Dashboard

Dashboard harus menjadi halaman utama.

Tampilkan:

```text
System status
CPU
RAM
Storage
Temperature
Uptime
Load average
Network
```

Contoh:

```text
┌───────────────────────────────────────┐
│ PMOS CONTROL PANEL                   │
├───────────────────────────────────────┤
│                                       │
│ ● ONLINE                              │
│                                       │
│ CPU        23%                        │
│ RAM        612 MB / 2 GB              │
│ STORAGE    8.2 GB / 32 GB             │
│ TEMP       42°C                       │
│ UPTIME     2d 14h                     │
│                                       │
│ CPU HISTORY                           │
│ RAM HISTORY                           │
│ NETWORK HISTORY                       │
└───────────────────────────────────────┘
```

Monitoring harus hemat resource.

---

# 8. System Monitoring

Gunakan Linux interfaces jika memungkinkan.

CPU:

```text
/proc/stat
```

Memory:

```text
/proc/meminfo
```

Storage:

```text
statfs
```

Temperature:

```text
/sys/class/thermal/
```

Network:

```text
/proc/net/dev
```

Jangan hardcode:

- thermal zone
- network interface
- CPU model
- architecture

Discovery dilakukan secara dinamis.

---

# 9. Process Manager

Tampilkan:

```text
PID
USER
CPU
RAM
COMMAND
```

Fitur:

- Search
- Sort
- Refresh
- Kill process

Kill process harus memiliki confirmation.

User tidak boleh membunuh process yang tidak memiliki permission.

---

# 10. Systemd Manager

Target menggunakan systemd.

Fitur:

```text
List services
Status
Start
Stop
Restart
Enable
Disable
Logs
```

Backend menggunakan systemd API/commands yang sesuai.

Frontend tidak boleh menjalankan `systemctl` secara langsung.

Semua operasi melalui backend service layer.

---

# 11. Web Terminal

Terminal adalah fitur utama.

Architecture:

```text
Browser
   │
WebSocket
   │
Go backend
   │
PTY
   │
Shell
```

Harus mendukung:

- stdin
- stdout
- stderr
- resize
- interactive shell
- disconnect handling

Terminal harus menggunakan authentication.

Jangan membuat arbitrary command API yang tidak terproteksi.

---

# 12. File Manager

Default filesystem root:

```text
/home/<panel-user>
```

Fitur:

- Browse
- Open
- Upload
- Download
- Rename
- Delete
- Create folder

Security:

- canonical path
- path traversal protection
- symlink escape protection
- configurable root directory

Request seperti:

```text
../../etc/passwd
```

harus ditolak.

---

# 13. Logs

Karena menggunakan systemd:

```text
journalctl
```

Fitur:

- Recent logs
- Live logs
- Search
- Filter
- Service logs
- Kernel logs

Live logs menggunakan WebSocket.

---

# 14. Network

MVP hanya read-only.

Tampilkan:

- Interface
- Status
- IPv4
- IPv6
- MAC
- RX
- TX

Jangan membuat network configuration editor pada MVP karena dapat menyebabkan server kehilangan koneksi.

---

# 15. Power Management

Fitur:

```text
Reboot
Shutdown
```

Gunakan systemd.

Semua operasi:

- authentication
- authorization
- POST request
- confirmation

Jangan menggunakan GET untuk operasi destructive.

---

# 16. Authentication

Buat login system.

Minimal:

```text
Username
Password
Login
Logout
Session
```

Password harus menggunakan secure password hashing.

Session menggunakan secure cookie.

Cookie harus:

```text
HttpOnly
SameSite
Secure jika HTTPS
```

Session memiliki expiration.

---

# 17. Security Model

Semua request:

```text
Browser
   │
Authentication
   │
Authorization
   │
API
   │
Backend
   │
Linux
```

Perhatikan:

- Command injection
- Path traversal
- XSS
- CSRF
- Session hijacking
- Unauthorized WebSocket
- Privilege escalation
- Malicious file upload
- Brute force

---

# 18. Privilege Model

Backend tidak dijalankan sebagai root secara default.

Model:

```text
pmos-panel
     │
     ▼
normal Linux user
```

Operasi privileged menggunakan mekanisme privilege escalation yang terbatas.

Jangan memberikan:

```text
sudo ALL=(ALL) NOPASSWD: ALL
```

---

# 19. Deployment

Target deployment:

```text
pmos-panel
```

Satu executable Go.

Frontend di-embed ke binary jika memungkinkan.

Node.js tidak diperlukan pada runtime HP.

Target:

```text
./pmos-panel
```

langsung menjalankan:

- HTTP server
- REST API
- WebSocket
- frontend

---

# 20. Systemd Service

Buat:

```text
pmos-panel.service
```

Lokasi:

```text
/etc/systemd/system/pmos-panel.service
```

Service harus:

- start otomatis saat boot
- restart ketika crash
- menjalankan non-root user
- menggunakan journald
- menunggu network
- melakukan graceful shutdown

---

# 21. Development Phases

## Phase 0 — Discovery ✅ DONE

Jangan coding.

Inspect:

```text
Repository
Target architecture
OS
Kernel
systemd
CPU
RAM
Storage
Network
Thermal sensors
Available tools
```

Output:

```text
ENVIRONMENT REPORT
```

---

## Phase 1 — Foundation ✅ DONE

Implement:

- Go backend
- frontend
- HTTP server
- static frontend
- authentication
- configuration
- logging

Target:

```text
Browser
  ↓
Login
  ↓
Dashboard
```

---

## Phase 2 — Monitoring ✅ DONE

Implement:

- CPU
- RAM
- Storage
- Temperature
- Network
- Load
- Uptime

Gunakan WebSocket untuk realtime data.

---

## Phase 3 — Terminal ✅ DONE

Implement:

- WebSocket
- PTY
- Shell
- resize
- stdin/stdout
- authentication

---

## Phase 4 — Systemd ✅ DONE

Implement:

- service list
- status
- start
- stop
- restart
- enable
- disable
- logs

---

## Phase 5 — Processes ✅ DONE

Implement:

- process list
- CPU
- RAM
- search
- sort
- kill

---

## Phase 6 — Files ✅ DONE

Implement:

- browse
- upload
- download
- rename
- delete
- mkdir

Dengan filesystem sandbox.

---

## Phase 7 — Logs ✅ DONE

Implement:

- journal
- service logs
- live logs (WebSocket)
- search
- filter

---

## Phase 8 — Network ✅ DONE

Implement:

- interface information
- IP
- MAC
- RX
- TX
- status

Read-only.

---

## Phase 9 — Power ✅ DONE

Implement:

- reboot
- shutdown

Dengan authentication dan confirmation.

---

## Phase 10 — Production Deployment ✅ DONE

Implement:

- embedded frontend
- single binary
- configuration
- systemd service
- auto-start
- graceful shutdown
- logging

---

## Phase 11 — Security Audit 🔲 TODO

Audit:

```text
Authentication
Authorization
Session
CSRF
XSS
Command injection
Path traversal
Symlink escape
File upload
WebSocket
Privilege escalation
Rate limiting
Secret handling
```

---

## Phase 12 — Optimization 🔲 TODO

Measure:

```text
Idle RAM
Idle CPU
Startup time
Dashboard performance
WebSocket performance
Terminal performance
```

Target:

- low RAM
- low CPU
- low storage
- fast startup
- minimal dependencies

---

# 22. Definition of Done

Project selesai ketika:

1. HP boot.
2. systemd menjalankan `pmos-panel`.
3. HP terhubung ke network.
4. Browser laptop membuka:

```text
http://IP-HP:8080
```

5. User dapat login.
6. Dashboard tampil.
7. CPU/RAM/storage/temperature tampil.
8. Terminal dapat digunakan.
9. Process dapat dilihat.
10. Systemd service dapat dikelola.
11. Journal dapat dilihat.
12. File dapat dikelola.
13. Network dapat dimonitor.
14. HP dapat reboot/shutdown dari panel.

Semua harus berjalan pada postmarketOS ARM dengan resource terbatas.