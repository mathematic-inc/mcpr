package clients

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jrandolf/mcpr/config"
)

func TestSyncToAntigravityConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "mcp_config.json")

	servers := []config.MCPServer{
		{
			Name:    "test-server",
			Command: "npx",
			Args:    []string{"-y", "test"},
			Env:     map[string]string{"KEY": "value"},
		},
		{
			Name:    "http-server",
			Type:    "http",
			URL:     "https://example.com/mcp",
			Headers: map[string]string{"API_KEY": "secret"},
		},
	}

	err = syncToAntigravityConfig(servers, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var cfg AntigravityConfig
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	if len(cfg.MCPServers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(cfg.MCPServers))
	}

	// Check command server
	cmdServer := cfg.MCPServers["test-server"]
	if cmdServer.Command != "npx" {
		t.Errorf("expected command 'npx', got %q", cmdServer.Command)
	}
	if len(cmdServer.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(cmdServer.Args))
	}

	// Check HTTP server - CRITICAL: Must use ServerURL
	httpServer := cfg.MCPServers["http-server"]
	if httpServer.ServerURL != "https://example.com/mcp" {
		t.Errorf("expected ServerURL 'https://example.com/mcp', got %q", httpServer.ServerURL)
	}
	if httpServer.Headers["API_KEY"] != "secret" {
		t.Errorf("expected header API_KEY=secret, got %q", httpServer.Headers["API_KEY"])
	}
}
