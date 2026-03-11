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
	getZedConfigPath = getZedConfigPathImpl
)

func init() {
	RegisterClient(&Client{
		Name:          "zed",
		DisplayName:   "Zed",
		GlobalPath:    func() (string, error) { return getZedConfigPath() },
		LocalPath:     nil,
		SupportsLocal: false,
		SyncFunc:      syncToZed,
	})
}

func getZedConfigPathImpl() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Zed uses ~/.config/zed/settings.json on all platforms
	return filepath.Join(home, ".config", "zed", "settings.json"), nil
}

func syncToZed(servers []config.MCPServer, path string) error {
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

	// Zed uses "context_servers" with a different format
	contextServers := make(map[string]any)
	for _, server := range servers {
		var serverConfig map[string]any
		if server.Type == "http" {
			serverConfig = map[string]any{
				"url":      server.URL,
				"settings": map[string]any{},
			}
			if len(server.Headers) > 0 {
				serverConfig["headers"] = server.Headers
			}
		} else {
			command := map[string]any{
				"path": server.Command,
			}
			if len(server.Args) > 0 {
				command["args"] = server.Args
			}
			if len(server.Env) > 0 {
				command["env"] = server.Env
			}
			serverConfig = map[string]any{
				"command":  command,
				"settings": map[string]any{},
			}
		}
		contextServers[server.Name] = serverConfig
	}

	settings["context_servers"] = contextServers

	return saveSettingsFile(path, settings)
}
