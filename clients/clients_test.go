package clients

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jrandolf/mcpr/config"
)

func TestGetClients(t *testing.T) {
	clients := GetClients()

	expectedClients := []string{"claude-desktop", "claude-code", "cursor", "windsurf", "zed", "opencode", "cline", "vscode", "continue", "codex", "gemini", "kilo-code", "zencoder", "antigravity"}

	for _, name := range expectedClients {
		if _, ok := clients[name]; !ok {
			t.Errorf("expected client %q to be present", name)
		}
	}

	if len(clients) != len(expectedClients) {
		t.Errorf("expected %d clients, got %d", len(expectedClients), len(clients))
	}
}

func TestGetClient(t *testing.T) {
	client, err := GetClient("claude-desktop")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.Name != "claude-desktop" {
		t.Errorf("expected name 'claude-desktop', got %q", client.Name)
	}

	if client.DisplayName != "Claude Desktop" {
		t.Errorf("expected display name 'Claude Desktop', got %q", client.DisplayName)
	}
}

func TestGetClient_NotFound(t *testing.T) {
	_, err := GetClient("nonexistent-client")
	if err == nil {
		t.Error("expected error for nonexistent client, got nil")
	}
}

func TestListClientNames(t *testing.T) {
	names := ListClientNames()

	if len(names) != 14 {
		t.Errorf("expected 14 client names, got %d", len(names))
	}

	// Check that all expected names are present
	expectedNames := map[string]bool{
		"claude-desktop": false,
		"claude-code":    false,
		"cursor":         false,
		"windsurf":       false,
		"zed":            false,
		"opencode":       false,
		"cline":          false,
		"vscode":         false,
		"continue":       false,
		"codex":          false,
		"gemini":         false,
		"kilo-code":      false,
		"zencoder":       false,
		"antigravity":    false,
	}

	for _, name := range names {
		if _, ok := expectedNames[name]; ok {
			expectedNames[name] = true
		} else {
			t.Errorf("unexpected client name: %q", name)
		}
	}

	for name, found := range expectedNames {
		if !found {
			t.Errorf("expected client name %q not found", name)
		}
	}
}

func TestClientConfigPath_ClaudeDesktop(t *testing.T) {
	client, _ := GetClient("claude-desktop")
	path, err := client.ConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	home, _ := os.UserHomeDir()

	var expected string
	switch runtime.GOOS {
	case "darwin":
		expected = filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		expected = filepath.Join(appData, "Claude", "claude_desktop_config.json")
	case "linux":
		expected = filepath.Join(home, ".config", "Claude", "claude_desktop_config.json")
	}

	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestClientConfigPath_ClaudeCode(t *testing.T) {
	client, _ := GetClient("claude-code")
	path, err := client.ConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".claude.json")

	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestClientConfigPath_Cursor(t *testing.T) {
	client, _ := GetClient("cursor")
	path, err := client.ConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".cursor", "mcp.json")

	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestClientConfigPath_Windsurf(t *testing.T) {
	client, _ := GetClient("windsurf")
	path, err := client.ConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	home, _ := os.UserHomeDir()

	var expected string
	switch runtime.GOOS {
	case "darwin":
		expected = filepath.Join(home, "Library", "Application Support", "Windsurf", "User", "globalStorage", "windsurf.mcp", "mcp.json")
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		expected = filepath.Join(appData, "Windsurf", "User", "globalStorage", "windsurf.mcp", "mcp.json")
	case "linux":
		expected = filepath.Join(home, ".config", "Windsurf", "User", "globalStorage", "windsurf.mcp", "mcp.json")
	}

	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestClientConfigPath_Antigravity(t *testing.T) {
	client, _ := GetClient("antigravity")
	path, err := client.ConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".gemini", "antigravity", "mcp_config.json")

	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestMCPClientConfig(t *testing.T) {
	cfg := MCPClientConfig{
		MCPServers: map[string]MCPServerEntry{
			"test-server": {
				Command: "npx",
				Args:    []string{"-y", "test-package"},
				Env:     map[string]string{"KEY": "value"},
			},
		},
	}

	if len(cfg.MCPServers) != 1 {
		t.Errorf("expected 1 server, got %d", len(cfg.MCPServers))
	}

	server := cfg.MCPServers["test-server"]
	if server.Command != "npx" {
		t.Errorf("expected command 'npx', got %q", server.Command)
	}
}

func TestSaveMCPConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	cfg := &MCPClientConfig{
		MCPServers: map[string]MCPServerEntry{
			"test-server": {
				Command: "npx",
				Args:    []string{"-y", "test"},
				Env:     map[string]string{"KEY": "value"},
			},
		},
	}

	err = saveMCPConfig(configPath, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Verify content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var loadedCfg MCPClientConfig
	err = json.Unmarshal(data, &loadedCfg)
	if err != nil {
		t.Fatalf("failed to parse config file: %v", err)
	}

	if len(loadedCfg.MCPServers) != 1 {
		t.Errorf("expected 1 server, got %d", len(loadedCfg.MCPServers))
	}
}

func TestSaveMCPConfig_CreatesDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "nested", "dir", "config.json")

	cfg := &MCPClientConfig{
		MCPServers: map[string]MCPServerEntry{},
	}

	err = saveMCPConfig(configPath, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}
}

func TestSyncToMCPConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	servers := []config.MCPServer{
		{
			Name:    "test-server",
			Command: "npx",
			Args:    []string{"-y", "test-package"},
			Env:     map[string]string{"KEY": "value"},
		},
	}

	err = syncToMCPConfig(servers, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var cfg MCPClientConfig
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	if len(cfg.MCPServers) != 1 {
		t.Errorf("expected 1 server, got %d", len(cfg.MCPServers))
	}

	server, ok := cfg.MCPServers["test-server"]
	if !ok {
		t.Fatal("expected 'test-server' to be present")
	}

	if server.Command != "npx" {
		t.Errorf("expected command 'npx', got %q", server.Command)
	}
}

func TestSyncToMCPConfig_ReplacesExisting(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	// Create existing config
	existingConfig := MCPClientConfig{
		MCPServers: map[string]MCPServerEntry{
			"existing-server": {
				Command: "node",
				Args:    []string{"existing.js"},
			},
			"another-server": {
				Command: "python",
			},
		},
	}
	data, _ := json.Marshal(existingConfig)
	err = os.WriteFile(configPath, data, 0o644)
	if err != nil {
		t.Fatalf("failed to write existing config: %v", err)
	}

	// Sync new servers (should replace entirely)
	servers := []config.MCPServer{
		{Name: "new-server", Command: "npx", Args: []string{"new"}},
	}

	err = syncToMCPConfig(servers, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify only new server is present (existing ones replaced)
	data, _ = os.ReadFile(configPath)
	var cfg MCPClientConfig
	json.Unmarshal(data, &cfg)

	if len(cfg.MCPServers) != 1 {
		t.Errorf("expected 1 server (replaced), got %d", len(cfg.MCPServers))
	}

	if _, ok := cfg.MCPServers["existing-server"]; ok {
		t.Error("expected 'existing-server' to be replaced")
	}

	if _, ok := cfg.MCPServers["another-server"]; ok {
		t.Error("expected 'another-server' to be replaced")
	}

	if _, ok := cfg.MCPServers["new-server"]; !ok {
		t.Error("expected 'new-server' to be present")
	}
}

func TestSyncToClaudeCode(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "settings.json")

	servers := []config.MCPServer{
		{
			Name:    "test-server",
			Type:    "stdio",
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

	err = syncToClaudeCode(servers, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify
	data, _ := os.ReadFile(configPath)

	var settings map[string]any
	json.Unmarshal(data, &settings)

	mcpServers, ok := settings["mcpServers"].(map[string]any)
	if !ok {
		t.Fatal("expected mcpServers to be present")
	}

	if len(mcpServers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(mcpServers))
	}

	// Check stdio server has type field
	stdioServer, ok := mcpServers["test-server"].(map[string]any)
	if !ok {
		t.Fatal("expected test-server to be present")
	}
	if stdioServer["type"] != "stdio" {
		t.Errorf("expected type 'stdio', got %v", stdioServer["type"])
	}
	if stdioServer["command"] != "npx" {
		t.Errorf("expected command 'npx', got %v", stdioServer["command"])
	}

	// Check http server has type field
	httpServer, ok := mcpServers["http-server"].(map[string]any)
	if !ok {
		t.Fatal("expected http-server to be present")
	}
	if httpServer["type"] != "http" {
		t.Errorf("expected type 'http', got %v", httpServer["type"])
	}
	if httpServer["url"] != "https://example.com/mcp" {
		t.Errorf("expected url 'https://example.com/mcp', got %v", httpServer["url"])
	}
}

func TestSyncToClaudeCode_PreservesOtherSettings(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "settings.json")

	// Create existing settings
	existingSettings := map[string]any{
		"otherSetting": "value",
		"anotherKey":   123,
		"mcpServers": map[string]any{
			"existing-server": map[string]any{
				"command": "node",
			},
		},
	}
	data, _ := json.Marshal(existingSettings)
	os.WriteFile(configPath, data, 0o644)

	servers := []config.MCPServer{
		{Name: "new-server", Command: "npx"},
	}

	err = syncToClaudeCode(servers, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify
	data, _ = os.ReadFile(configPath)
	var settings map[string]any
	json.Unmarshal(data, &settings)

	// Check other settings preserved
	if settings["otherSetting"] != "value" {
		t.Error("expected 'otherSetting' to be preserved")
	}

	// Check mcpServers replaced (only new server)
	mcpServers := settings["mcpServers"].(map[string]any)
	if len(mcpServers) != 1 {
		t.Errorf("expected 1 server, got %d", len(mcpServers))
	}

	if _, ok := mcpServers["existing-server"]; ok {
		t.Error("expected 'existing-server' to be replaced")
	}

	if _, ok := mcpServers["new-server"]; !ok {
		t.Error("expected 'new-server' to be present")
	}
}

func TestClientDisplayNames(t *testing.T) {
	testCases := []struct {
		name        string
		displayName string
	}{
		{"claude-desktop", "Claude Desktop"},
		{"claude-code", "Claude Code"},
		{"cursor", "Cursor"},
		{"windsurf", "Windsurf"},
		{"zed", "Zed"},
		{"opencode", "OpenCode"},
		{"cline", "Cline"},
		{"vscode", "VS Code (Copilot)"},
		{"continue", "Continue"},
		{"codex", "Codex (OpenAI)"},
		{"gemini", "Gemini CLI"},
		{"kilo-code", "Kilo Code"},
		{"zencoder", "ZenCoder"},
		{"antigravity", "Antigravity (Google)"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := GetClient(tc.name)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if client.DisplayName != tc.displayName {
				t.Errorf("expected display name %q, got %q", tc.displayName, client.DisplayName)
			}
		})
	}
}

func TestSyncMultipleServers(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	servers := []config.MCPServer{
		{Name: "server1", Command: "cmd1", Args: []string{"arg1"}},
		{Name: "server2", Command: "cmd2", Args: []string{"arg2"}},
		{Name: "server3", Command: "cmd3", Env: map[string]string{"KEY": "val"}},
	}

	err = syncToMCPConfig(servers, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify
	data, _ := os.ReadFile(configPath)

	var cfg MCPClientConfig
	json.Unmarshal(data, &cfg)

	if len(cfg.MCPServers) != 3 {
		t.Errorf("expected 3 servers, got %d", len(cfg.MCPServers))
	}

	for _, server := range servers {
		if _, ok := cfg.MCPServers[server.Name]; !ok {
			t.Errorf("expected server %q to be present", server.Name)
		}
	}
}

func TestSyncServerWithNoArgs(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	servers := []config.MCPServer{
		{Name: "simple-server", Command: "my-server"},
	}

	err = syncToMCPConfig(servers, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify
	data, _ := os.ReadFile(configPath)

	var cfg MCPClientConfig
	json.Unmarshal(data, &cfg)

	server := cfg.MCPServers["simple-server"]
	if server.Command != "my-server" {
		t.Errorf("expected command 'my-server', got %q", server.Command)
	}
}

func TestClientSupportsLocal(t *testing.T) {
	testCases := []struct {
		name          string
		supportsLocal bool
	}{
		{"claude-desktop", false},
		{"claude-code", true},
		{"cursor", true},
		{"windsurf", true},
		{"zed", false},
		{"opencode", true},
		{"cline", false},
		{"vscode", true},
		{"continue", false},
		{"codex", false},
		{"gemini", true},
		{"kilo-code", true},
		{"antigravity", true},
		{"zencoder", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client, err := GetClient(tc.name)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if client.SupportsLocal != tc.supportsLocal {
				t.Errorf("expected SupportsLocal=%v, got %v", tc.supportsLocal, client.SupportsLocal)
			}
		})
	}
}

func TestClientSync_Global(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	// Override the global path function
	originalFunc := getClaudeDesktopConfigPath
	getClaudeDesktopConfigPath = func() (string, error) {
		return configPath, nil
	}
	defer func() { getClaudeDesktopConfigPath = originalFunc }()

	client, _ := GetClient("claude-desktop")
	servers := []config.MCPServer{
		{Name: "test-server", Command: "test"},
	}

	path, err := client.Sync(servers, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != configPath {
		t.Errorf("expected path %q, got %q", configPath, path)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}
}

func TestClientSync_Local(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	localPath := filepath.Join(tempDir, ".cursor", "mcp.json")

	// Override the local path function
	originalFunc := getCursorLocalPath
	getCursorLocalPath = func() (string, error) {
		return localPath, nil
	}
	defer func() { getCursorLocalPath = originalFunc }()

	client, _ := GetClient("cursor")
	servers := []config.MCPServer{
		{Name: "test-server", Command: "test"},
	}

	path, err := client.Sync(servers, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != localPath {
		t.Errorf("expected path %q, got %q", localPath, path)
	}

	// Verify file was created
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}
}

func TestClientSync_LocalNotSupported(t *testing.T) {
	client, _ := GetClient("claude-desktop")
	servers := []config.MCPServer{
		{Name: "test-server", Command: "test"},
	}

	_, err := client.Sync(servers, true)
	if err == nil {
		t.Error("expected error for local sync on unsupported client")
	}
}

func TestClaudeCodeLocalPath(t *testing.T) {
	cwd, _ := os.Getwd()
	expected := filepath.Join(cwd, ".mcp.json")

	path, err := getClaudeCodeLocalPathImpl()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestCursorLocalPath(t *testing.T) {
	cwd, _ := os.Getwd()
	expected := filepath.Join(cwd, ".cursor", "mcp.json")

	path, err := getCursorLocalPathImpl()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestWindsurfLocalPath(t *testing.T) {
	cwd, _ := os.Getwd()
	expected := filepath.Join(cwd, ".windsurf", "mcp.json")

	path, err := getWindsurfLocalPathImpl()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestVSCodeLocalPath(t *testing.T) {
	cwd, _ := os.Getwd()
	expected := filepath.Join(cwd, ".vscode", "mcp.json")

	path, err := getVSCodeLocalPathImpl()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestSyncToZed(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "settings.json")

	servers := []config.MCPServer{
		{
			Name:    "test-server",
			Command: "npx",
			Args:    []string{"-y", "test-package"},
			Env:     map[string]string{"KEY": "value"},
		},
	}

	err = syncToZed(servers, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var settings map[string]any
	err = json.Unmarshal(data, &settings)
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	// Check context_servers key exists
	contextServers, ok := settings["context_servers"].(map[string]any)
	if !ok {
		t.Fatal("expected 'context_servers' to be present")
	}

	if len(contextServers) != 1 {
		t.Errorf("expected 1 server, got %d", len(contextServers))
	}

	// Check server structure
	serverEntry, ok := contextServers["test-server"].(map[string]any)
	if !ok {
		t.Fatal("expected 'test-server' to be present")
	}

	command, ok := serverEntry["command"].(map[string]any)
	if !ok {
		t.Fatal("expected 'command' to be present")
	}

	if command["path"] != "npx" {
		t.Errorf("expected command path 'npx', got %v", command["path"])
	}
}

func TestSyncToZed_PreservesOtherSettings(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "settings.json")

	// Create existing settings
	existingSettings := map[string]any{
		"theme": "dark",
		"context_servers": map[string]any{
			"existing-server": map[string]any{
				"command": map[string]any{"path": "node"},
			},
		},
	}
	data, _ := json.Marshal(existingSettings)
	os.WriteFile(configPath, data, 0o644)

	servers := []config.MCPServer{
		{Name: "new-server", Command: "npx"},
	}

	err = syncToZed(servers, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify
	data, _ = os.ReadFile(configPath)
	var settings map[string]any
	json.Unmarshal(data, &settings)

	// Check other settings preserved
	if settings["theme"] != "dark" {
		t.Error("expected 'theme' to be preserved")
	}

	// Check context_servers replaced
	contextServers := settings["context_servers"].(map[string]any)
	if len(contextServers) != 1 {
		t.Errorf("expected 1 server, got %d", len(contextServers))
	}

	if _, ok := contextServers["existing-server"]; ok {
		t.Error("expected 'existing-server' to be replaced")
	}

	if _, ok := contextServers["new-server"]; !ok {
		t.Error("expected 'new-server' to be present")
	}
}

func TestSyncToVSCodeMCP(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "mcp.json")

	servers := []config.MCPServer{
		{
			Name:    "test-server",
			Command: "npx",
			Args:    []string{"-y", "test-package"},
			Env:     map[string]string{"KEY": "value"},
		},
	}

	err = syncToVSCodeMCP(servers, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var cfg map[string]any
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	// Check servers key exists
	serversMap, ok := cfg["servers"].(map[string]any)
	if !ok {
		t.Fatal("expected 'servers' to be present")
	}

	if len(serversMap) != 1 {
		t.Errorf("expected 1 server, got %d", len(serversMap))
	}

	// Check server entry
	serverEntry, ok := serversMap["test-server"].(map[string]any)
	if !ok {
		t.Fatal("expected 'test-server' to be present")
	}

	if serverEntry["command"] != "npx" {
		t.Errorf("expected command 'npx', got %v", serverEntry["command"])
	}
}

func TestSyncToContinue(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	servers := []config.MCPServer{
		{
			Name:    "test-server",
			Command: "npx",
			Args:    []string{"-y", "test-package"},
			Env:     map[string]string{"KEY": "value"},
		},
	}

	err = syncToContinue(servers, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	var cfg map[string]any
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	// Check mcpServers is an array
	mcpServers, ok := cfg["mcpServers"].([]any)
	if !ok {
		t.Fatal("expected 'mcpServers' to be an array")
	}

	if len(mcpServers) != 1 {
		t.Errorf("expected 1 server, got %d", len(mcpServers))
	}

	// Check server entry structure
	serverEntry, ok := mcpServers[0].(map[string]any)
	if !ok {
		t.Fatal("expected server entry to be a map")
	}

	if serverEntry["name"] != "test-server" {
		t.Errorf("expected name 'test-server', got %v", serverEntry["name"])
	}

	transport, ok := serverEntry["transport"].(map[string]any)
	if !ok {
		t.Fatal("expected 'transport' to be present")
	}

	if transport["type"] != "stdio" {
		t.Errorf("expected transport type 'stdio', got %v", transport["type"])
	}

	if transport["command"] != "npx" {
		t.Errorf("expected command 'npx', got %v", transport["command"])
	}
}

func TestSyncToContinue_PreservesOtherSettings(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	// Create existing settings
	existingSettings := map[string]any{
		"models": []string{"gpt-4"},
		"mcpServers": []map[string]any{
			{"name": "existing-server"},
		},
	}
	data, _ := json.Marshal(existingSettings)
	os.WriteFile(configPath, data, 0o644)

	servers := []config.MCPServer{
		{Name: "new-server", Command: "npx"},
	}

	err = syncToContinue(servers, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify
	data, _ = os.ReadFile(configPath)
	var settings map[string]any
	json.Unmarshal(data, &settings)

	// Check other settings preserved
	models, ok := settings["models"].([]any)
	if !ok || len(models) != 1 {
		t.Error("expected 'models' to be preserved")
	}

	// Check mcpServers replaced
	mcpServers := settings["mcpServers"].([]any)
	if len(mcpServers) != 1 {
		t.Errorf("expected 1 server, got %d", len(mcpServers))
	}

	serverEntry := mcpServers[0].(map[string]any)
	if serverEntry["name"] != "new-server" {
		t.Error("expected 'new-server' to be present")
	}
}

func TestGeminiLocalPath(t *testing.T) {
	cwd, _ := os.Getwd()
	expected := filepath.Join(cwd, ".gemini", "settings.json")

	path, err := getGeminiLocalPathImpl()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestKiloCodeLocalPath(t *testing.T) {
	cwd, _ := os.Getwd()
	expected := filepath.Join(cwd, ".kilocode", "mcp.json")

	path, err := getKiloCodeLocalPathImpl()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestSyncToCodex(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.toml")

	servers := []config.MCPServer{
		{
			Name:    "test-server",
			Command: "npx",
			Args:    []string{"-y", "test-package"},
			Env:     map[string]string{"KEY": "value"},
		},
	}

	err = syncToCodex(servers, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	content := string(data)

	// Check for TOML format
	if !tomlHasPrefix(content, "[mcp_servers.test-server]") && !contains(content, "[mcp_servers.test-server]") {
		t.Error("expected TOML section [mcp_servers.test-server] to be present")
	}

	if !contains(content, `command = "npx"`) {
		t.Error("expected command to be present")
	}
}

func TestSyncToCodex_PreservesOtherSettings(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.toml")

	// Create existing settings
	existingContent := `model = "gpt-4"
temperature = 0.7

[mcp_servers.existing-server]
command = "node"
args = ["old.js"]
`
	os.WriteFile(configPath, []byte(existingContent), 0o644)

	servers := []config.MCPServer{
		{Name: "new-server", Command: "npx"},
	}

	err = syncToCodex(servers, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify
	data, _ := os.ReadFile(configPath)
	content := string(data)

	// Check other settings preserved
	if !contains(content, `model = "gpt-4"`) {
		t.Error("expected 'model' to be preserved")
	}

	// Check existing server replaced
	if contains(content, "[mcp_servers.existing-server]") {
		t.Error("expected 'existing-server' to be replaced")
	}

	// Check new server present
	if !contains(content, "[mcp_servers.new-server]") {
		t.Error("expected 'new-server' to be present")
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstring(s, substr) >= 0)
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestSyncIdempotency_MCPConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	servers := []config.MCPServer{
		{
			Name:    "server-a",
			Command: "cmd-a",
			Args:    []string{"arg1", "arg2"},
			Env:     map[string]string{"KEY_Z": "val_z", "KEY_A": "val_a", "KEY_M": "val_m"},
		},
		{
			Name:    "server-b",
			Command: "cmd-b",
		},
	}

	// First sync
	err = syncToMCPConfig(servers, configPath)
	if err != nil {
		t.Fatalf("first sync failed: %v", err)
	}

	firstContent, _ := os.ReadFile(configPath)

	// Second sync (should produce identical output)
	err = syncToMCPConfig(servers, configPath)
	if err != nil {
		t.Fatalf("second sync failed: %v", err)
	}

	secondContent, _ := os.ReadFile(configPath)

	if string(firstContent) != string(secondContent) {
		t.Errorf("sync is not idempotent:\nFirst:\n%s\n\nSecond:\n%s", firstContent, secondContent)
	}

	// Third sync to be extra sure
	err = syncToMCPConfig(servers, configPath)
	if err != nil {
		t.Fatalf("third sync failed: %v", err)
	}

	thirdContent, _ := os.ReadFile(configPath)

	if string(firstContent) != string(thirdContent) {
		t.Errorf("sync is not idempotent after third run:\nFirst:\n%s\n\nThird:\n%s", firstContent, thirdContent)
	}
}

func TestSyncIdempotency_Codex(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.toml")

	servers := []config.MCPServer{
		{
			Name:    "server-a",
			Command: "cmd-a",
			Args:    []string{"arg1", "arg2"},
			Env:     map[string]string{"KEY_Z": "val_z", "KEY_A": "val_a", "KEY_M": "val_m"},
		},
		{
			Name:    "server-b",
			Command: "cmd-b",
		},
	}

	// First sync
	err = syncToCodex(servers, configPath)
	if err != nil {
		t.Fatalf("first sync failed: %v", err)
	}

	firstContent, _ := os.ReadFile(configPath)

	// Second sync
	err = syncToCodex(servers, configPath)
	if err != nil {
		t.Fatalf("second sync failed: %v", err)
	}

	secondContent, _ := os.ReadFile(configPath)

	if string(firstContent) != string(secondContent) {
		t.Errorf("Codex sync is not idempotent:\nFirst:\n%s\n\nSecond:\n%s", firstContent, secondContent)
	}
}

func TestSyncIdempotency_Continue(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	servers := []config.MCPServer{
		{
			Name:    "server-a",
			Command: "cmd-a",
			Args:    []string{"arg1", "arg2"},
			Env:     map[string]string{"KEY_Z": "val_z", "KEY_A": "val_a"},
		},
		{
			Name:    "server-b",
			Command: "cmd-b",
		},
	}

	// First sync
	err = syncToContinue(servers, configPath)
	if err != nil {
		t.Fatalf("first sync failed: %v", err)
	}

	firstContent, _ := os.ReadFile(configPath)

	// Second sync
	err = syncToContinue(servers, configPath)
	if err != nil {
		t.Fatalf("second sync failed: %v", err)
	}

	secondContent, _ := os.ReadFile(configPath)

	if string(firstContent) != string(secondContent) {
		t.Errorf("Continue sync is not idempotent:\nFirst:\n%s\n\nSecond:\n%s", firstContent, secondContent)
	}
}

func TestSyncIdempotency_Zed(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "settings.json")

	servers := []config.MCPServer{
		{
			Name:    "server-a",
			Command: "cmd-a",
			Args:    []string{"arg1", "arg2"},
			Env:     map[string]string{"KEY_Z": "val_z", "KEY_A": "val_a"},
		},
		{
			Name:    "server-b",
			Command: "cmd-b",
		},
	}

	// First sync
	err = syncToZed(servers, configPath)
	if err != nil {
		t.Fatalf("first sync failed: %v", err)
	}

	firstContent, _ := os.ReadFile(configPath)

	// Second sync
	err = syncToZed(servers, configPath)
	if err != nil {
		t.Fatalf("second sync failed: %v", err)
	}

	secondContent, _ := os.ReadFile(configPath)

	if string(firstContent) != string(secondContent) {
		t.Errorf("Zed sync is not idempotent:\nFirst:\n%s\n\nSecond:\n%s", firstContent, secondContent)
	}
}

func TestSyncIdempotency_VSCode(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "mcp.json")

	servers := []config.MCPServer{
		{
			Name:    "server-a",
			Command: "cmd-a",
			Args:    []string{"arg1", "arg2"},
			Env:     map[string]string{"KEY_Z": "val_z", "KEY_A": "val_a"},
		},
		{
			Name:    "server-b",
			Command: "cmd-b",
		},
	}

	// First sync
	err = syncToVSCodeMCP(servers, configPath)
	if err != nil {
		t.Fatalf("first sync failed: %v", err)
	}

	firstContent, _ := os.ReadFile(configPath)

	// Second sync
	err = syncToVSCodeMCP(servers, configPath)
	if err != nil {
		t.Fatalf("second sync failed: %v", err)
	}

	secondContent, _ := os.ReadFile(configPath)

	if string(firstContent) != string(secondContent) {
		t.Errorf("VS Code sync is not idempotent:\nFirst:\n%s\n\nSecond:\n%s", firstContent, secondContent)
	}
}

func TestOpenCodeConfigPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "opencode", "opencode.json")

	path, err := getOpenCodeConfigPathImpl()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestOpenCodeLocalPath(t *testing.T) {
	cwd, _ := os.Getwd()
	expected := filepath.Join(cwd, "opencode.json")

	path, err := getOpenCodeLocalPathImpl()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestSyncToOpenCode(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "opencode.json")

	servers := []config.MCPServer{
		{
			Name:    "test-server",
			Type:    "stdio",
			Command: "npx",
			Args:    []string{"-y", "test-package"},
			Env:     map[string]string{"KEY": "value"},
		},
		{
			Name:    "http-server",
			Type:    "http",
			URL:     "https://example.com/mcp",
			Headers: map[string]string{"Authorization": "Bearer token"},
		},
	}

	err = syncToOpenCode(servers, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify
	data, _ := os.ReadFile(configPath)

	var settings map[string]any
	json.Unmarshal(data, &settings)

	mcp, ok := settings["mcp"].(map[string]any)
	if !ok {
		t.Fatal("expected 'mcp' to be present")
	}

	if len(mcp) != 2 {
		t.Errorf("expected 2 servers, got %d", len(mcp))
	}

	// Check local server (stdio)
	localServer, ok := mcp["test-server"].(map[string]any)
	if !ok {
		t.Fatal("expected 'test-server' to be present")
	}
	if localServer["type"] != "local" {
		t.Errorf("expected type 'local', got %v", localServer["type"])
	}
	command, ok := localServer["command"].([]any)
	if !ok {
		t.Fatal("expected 'command' to be an array")
	}
	if len(command) != 3 {
		t.Errorf("expected command array length 3, got %d", len(command))
	}
	if command[0] != "npx" {
		t.Errorf("expected command[0] 'npx', got %v", command[0])
	}
	if command[1] != "-y" {
		t.Errorf("expected command[1] '-y', got %v", command[1])
	}
	if command[2] != "test-package" {
		t.Errorf("expected command[2] 'test-package', got %v", command[2])
	}
	environment, ok := localServer["environment"].(map[string]any)
	if !ok {
		t.Fatal("expected 'environment' to be present")
	}
	if environment["KEY"] != "value" {
		t.Errorf("expected environment KEY 'value', got %v", environment["KEY"])
	}

	// Check remote server (http)
	remoteServer, ok := mcp["http-server"].(map[string]any)
	if !ok {
		t.Fatal("expected 'http-server' to be present")
	}
	if remoteServer["type"] != "remote" {
		t.Errorf("expected type 'remote', got %v", remoteServer["type"])
	}
	if remoteServer["url"] != "https://example.com/mcp" {
		t.Errorf("expected url 'https://example.com/mcp', got %v", remoteServer["url"])
	}
	headers, ok := remoteServer["headers"].(map[string]any)
	if !ok {
		t.Fatal("expected 'headers' to be present")
	}
	if headers["Authorization"] != "Bearer token" {
		t.Errorf("expected Authorization header 'Bearer token', got %v", headers["Authorization"])
	}
}

func TestSyncToOpenCode_PreservesOtherSettings(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "opencode.json")

	// Create existing settings
	existingSettings := map[string]any{
		"$schema": "https://opencode.ai/config.json",
		"theme":   "dark",
		"mcp": map[string]any{
			"existing-server": map[string]any{
				"type":    "local",
				"command": []string{"node", "old.js"},
			},
		},
	}
	data, _ := json.Marshal(existingSettings)
	os.WriteFile(configPath, data, 0o644)

	servers := []config.MCPServer{
		{Name: "new-server", Command: "npx"},
	}

	err = syncToOpenCode(servers, configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify
	data, _ = os.ReadFile(configPath)
	var settings map[string]any
	json.Unmarshal(data, &settings)

	// Check other settings preserved
	if settings["$schema"] != "https://opencode.ai/config.json" {
		t.Error("expected '$schema' to be preserved")
	}
	if settings["theme"] != "dark" {
		t.Error("expected 'theme' to be preserved")
	}

	// Check mcp replaced
	mcp := settings["mcp"].(map[string]any)
	if len(mcp) != 1 {
		t.Errorf("expected 1 server, got %d", len(mcp))
	}

	if _, ok := mcp["existing-server"]; ok {
		t.Error("expected 'existing-server' to be replaced")
	}

	if _, ok := mcp["new-server"]; !ok {
		t.Error("expected 'new-server' to be present")
	}
}

func TestSyncIdempotency_OpenCode(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "opencode.json")

	servers := []config.MCPServer{
		{
			Name:    "server-a",
			Command: "cmd-a",
			Args:    []string{"arg1", "arg2"},
			Env:     map[string]string{"KEY_Z": "val_z", "KEY_A": "val_a"},
		},
		{
			Name:    "server-b",
			Command: "cmd-b",
		},
	}

	// First sync
	err = syncToOpenCode(servers, configPath)
	if err != nil {
		t.Fatalf("first sync failed: %v", err)
	}

	firstContent, _ := os.ReadFile(configPath)

	// Second sync
	err = syncToOpenCode(servers, configPath)
	if err != nil {
		t.Fatalf("second sync failed: %v", err)
	}

	secondContent, _ := os.ReadFile(configPath)

	if string(firstContent) != string(secondContent) {
		t.Errorf("OpenCode sync is not idempotent:\nFirst:\n%s\n\nSecond:\n%s", firstContent, secondContent)
	}
}
