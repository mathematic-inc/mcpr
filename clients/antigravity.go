package clients

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jrandolf/mcpr/config"
)

// Path functions as variables for testing
var (
	getAntigravityConfigPath = getAntigravityConfigPathImpl
	getAntigravityLocalPath  = getAntigravityLocalPathImpl
)

func init() {
	RegisterClient(&Client{
		Name:          "antigravity",
		DisplayName:   "Antigravity (Google)",
		GlobalPath:    func() (string, error) { return getAntigravityConfigPath() },
		LocalPath:     func() (string, error) { return getAntigravityLocalPath() },
		SupportsLocal: true,
		SyncFunc:      syncToAntigravityConfig,
	})
}

func getAntigravityConfigPathImpl() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gemini", "antigravity", "mcp_config.json"), nil
}

func getAntigravityLocalPathImpl() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, ".antigravity", "mcp_config.json"), nil
}

// AntigravityConfig represents the config format for Antigravity
type AntigravityConfig struct {
	MCPServers map[string]AntigravityServerEntry `json:"mcpServers"`
}

type AntigravityServerEntry struct {
	Command   string            `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	ServerURL string            `json:"serverUrl,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
}

func syncToAntigravityConfig(servers []config.MCPServer, path string) error {
	cfg := &AntigravityConfig{
		MCPServers: make(map[string]AntigravityServerEntry),
	}

	for _, server := range servers {
		entry := AntigravityServerEntry{}
		if server.Type == "http" {
			entry.ServerURL = server.URL
			entry.Headers = server.Headers
		} else {
			entry.Command = server.Command
			entry.Args = server.Args
			entry.Env = server.Env
		}
		cfg.MCPServers[server.Name] = entry
	}

	return saveAntigravityConfig(path, cfg)
}

func saveAntigravityConfig(path string, cfg *AntigravityConfig) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
