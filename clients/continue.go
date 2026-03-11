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
	getContinueConfigPath = getContinueConfigPathImpl
)

func init() {
	RegisterClient(&Client{
		Name:          "continue",
		DisplayName:   "Continue",
		GlobalPath:    func() (string, error) { return getContinueConfigPath() },
		LocalPath:     nil,
		SupportsLocal: false,
		SyncFunc:      syncToContinue,
	})
}

func getContinueConfigPathImpl() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".continue", "config.json"), nil
}

func syncToContinue(servers []config.MCPServer, path string) error {
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

	// Continue uses "mcpServers" array with transport config
	mcpServers := make([]map[string]any, 0, len(servers))
	for _, server := range servers {
		var transport map[string]any
		if server.Type == "http" {
			transport = map[string]any{
				"type": "sse",
				"url":  server.URL,
			}
			if len(server.Headers) > 0 {
				transport["headers"] = server.Headers
			}
		} else {
			transport = map[string]any{
				"type":    "stdio",
				"command": server.Command,
			}
			if len(server.Args) > 0 {
				transport["args"] = server.Args
			}
			if len(server.Env) > 0 {
				transport["env"] = server.Env
			}
		}

		mcpServers = append(mcpServers, map[string]any{
			"name":      server.Name,
			"transport": transport,
		})
	}

	settings["mcpServers"] = mcpServers

	return saveSettingsFile(path, settings)
}
