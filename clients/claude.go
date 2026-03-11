package clients

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mathematic-inc/mcpr/config"
)

// Path functions as variables for testing
var (
	getClaudeDesktopConfigPath = getClaudeDesktopConfigPathImpl
	getClaudeCodeConfigPath    = getClaudeCodeConfigPathImpl
	getClaudeCodeLocalPath     = getClaudeCodeLocalPathImpl
)

func init() {
	RegisterClient(&Client{
		Name:          "claude-desktop",
		DisplayName:   "Claude Desktop",
		GlobalPath:    func() (string, error) { return getClaudeDesktopConfigPath() },
		LocalPath:     nil,
		SupportsLocal: false,
		SyncFunc:      syncToMCPConfig,
	})

	RegisterClient(&Client{
		Name:          "claude-code",
		DisplayName:   "Claude Code",
		GlobalPath:    func() (string, error) { return getClaudeCodeConfigPath() },
		LocalPath:     func() (string, error) { return getClaudeCodeLocalPath() },
		SupportsLocal: true,
		SyncFunc:      syncToClaudeCode,
	})
}

func getClaudeDesktopConfigPathImpl() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Claude", "claude_desktop_config.json"), nil
	case "linux":
		return filepath.Join(home, ".config", "Claude", "claude_desktop_config.json"), nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func getClaudeCodeConfigPathImpl() (string, error) {
	if claudeConfigDir := os.Getenv("CLAUDE_CONFIG_DIR"); claudeConfigDir != "" {
		return filepath.Join(claudeConfigDir, "claude.json"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".claude.json"), nil
}

func getClaudeCodeLocalPathImpl() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, ".mcp.json"), nil
}

func syncToClaudeCode(servers []config.MCPServer, path string) error {
	var settings map[string]any
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		settings = make(map[string]any)
	} else if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	} else {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}
	}

	mcpServers := make(map[string]any)
	for _, server := range servers {
		entry := make(map[string]any)
		if server.Type == "http" {
			entry["type"] = "http"
			entry["url"] = server.URL
			if len(server.Headers) > 0 {
				entry["headers"] = server.Headers
			}
		} else {
			entry["type"] = "stdio"
			entry["command"] = server.Command
			if len(server.Args) > 0 {
				entry["args"] = server.Args
			}
			if len(server.Env) > 0 {
				entry["env"] = server.Env
			}
		}
		mcpServers[server.Name] = entry
	}

	settings["mcpServers"] = mcpServers

	return saveSettingsFile(path, settings)
}
