package clients

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mathematic-inc/mcpr/config"
)

// Path functions as variables for testing
var (
	getVSCodeConfigPath = getVSCodeConfigPathImpl
	getVSCodeLocalPath  = getVSCodeLocalPathImpl
)

func init() {
	RegisterClient(&Client{
		Name:          "vscode",
		DisplayName:   "VS Code (Copilot)",
		GlobalPath:    func() (string, error) { return getVSCodeConfigPath() },
		LocalPath:     func() (string, error) { return getVSCodeLocalPath() },
		SupportsLocal: true,
		SyncFunc:      syncToVSCodeMCP,
	})
}

func getVSCodeConfigPathImpl() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Code", "User", "settings.json"), nil
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Code", "User", "settings.json"), nil
	case "linux":
		configDir := os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(home, ".config")
		}
		return filepath.Join(configDir, "Code", "User", "settings.json"), nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func getVSCodeLocalPathImpl() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(cwd, ".vscode", "mcp.json"), nil
}

func syncToVSCodeMCP(servers []config.MCPServer, path string) error {
	// VS Code uses "servers" key in mcp.json
	serversMap := make(map[string]any)
	for _, server := range servers {
		var entry map[string]any
		if server.Type == "http" {
			entry = map[string]any{
				"url": server.URL,
			}
			if len(server.Headers) > 0 {
				entry["headers"] = server.Headers
			}
		} else {
			entry = map[string]any{
				"command": server.Command,
			}
			if len(server.Args) > 0 {
				entry["args"] = server.Args
			}
			if len(server.Env) > 0 {
				entry["env"] = server.Env
			}
		}
		serversMap[server.Name] = entry
	}

	config := map[string]any{
		"servers": serversMap,
	}

	return saveSettingsFile(path, config)
}
