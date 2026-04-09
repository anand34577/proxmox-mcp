# 🖥️ Proxmox MCP Server

[](https://go.dev/dl/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[](https://modelcontextprotocol.io)
[](https://www.proxmox.com/en/proxmox-virtual-environment)
[](#building)

A **production-grade Model Context Protocol (MCP) server** written in Go that exposes **27 tools** for full Proxmox VE administration. Connect any MCP-compatible AI client — Claude Desktop, Cursor, Continue, or your own — and manage your entire Proxmox infrastructure through natural language.

---

## ✨ Features

- 🔐 **API Token auth** — no username/password in config, zero session management
- 🛡️ **Destructive-operation guard** — stop, delete, rollback gated behind a single env flag
- 📡 **Three transports** — `stdio` (Claude Desktop), `sse`, and streamable `http`
- 📝 **Structured logging** — JSON or console output via `go.uber.org/zap`
- ⚙️ **12-factor config** — every setting is an environment variable, no config file required
- 🐳 **Scratch Docker image** — zero-dependency binary, minimal attack surface
- 🔄 **Context-aware** — every Proxmox call is cancellable via `context.Context`
- 🏗️ **Scalable architecture** — add new tool groups with a single file + one registration call

---

## 📋 Tool Reference

### Cluster (3 tools)

| Tool                    | Description                                                               |
| ----------------------- | ------------------------------------------------------------------------- |
| `get_cluster_status`    | Overall cluster health, quorum, node count                                |
| `get_cluster_resources` | All resources (VMs, containers, storage, nodes) with optional type filter |
| `get_proxmox_version`   | Proxmox VE server version string                                          |

### Nodes (3 tools)

| Tool               | Description                                       |
| ------------------ | ------------------------------------------------- |
| `list_nodes`       | All nodes with CPU, memory, and disk metrics      |
| `get_node_status`  | Detailed runtime status for a single node         |
| `get_node_storage` | All storage pools on a node with usage statistics |

### QEMU Virtual Machines (14 tools)

| Tool                   | Destructive | Description                     |
| ---------------------- |:-----------:| ------------------------------- |
| `list_vms`             |             | List all VMs on a node          |
| `get_vm_status`        |             | Runtime status of a specific VM |
| `start_vm`             |             | Power on a stopped or paused VM |
| `shutdown_vm`          |             | Graceful ACPI shutdown          |
| `reboot_vm`            |             | ACPI reboot signal              |
| `suspend_vm`           |             | Pause/freeze a running VM       |
| `resume_vm`            |             | Resume a paused VM              |
| `stop_vm`              | ✓           | Force-stop (power cord pull)    |
| `delete_vm`            | ✓           | Permanently delete VM and disks |
| `clone_vm`             |             | Full or linked clone            |
| `list_vm_snapshots`    |             | List all snapshots              |
| `create_vm_snapshot`   |             | Create a new snapshot           |
| `rollback_vm_snapshot` | ✓           | Rollback to a snapshot          |
| `delete_vm_snapshot`   | ✓           | Delete a snapshot               |

### LXC Containers (6 tools)

| Tool                   | Destructive | Description                      |
| ---------------------- |:-----------:| -------------------------------- |
| `list_containers`      |             | List all containers on a node    |
| `get_container_status` |             | Runtime status of a container    |
| `start_container`      |             | Start a stopped container        |
| `shutdown_container`   |             | Graceful init shutdown           |
| `stop_container`       | ✓           | Force-stop a container           |
| `delete_container`     | ✓           | Delete container and its storage |

### Storage (2 tools)

| Tool                 | Description                                       |
| -------------------- | ------------------------------------------------- |
| `list_node_storage`  | All storage pools with total/used/available space |
| `get_storage_detail` | Detailed info for a specific storage pool         |

---

## 🏗️ Project Structure

```
proxmox-mcp/
├── cmd/
│   └── server/
│       └── main.go              # Entry point — wiring, transport, graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go            # Viper-based config with full validation
│   ├── logger/
│   │   └── logger.go            # Structured zap logger wrapper
│   ├── proxmox/
│   │   └── client.go            # Proxmox API client wrapper (all methods)
│   └── tools/
│       ├── helpers.go           # Shared arg/response helpers + security guard
│       ├── register.go          # RegisterAll() — single entry point for all tools
│       ├── cluster.go           # Cluster tool handlers
│       ├── nodes.go             # Node tool handlers
│       ├── vms.go               # QEMU VM tool handlers
│       ├── containers.go        # LXC container tool handlers
│       └── storage.go           # Storage tool handlers
├── go.mod
├── go.sum
├── Makefile                     # Unix build targets
├── build.ps1                    # Windows PowerShell build script
├── Dockerfile                   # Multi-stage scratch image
└── README.md
```

---

## 🚀 Quick Start

### 1. Prerequisites

- **Go 1.23+** — [go.dev/dl](https://go.dev/dl/)
- **Git** — [git-scm.com](https://git-scm.com)
- A running **Proxmox VE 7.x or 8.x** instance

### 2. Create a Proxmox API Token

In the Proxmox web UI, navigate to:  
**Datacenter → Permissions → API Tokens → Add**

```
User:                    root@pam
Token ID:                mcp
Privilege Separation:    unchecked (or configure explicit ACLs below)
```

> 💡 **Least-privilege tip:** Instead of `root@pam`, create a dedicated user (`mcpuser@pam`), assign the `PVEAdmin` role, and scope it to specific pools or nodes.

### 3. Clone and Build

```bash
git clone https://github.com/yourorg/proxmox-mcp.git
cd proxmox-mcp
go mod tidy
```

**Linux / macOS:**

```bash
make build
# Binary: ./bin/proxmox-mcp
```

**Windows (PowerShell):**

```powershell
# Allow scripts (run once in admin PowerShell)
Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy RemoteSigned

.\build.ps1 build
# Binary: .\bin\proxmox-mcp.exe
```

**Windows (CMD):**

```cmd
set CGO_ENABLED=0
go build -ldflags="-s -w" -o .\bin\proxmox-mcp.exe .\cmd\server
```

### 4. Run the Server

```bash
# Linux / macOS
export PROXMOX_MCP_PROXMOX_BASE_URL=https://192.168.1.10:8006/api2/json
export PROXMOX_MCP_PROXMOX_TOKEN_ID=root@pam!mcp
export PROXMOX_MCP_PROXMOX_SECRET=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
export PROXMOX_MCP_PROXMOX_INSECURE_SKIP_VERIFY=true   # only for self-signed certs
./bin/proxmox-mcp
```

```powershell
# Windows PowerShell
$env:PROXMOX_MCP_PROXMOX_BASE_URL             = "https://192.168.1.10:8006/api2/json"
$env:PROXMOX_MCP_PROXMOX_TOKEN_ID             = "root@pam!mcp"
$env:PROXMOX_MCP_PROXMOX_SECRET               = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
$env:PROXMOX_MCP_PROXMOX_INSECURE_SKIP_VERIFY = "true"
.\bin\proxmox-mcp.exe
```

---

## ⚙️ Configuration

All configuration is done via **environment variables**. No config file is required.

| Environment Variable                           | Default                            | Required | Description                                  |
| ---------------------------------------------- | ---------------------------------- |:--------:| -------------------------------------------- |
| `PROXMOX_MCP_PROXMOX_BASE_URL`                 | `https://localhost:8006/api2/json` | ✓        | Full Proxmox API URL                         |
| `PROXMOX_MCP_PROXMOX_TOKEN_ID`                 | —                                  | ✓        | API token ID (`user@realm!tokenname`)        |
| `PROXMOX_MCP_PROXMOX_SECRET`                   | —                                  | ✓        | API token secret UUID                        |
| `PROXMOX_MCP_PROXMOX_INSECURE_SKIP_VERIFY`     | `false`                            |          | Skip TLS verification (dev/self-signed only) |
| `PROXMOX_MCP_PROXMOX_REQUEST_TIMEOUT`          | `30s`                              |          | Per-request HTTP timeout                     |
| `PROXMOX_MCP_SERVER_TRANSPORT`                 | `stdio`                            |          | `stdio` \| `sse` \| `http`                   |
| `PROXMOX_MCP_SERVER_LISTEN_ADDR`               | `:8080`                            |          | Bind address for SSE/HTTP transports         |
| `PROXMOX_MCP_SERVER_NAME`                      | `Proxmox MCP Server`               |          | Advertised server name                       |
| `PROXMOX_MCP_SERVER_VERSION`                   | `1.0.0`                            |          | Advertised version string                    |
| `PROXMOX_MCP_SERVER_GRACEFUL_SHUTDOWN_TIMEOUT` | `30s`                              |          | Shutdown wait time                           |
| `PROXMOX_MCP_SECURITY_ALLOW_DESTRUCTIVE`       | `false`                            |          | Enable destructive operations                |
| `PROXMOX_MCP_SECURITY_ALLOWED_NODES`           | _(all)_                            |          | Comma-separated node whitelist               |
| `PROXMOX_MCP_LOG_LEVEL`                        | `info`                             |          | `debug` \| `info` \| `warn` \| `error`       |
| `PROXMOX_MCP_LOG_FORMAT`                       | `json`                             |          | `json` \| `console`                          |
| `PROXMOX_MCP_CONFIG_FILE`                      | —                                  |          | Optional path to YAML/TOML/JSON config file  |

### Using a Config File (Optional)

```yaml
# proxmox-mcp.yaml
server:
  transport: sse
  listen_addr: ":8080"

proxmox:
  base_url: "https://192.168.1.10:8006/api2/json"
  token_id: "root@pam!mcp"
  secret: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  insecure_skip_verify: true
  request_timeout: 30s

security:
  allow_destructive: false

log:
  level: info
  format: console
```

```bash
PROXMOX_MCP_CONFIG_FILE=./proxmox-mcp.yaml ./bin/proxmox-mcp
```

---

## 🔌 Client Integration

### Claude Desktop

Edit `~/.config/claude/claude_desktop_config.json` (Linux/macOS):

```json
{
  "mcpServers": {
    "proxmox": {
      "command": "/absolute/path/to/proxmox-mcp",
      "env": {
        "PROXMOX_MCP_PROXMOX_BASE_URL": "https://192.168.1.10:8006/api2/json",
        "PROXMOX_MCP_PROXMOX_TOKEN_ID": "root@pam!mcp",
        "PROXMOX_MCP_PROXMOX_SECRET": "your-secret-uuid-here",
        "PROXMOX_MCP_PROXMOX_INSECURE_SKIP_VERIFY": "true",
        "PROXMOX_MCP_LOG_FORMAT": "console"
      }
    }
  }
}
```

**Windows path** — edit `%APPDATA%\Claude\claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "proxmox": {
      "command": "C:\\path\\to\\proxmox-mcp\\bin\\proxmox-mcp.exe",
      "env": {
        "PROXMOX_MCP_PROXMOX_BASE_URL": "https://192.168.1.10:8006/api2/json",
        "PROXMOX_MCP_PROXMOX_TOKEN_ID": "root@pam!mcp",
        "PROXMOX_MCP_PROXMOX_SECRET": "your-secret-uuid-here",
        "PROXMOX_MCP_PROXMOX_INSECURE_SKIP_VERIFY": "true"
      }
    }
  }
}
```

### Cursor / Continue

Point the MCP server config to the binary path and supply the same environment variables as above under your IDE's MCP server settings.

### SSE / HTTP Mode (Remote Access)

Start the server in `sse` or `http` transport for network-accessible deployments:

```bash
PROXMOX_MCP_SERVER_TRANSPORT=sse \
PROXMOX_MCP_SERVER_LISTEN_ADDR=:8080 \
PROXMOX_MCP_PROXMOX_BASE_URL=https://192.168.1.10:8006/api2/json \
PROXMOX_MCP_PROXMOX_TOKEN_ID=root@pam!mcp \
PROXMOX_MCP_PROXMOX_SECRET=your-secret-uuid \
./bin/proxmox-mcp
```

Then configure your MCP client to connect to `http://your-server:8080`.

---

## 🐳 Docker

### Build

```bash
docker build -t proxmox-mcp:latest .
```

### Run (SSE mode)

```bash
docker run -d \
  --name proxmox-mcp \
  -p 8080:8080 \
  -e PROXMOX_MCP_SERVER_TRANSPORT=sse \
  -e PROXMOX_MCP_SERVER_LISTEN_ADDR=:8080 \
  -e PROXMOX_MCP_PROXMOX_BASE_URL=https://192.168.1.10:8006/api2/json \
  -e PROXMOX_MCP_PROXMOX_TOKEN_ID=root@pam!mcp \
  -e PROXMOX_MCP_PROXMOX_SECRET=your-secret-uuid \
  -e PROXMOX_MCP_PROXMOX_INSECURE_SKIP_VERIFY=true \
  proxmox-mcp:latest
```

### Docker Compose

```yaml
version: "3.9"
services:
  proxmox-mcp:
    image: proxmox-mcp:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      PROXMOX_MCP_SERVER_TRANSPORT: sse
      PROXMOX_MCP_SERVER_LISTEN_ADDR: ":8080"
      PROXMOX_MCP_PROXMOX_BASE_URL: "https://192.168.1.10:8006/api2/json"
      PROXMOX_MCP_PROXMOX_TOKEN_ID: "root@pam!mcp"
      PROXMOX_MCP_PROXMOX_SECRET: "your-secret-uuid"
      PROXMOX_MCP_PROXMOX_INSECURE_SKIP_VERIFY: "true"
      PROXMOX_MCP_LOG_FORMAT: "json"
```

---

## 🔒 Security

### Destructive Operations

All destructive operations are **disabled by default**. They require an explicit opt-in:

```bash
PROXMOX_MCP_SECURITY_ALLOW_DESTRUCTIVE=true
```

The following tools are gated behind this flag:

- `stop_vm` (force power-off)
- `delete_vm`
- `rollback_vm_snapshot`
- `delete_vm_snapshot`
- `stop_container` (force)
- `delete_container`

### Node Allowlist

Restrict operations to specific nodes:

```bash
PROXMOX_MCP_SECURITY_ALLOWED_NODES=pve01,pve02
```

### API Token Best Practices

1. **Never use root@pam in production.** Create a dedicated `mcpuser@pam` account.
2. Assign the minimum required role — `PVEAdmin` scoped to specific pools, or custom roles using just the ACLs you need (e.g. `VM.PowerMgmt`, `VM.Snapshot`).
3. Rotate the token secret periodically.
4. Do not pass secrets via command-line arguments — use environment variables or a secrets manager.

### TLS

Set `PROXMOX_MCP_PROXMOX_INSECURE_SKIP_VERIFY=false` (the default) and install a valid certificate on your Proxmox node in production. Let's Encrypt via ACME is supported natively by Proxmox VE.

---

## 🛠️ Development

### Prerequisites

```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Makefile Targets (Linux/macOS)

```bash
make build     # Compile binary to ./bin/
make test      # Run tests with race detector
make lint      # Run golangci-lint
make clean     # Remove ./bin/
make all       # lint + test + build
```

### PowerShell Targets (Windows)

```powershell
.\build.ps1 build   # Compile binary to .\bin\
.\build.ps1 test    # Run tests with race detector
.\build.ps1 lint    # Run golangci-lint
.\build.ps1 clean   # Remove .\bin\
.\build.ps1 all     # lint + test + build
```

### Adding a New Tool Group

1. Create `internal/tools/mygroup.go` with a `RegisterMyGroupTools(s, h)` function
2. Add `RegisterMyGroupTools(s, h)` to `internal/tools/register.go`
3. Add corresponding methods to `internal/proxmox/client.go`

That's it — no other wiring required.

### Adding a New Proxmox API Method

1. Add the method to `internal/proxmox/client.go` following the existing pattern
2. Add the tool handler in the appropriate `internal/tools/*.go` file
3. Register the tool inside the corresponding `Register*Tools()` function

---

## 🔭 Observability

### Log Levels

```bash
PROXMOX_MCP_LOG_LEVEL=debug   # All requests + arg tracing
PROXMOX_MCP_LOG_LEVEL=info    # Lifecycle events + tool invocations (default)
PROXMOX_MCP_LOG_LEVEL=warn    # Destructive operations + warnings only
PROXMOX_MCP_LOG_LEVEL=error   # Errors only
```

### Log Formats

```bash
PROXMOX_MCP_LOG_FORMAT=json     # Structured JSON (default, ideal for log aggregators)
PROXMOX_MCP_LOG_FORMAT=console  # Human-readable coloured output for local dev
```

### Sample Log Output (console format)

```
2026-04-09T00:30:01.123+0530  INFO  starting Proxmox MCP Server  {"name": "Proxmox MCP Server", "transport": "stdio"}
2026-04-09T00:30:01.456+0530  INFO  connected to Proxmox VE      {"url": "https://192.168.1.10:8006/api2/json"}
2026-04-09T00:30:01.457+0530  INFO  tools registered             {"allow_destructive": false}
2026-04-09T00:30:05.012+0530  INFO  tool: list_nodes
2026-04-09T00:30:06.234+0530  INFO  tool: start_vm               {"node": "pve01", "vmid": 101}
2026-04-09T00:30:10.789+0530  WARN  tool: stop_vm (force)        {"node": "pve01", "vmid": 101}
```

---

## 📦 Dependencies

| Package                                                                 | Version   | Purpose                  |
| ----------------------------------------------------------------------- | --------- | ------------------------ |
| [`luthermonson/go-proxmox`](https://github.com/luthermonson/go-proxmox) | `v0.2.0`  | Proxmox VE API client    |
| [`mark3labs/mcp-go`](https://github.com/mark3labs/mcp-go)               | `v0.32.0` | MCP server framework     |
| [`spf13/viper`](https://github.com/spf13/viper)                         | `v1.19.0` | Configuration management |
| [`go.uber.org/zap`](https://pkg.go.dev/go.uber.org/zap)                 | `v1.27.0` | Structured logging       |

---

## 🗺️ Roadmap

- [ ] **VM Config editor** — update CPU, memory, network via `update_vm_config`
- [ ] **Backup management** — trigger vzdump backups and list backup jobs
- [ ] **User & ACL management** — create users, assign roles, manage permissions
- [ ] **Network SDN** — manage VNets, zones, and VLAN mappings
- [ ] **Firewall rules** — manage node and VM-level firewall rules
- [ ] **ISO / Template management** — download ISOs, list templates
- [ ] **HA management** — configure and query HA groups and resources
- [ ] **Prometheus metrics endpoint** — expose tool call latency and error rates

---

## 🐛 Troubleshooting

| Error                                           | Cause                     | Fix                                                            |
| ----------------------------------------------- | ------------------------- | -------------------------------------------------------------- |
| `connecting to ... : unauthorized`              | Wrong token ID or secret  | Double-check `TOKEN_ID` format: `user@realm!tokenname`         |
| `x509: certificate signed by unknown authority` | Self-signed Proxmox cert  | Set `PROXMOX_MCP_PROXMOX_INSECURE_SKIP_VERIFY=true` (dev only) |
| `operation "delete_vm" is disabled`             | Destructive guard         | Set `PROXMOX_MCP_SECURITY_ALLOW_DESTRUCTIVE=true`              |
| `go: command not found` (Windows)               | Go not in PATH            | Restart PowerShell after installing Go                         |
| `go mod tidy` fails first run                   | Module cache race         | Run `go mod tidy` a second time                                |
| `Access denied` on `build.ps1`                  | Execution policy          | Run `Set-ExecutionPolicy -Scope CurrentUser RemoteSigned`      |
| Server starts but no tools appear               | Wrong transport in client | Match client URL scheme to `PROXMOX_MCP_SERVER_TRANSPORT`      |

---

## 🤝 Contributing

Contributions are welcome. Please follow these steps:

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Write your code following the existing patterns in `internal/tools/`
4. Run `make all` (or `.\build.ps1 all`) — lint, test, and build must all pass
5. Commit with a conventional commit message: `feat: add backup management tools`
6. Open a Pull Request against `main`

---

## 📄 License

MIT License — see [LICENSE](LICENSE) for details.

---

## 🙏 Acknowledgements

- [luthermonson/go-proxmox](https://github.com/luthermonson/go-proxmox) — the excellent Go Proxmox API client this server is built on
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) — the MCP server framework
- [Proxmox VE](https://www.proxmox.com/en/proxmox-virtual-environment) — the virtualisation platform being managed
