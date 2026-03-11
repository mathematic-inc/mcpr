# MCPR - MCP Registry

A CLI tool for managing Model Context Protocol (MCP) servers across multiple AI-powered code editors and development tools.

## Overview

MCPR provides a centralized way to configure and synchronize MCP server configurations to various client applications that support the MCP standard. Instead of manually editing configuration files for each tool, define your MCP servers once and sync them everywhere.

## Features

- **Centralized Configuration** - Define MCP servers in one place, sync to all your tools
- **Multi-Client Support** - Works with Claude Desktop, Cursor, VS Code, Windsurf, Zed, and more
- **Stdio & HTTP Servers** - Support for both command-line and HTTP/SSE-based MCP servers
- **Project-Local Configs** - Global or per-project server configurations
- **Shell Completion** - Tab completion for commands and server names
- **Atomic Syncing** - Automatic sync to all clients after configuration changes

## Installation

```bash
go install github.com/mathematic-inc/mcpr@latest
```

Or build from source:

```bash
git clone https://github.com/mathematic-inc/mcpr.git
cd mcpr
go build -o mcpr .
```

## Quick Start

```bash
# Add a stdio-based MCP server
mcpr add stdio npx -y @modelcontextprotocol/server-filesystem /path/to/dir

# Add an HTTP-based MCP server
mcpr add http https://example.com/mcp --name my-server

# Sync to a client
mcpr client sync claude-desktop

# List configured servers
mcpr list

# List supported clients
mcpr list --clients
```

## Commands

### `mcpr add`

Add new MCP server configurations.

#### `mcpr add stdio [command] [args...]`

Add a stdio-based MCP server that communicates via stdin/stdout.

```bash
# Basic usage
mcpr add stdio npx -y @modelcontextprotocol/server-filesystem /home/user

# With custom name
mcpr add stdio --name filesystem npx -y @modelcontextprotocol/server-filesystem /home/user

# With environment variables
mcpr add stdio --env API_KEY=secret --env DEBUG=true npx my-server

# Add to local project config
mcpr add stdio --local npx my-project-server
```

**Flags:**
- `--name, -n` - Custom name for the server (defaults to command name)
- `--env, -e` - Environment variables in KEY=VALUE format (repeatable)
- `--local, -l` - Add to local project configuration

#### `mcpr add http [url]`

Add an HTTP/SSE-based MCP server.

```bash
# Basic usage
mcpr add http https://example.com/mcp

# With custom name and headers
mcpr add http --name my-api --header "Authorization=Bearer token" https://api.example.com/mcp
```

**Flags:**
- `--name, -n` - Custom name for the server (defaults to URL host)
- `--header, -H` - HTTP headers in Key=Value format (repeatable)
- `--local, -l` - Add to local project configuration

### `mcpr remove`

Remove an MCP server from configuration. Alias: `rm`

```bash
mcpr remove my-server
mcpr rm my-server
```

After removal, all synced clients are automatically updated.

### `mcpr client`

Manage client synchronization.

#### `mcpr client sync [client-name]`

Sync MCP servers to a client application.

```bash
# Sync all servers to a specific client
mcpr client sync claude-desktop

# Resync all previously synced clients
mcpr client sync

# Sync specific servers only
mcpr client sync cursor --servers server1,server2

# Sync to local client config
mcpr client sync claude-code --local
```

**Flags:**
- `--servers, -s` - Comma-separated list of specific servers to sync
- `--local, -l` - Use local client configuration

#### `mcpr client remove [client-name]`

Remove a client from the sync list.

```bash
mcpr client remove cursor
mcpr client remove claude-code --local
```

**Flags:**
- `--local, -l` - Remove from local configuration

### `mcpr list`

Display configured items.

```bash
# List all configured servers
mcpr list

# List all supported clients
mcpr list --clients
mcpr list -c
```

**Flags:**
- `--clients, -c` - List supported clients instead of servers

## Supported Clients

| Client | Description | Local Config Support |
|--------|-------------|---------------------|
| `claude-desktop` | Claude Desktop app | No |
| `claude-code` | Claude Code CLI | Yes |
| `cursor` | Cursor editor | Yes |
| `windsurf` | Windsurf editor | Yes |
| `zed` | Zed editor | No |
| `vscode` | VS Code with GitHub Copilot | Yes |
| `continue` | Continue AI extension | No |
| `cline` | Cline VS Code extension | No |
| `codex` | OpenAI Codex CLI | No |
| `gemini` | Google Gemini CLI | No |
| `kilocode` | Kilo Code VS Code extension | No |
| `zencoder` | ZenCoder VS Code extension | No |

## Configuration

### File Locations

- **Global config:** `~/.config/mcpr/config.json`
- **Local config:** `mcpr.json` in project directory (or parent directories)

### Configuration Structure

```json
{
  "servers": [
    {
      "name": "filesystem",
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user"],
      "env": {}
    },
    {
      "name": "my-api",
      "type": "http",
      "url": "https://api.example.com/mcp",
      "headers": {
        "Authorization": "Bearer token"
      }
    }
  ],
  "synced_clients": [
    {
      "name": "claude-desktop",
      "local": false,
      "servers": null
    },
    {
      "name": "cursor",
      "local": false,
      "servers": ["filesystem"]
    }
  ]
}
```

### Server Types

#### Stdio Servers

Stdio servers communicate via standard input/output streams:

```json
{
  "name": "my-server",
  "type": "stdio",
  "command": "npx",
  "args": ["-y", "my-mcp-server"],
  "env": {
    "API_KEY": "secret"
  }
}
```

#### HTTP Servers

HTTP servers communicate via HTTP/SSE:

```json
{
  "name": "my-api",
  "type": "http",
  "url": "https://api.example.com/mcp",
  "headers": {
    "Authorization": "Bearer token"
  }
}
```

## Examples

### Setting Up a Development Environment

```bash
# Add common MCP servers
mcpr add stdio --name filesystem npx -y @modelcontextprotocol/server-filesystem ~/projects
mcpr add stdio --name git npx -y @modelcontextprotocol/server-git
mcpr add stdio --name github npx -y @modelcontextprotocol/server-github --env GITHUB_TOKEN=ghp_xxx

# Sync to all your tools
mcpr client sync claude-desktop
mcpr client sync cursor
mcpr client sync vscode
```

### Project-Specific Configuration

```bash
# In your project directory
cd ~/projects/my-app

# Add project-specific server
mcpr add stdio --local --name my-app-server ./scripts/mcp-server.js

# Sync to Claude Code's local config
mcpr client sync claude-code --local
```

### Syncing Specific Servers

```bash
# Only sync certain servers to a client
mcpr client sync cursor --servers filesystem,git
```

## How It Works

1. **Add servers** to MCPR's central configuration
2. **Sync clients** to copy server configs to each tool's native format
3. **Auto-resync** happens when you add or remove servers

MCPR reads each client's existing configuration and updates only the MCP server sections, preserving all other settings.

## License

Apache-2.0

> This project is free and open-source work by a 501(c)(3) non-profit. If you find it useful, please consider [donating](https://github.com/sponsors/mathematic-inc).
