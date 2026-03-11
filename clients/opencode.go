package clients

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mathematic-inc/mcpr/config"
)

// Path functions as variables for testing
var (
	getOpenCodeConfigPath = getOpenCodeConfigPathImpl
	getOpenCodeLocalPath  = getOpenCodeLocalPathImpl
)

func init() {
	RegisterClient(&Client{
		Name:          "opencode",
		DisplayName:   "OpenCode",
		GlobalPath:    func() (string, error) { return getOpenCodeConfigPath() },
		LocalPath:     func() (string, error) { return getOpenCodeLocalPath() },
		SupportsLocal: true,
		SyncFunc:      syncToOpenCode,
	})
}

func getOpenCodeConfigPathImpl() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "opencode", "opencode.json"), nil
}

func getOpenCodeLocalPathImpl() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, "opencode.json"), nil
}

// syncToOpenCode syncs servers to OpenCode's config format
// OpenCode uses "mcp" key with a different structure:
// - type: "local" or "remote" (instead of stdio/http)
// - command: array of strings (command + args combined)
// - environment: object (instead of env)
// - url/headers for remote servers
func syncToOpenCode(servers []config.MCPServer, path string) error {
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
		var entry map[string]any
		if server.Type == "http" {
			entry = map[string]any{
				"type": "remote",
				"url":  server.URL,
			}
			if len(server.Headers) > 0 {
				entry["headers"] = server.Headers
			}
		} else {
			// Build command array: command + args
			command := []string{server.Command}
			command = append(command, server.Args...)

			entry = map[string]any{
				"type":    "local",
				"command": command,
			}
			if len(server.Env) > 0 {
				entry["environment"] = server.Env
			}
		}
		mcpServers[server.Name] = entry
	}

	settings["mcp"] = mcpServers

	return saveSettingsFile(path, settings)
}
