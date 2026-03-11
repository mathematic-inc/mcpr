package clients

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mathematic-inc/mcpr/config"
)

// Client represents an MCP client that can have servers installed
type Client struct {
	Name          string
	DisplayName   string
	GlobalPath    func() (string, error)
	LocalPath     func() (string, error) // nil if no local config supported
	SupportsLocal bool
	SyncFunc      func(servers []config.MCPServer, path string) error
}

// MCPClientConfig represents the MCP configuration format used by clients
type MCPClientConfig struct {
	MCPServers map[string]MCPServerEntry `json:"mcpServers"`
}

// MCPServerEntry represents a single MCP server entry in client config
type MCPServerEntry struct {
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// clientRegistry holds all registered clients
var clientRegistry = make(map[string]*Client)

// RegisterClient adds a client to the registry
func RegisterClient(client *Client) {
	clientRegistry[client.Name] = client
}

// GetClients returns all supported MCP clients
func GetClients() map[string]*Client {
	return clientRegistry
}

// GetClient returns a specific client by name
func GetClient(name string) (*Client, error) {
	client, ok := clientRegistry[name]
	if !ok {
		return nil, fmt.Errorf("unknown client: %s", name)
	}
	return client, nil
}

// ListClientNames returns all supported client names
func ListClientNames() []string {
	names := make([]string, 0, len(clientRegistry))
	for name := range clientRegistry {
		names = append(names, name)
	}
	return names
}

// Sync synchronizes MCP servers to the client, replacing the existing config
func (c *Client) Sync(servers []config.MCPServer, local bool) (string, error) {
	var path string
	var err error

	if local {
		if !c.SupportsLocal {
			return "", fmt.Errorf("%s does not support local config", c.DisplayName)
		}
		path, err = c.LocalPath()
	} else {
		path, err = c.GlobalPath()
	}

	if err != nil {
		return "", err
	}

	if err := c.SyncFunc(servers, path); err != nil {
		return "", err
	}

	return path, nil
}

// ConfigPath returns the global config path for display
func (c *Client) ConfigPath() (string, error) {
	return c.GlobalPath()
}

// syncToMCPConfig syncs servers to a standard MCP config file (replaces entirely)
func syncToMCPConfig(servers []config.MCPServer, path string) error {
	cfg := &MCPClientConfig{
		MCPServers: make(map[string]MCPServerEntry),
	}

	for _, server := range servers {
		entry := MCPServerEntry{}
		if server.Type == "http" {
			entry.URL = server.URL
			entry.Headers = server.Headers
		} else {
			entry.Command = server.Command
			entry.Args = server.Args
			entry.Env = server.Env
		}
		cfg.MCPServers[server.Name] = entry
	}

	return saveMCPConfig(path, cfg)
}

// syncToSettingsWithKey syncs servers to a settings file with a specific key (preserves other settings)
func syncToSettingsWithKey(servers []config.MCPServer, path string, key string) error {
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
		mcpServers[server.Name] = entry
	}

	settings[key] = mcpServers

	return saveSettingsFile(path, settings)
}

// syncToSettingsWithMcpServers syncs servers to a settings file with mcpServers key
func syncToSettingsWithMcpServers(servers []config.MCPServer, path string) error {
	return syncToSettingsWithKey(servers, path, "mcpServers")
}

// saveSettingsFile saves a settings map to disk
func saveSettingsFile(path string, settings map[string]any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	outData, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(path, outData, 0o644)
}

// saveMCPConfig saves the MCP config to disk
func saveMCPConfig(path string, cfg *MCPClientConfig) error {
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
