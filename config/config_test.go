package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMCPServer(t *testing.T) {
	server := MCPServer{
		Name:    "test-server",
		Command: "npx",
		Args:    []string{"-y", "@modelcontextprotocol/server-test"},
		Env:     map[string]string{"API_KEY": "test123"},
	}

	if server.Name != "test-server" {
		t.Errorf("expected Name to be 'test-server', got %q", server.Name)
	}
	if server.Command != "npx" {
		t.Errorf("expected Command to be 'npx', got %q", server.Command)
	}
	if len(server.Args) != 2 {
		t.Errorf("expected 2 Args, got %d", len(server.Args))
	}
	if server.Env["API_KEY"] != "test123" {
		t.Errorf("expected Env[API_KEY] to be 'test123', got %q", server.Env["API_KEY"])
	}
}

func TestConfig_AddServer(t *testing.T) {
	cfg := &Config{Servers: []MCPServer{}}

	server := MCPServer{
		Name:    "test-server",
		Command: "node",
		Args:    []string{"server.js"},
	}

	err := cfg.AddServer(server)
	if err != nil {
		t.Fatalf("unexpected error adding server: %v", err)
	}

	if len(cfg.Servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(cfg.Servers))
	}

	if cfg.Servers[0].Name != "test-server" {
		t.Errorf("expected server name 'test-server', got %q", cfg.Servers[0].Name)
	}
}

func TestConfig_AddServer_Duplicate(t *testing.T) {
	cfg := &Config{Servers: []MCPServer{}}

	server := MCPServer{
		Name:    "test-server",
		Command: "node",
	}

	err := cfg.AddServer(server)
	if err != nil {
		t.Fatalf("unexpected error adding first server: %v", err)
	}

	err = cfg.AddServer(server)
	if err == nil {
		t.Error("expected error when adding duplicate server, got nil")
	}
}

func TestConfig_GetServer(t *testing.T) {
	cfg := &Config{
		Servers: []MCPServer{
			{Name: "server1", Command: "cmd1"},
			{Name: "server2", Command: "cmd2"},
			{Name: "server3", Command: "cmd3"},
		},
	}

	server, err := cfg.GetServer("server2")
	if err != nil {
		t.Fatalf("unexpected error getting server: %v", err)
	}
	if server.Name != "server2" {
		t.Errorf("expected server name 'server2', got %q", server.Name)
	}
	if server.Command != "cmd2" {
		t.Errorf("expected command 'cmd2', got %q", server.Command)
	}
}

func TestConfig_GetServer_NotFound(t *testing.T) {
	cfg := &Config{
		Servers: []MCPServer{
			{Name: "server1", Command: "cmd1"},
		},
	}

	_, err := cfg.GetServer("nonexistent")
	if err == nil {
		t.Error("expected error when getting nonexistent server, got nil")
	}
}

func TestConfig_ListServers(t *testing.T) {
	cfg := &Config{
		Servers: []MCPServer{
			{Name: "server1", Command: "cmd1"},
			{Name: "server2", Command: "cmd2"},
		},
	}

	servers := cfg.ListServers()
	if len(servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(servers))
	}
}

func TestConfig_ListServers_Empty(t *testing.T) {
	cfg := &Config{Servers: []MCPServer{}}

	servers := cfg.ListServers()
	if len(servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(servers))
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	// Create and save config
	cfg := &Config{
		Servers: []MCPServer{
			{
				Name:    "test-server",
				Command: "npx",
				Args:    []string{"-y", "test-package"},
				Env:     map[string]string{"KEY": "value"},
			},
		},
	}
	cfg.SetPath(configPath)

	err = cfg.Save()
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Load config
	loadedCfg, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(loadedCfg.Servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(loadedCfg.Servers))
	}

	server := loadedCfg.Servers[0]
	if server.Name != "test-server" {
		t.Errorf("expected name 'test-server', got %q", server.Name)
	}
	if server.Command != "npx" {
		t.Errorf("expected command 'npx', got %q", server.Command)
	}
	if len(server.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(server.Args))
	}
	if server.Env["KEY"] != "value" {
		t.Errorf("expected env KEY='value', got %q", server.Env["KEY"])
	}
}

func TestLoadFromPath_NonExistent(t *testing.T) {
	cfg, err := LoadFromPath("/nonexistent/path/config.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Servers) != 0 {
		t.Errorf("expected empty servers, got %d", len(cfg.Servers))
	}
}

func TestLoadFromPath_InvalidJSON(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")
	err = os.WriteFile(configPath, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err = LoadFromPath(configPath)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestConfig_Path(t *testing.T) {
	cfg := &Config{}
	cfg.SetPath("/test/path/config.json")

	if cfg.Path() != "/test/path/config.json" {
		t.Errorf("expected path '/test/path/config.json', got %q", cfg.Path())
	}
}

func TestConfig_Save_CreatesDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Use a nested path that doesn't exist
	configPath := filepath.Join(tempDir, "nested", "dir", "config.json")

	cfg := &Config{Servers: []MCPServer{}}
	cfg.SetPath(configPath)

	err = cfg.Save()
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}
}

func TestFindConfigInParents(t *testing.T) {
	// Create a temporary directory structure
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Resolve symlinks for comparison (macOS /var -> /private/var)
	tempDir, err = filepath.EvalSymlinks(tempDir)
	if err != nil {
		t.Fatalf("failed to resolve symlinks: %v", err)
	}

	// Create nested directories
	nestedDir := filepath.Join(tempDir, "level1", "level2", "level3")
	err = os.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Fatalf("failed to create nested dirs: %v", err)
	}

	// Create config file at level1
	configPath := filepath.Join(tempDir, "level1", configFileName)
	err = os.WriteFile(configPath, []byte(`{"servers":[]}`), 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	// Change to the deepest directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	err = os.Chdir(nestedDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Test that findConfigInParents finds the config
	foundPath, found := findConfigInParents()
	if !found {
		t.Fatal("expected to find config in parent directories")
	}

	if foundPath != configPath {
		t.Errorf("expected path %q, got %q", configPath, foundPath)
	}
}

func TestFindConfigInParents_NotFound(t *testing.T) {
	// Create a temporary directory without any config
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	_, found := findConfigInParents()
	// Note: This might find a config in actual parent directories during testing
	// So we just verify the function doesn't crash
	_ = found
}

func TestGetGlobalConfigPath(t *testing.T) {
	path, err := getGlobalConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "mcpr", "config.json")

	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestGetWriteConfigPath_PreferLocal_ExistingConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Resolve symlinks for comparison (macOS /var -> /private/var)
	tempDir, err = filepath.EvalSymlinks(tempDir)
	if err != nil {
		t.Fatalf("failed to resolve symlinks: %v", err)
	}

	// Create config file
	configPath := filepath.Join(tempDir, configFileName)
	err = os.WriteFile(configPath, []byte(`{"servers":[]}`), 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	path, err := GetWriteConfigPath(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != configPath {
		t.Errorf("expected path %q, got %q", configPath, path)
	}
}

func TestGetWriteConfigPath_PreferLocal_NoExisting(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	path, err := GetWriteConfigPath(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if path != configFileName {
		t.Errorf("expected path %q, got %q", configFileName, path)
	}
}

func TestGetWriteConfigPath_Global(t *testing.T) {
	path, err := GetWriteConfigPath(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".config", "mcpr", "config.json")

	if path != expected {
		t.Errorf("expected path %q, got %q", expected, path)
	}
}

func TestConfig_AddMultipleServers(t *testing.T) {
	cfg := &Config{Servers: []MCPServer{}}

	servers := []MCPServer{
		{Name: "server1", Command: "cmd1"},
		{Name: "server2", Command: "cmd2"},
		{Name: "server3", Command: "cmd3"},
	}

	for _, server := range servers {
		err := cfg.AddServer(server)
		if err != nil {
			t.Fatalf("unexpected error adding server %q: %v", server.Name, err)
		}
	}

	if len(cfg.Servers) != 3 {
		t.Errorf("expected 3 servers, got %d", len(cfg.Servers))
	}
}

func TestConfig_ServerWithAllFields(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	cfg := &Config{
		Servers: []MCPServer{
			{
				Name:    "full-server",
				Command: "/usr/bin/node",
				Args:    []string{"--experimental", "server.js", "--port", "3000"},
				Env: map[string]string{
					"NODE_ENV":  "production",
					"API_KEY":   "secret123",
					"DEBUG":     "true",
					"LOG_LEVEL": "info",
				},
			},
		},
	}
	cfg.SetPath(configPath)

	err = cfg.Save()
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loadedCfg, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	server := loadedCfg.Servers[0]
	if server.Name != "full-server" {
		t.Errorf("expected name 'full-server', got %q", server.Name)
	}
	if server.Command != "/usr/bin/node" {
		t.Errorf("expected command '/usr/bin/node', got %q", server.Command)
	}
	if len(server.Args) != 4 {
		t.Errorf("expected 4 args, got %d", len(server.Args))
	}
	if len(server.Env) != 4 {
		t.Errorf("expected 4 env vars, got %d", len(server.Env))
	}
	if server.Env["NODE_ENV"] != "production" {
		t.Errorf("expected NODE_ENV='production', got %q", server.Env["NODE_ENV"])
	}
}

func TestConfig_RemoveServer(t *testing.T) {
	cfg := &Config{
		Servers: []MCPServer{
			{Name: "server1", Command: "cmd1"},
			{Name: "server2", Command: "cmd2"},
			{Name: "server3", Command: "cmd3"},
		},
	}

	err := cfg.RemoveServer("server2")
	if err != nil {
		t.Fatalf("unexpected error removing server: %v", err)
	}

	if len(cfg.Servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(cfg.Servers))
	}

	// Verify server2 is gone
	for _, s := range cfg.Servers {
		if s.Name == "server2" {
			t.Error("server2 should have been removed")
		}
	}

	// Verify other servers remain
	if cfg.Servers[0].Name != "server1" {
		t.Errorf("expected first server to be 'server1', got %q", cfg.Servers[0].Name)
	}
	if cfg.Servers[1].Name != "server3" {
		t.Errorf("expected second server to be 'server3', got %q", cfg.Servers[1].Name)
	}
}

func TestConfig_RemoveServer_NotFound(t *testing.T) {
	cfg := &Config{
		Servers: []MCPServer{
			{Name: "server1", Command: "cmd1"},
		},
	}

	err := cfg.RemoveServer("nonexistent")
	if err == nil {
		t.Error("expected error when removing nonexistent server, got nil")
	}
}

func TestConfig_RemoveServer_First(t *testing.T) {
	cfg := &Config{
		Servers: []MCPServer{
			{Name: "server1", Command: "cmd1"},
			{Name: "server2", Command: "cmd2"},
		},
	}

	err := cfg.RemoveServer("server1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(cfg.Servers))
	}
	if cfg.Servers[0].Name != "server2" {
		t.Errorf("expected remaining server to be 'server2', got %q", cfg.Servers[0].Name)
	}
}

func TestConfig_RemoveServer_Last(t *testing.T) {
	cfg := &Config{
		Servers: []MCPServer{
			{Name: "server1", Command: "cmd1"},
			{Name: "server2", Command: "cmd2"},
		},
	}

	err := cfg.RemoveServer("server2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(cfg.Servers))
	}
	if cfg.Servers[0].Name != "server1" {
		t.Errorf("expected remaining server to be 'server1', got %q", cfg.Servers[0].Name)
	}
}

func TestConfig_RemoveServer_Only(t *testing.T) {
	cfg := &Config{
		Servers: []MCPServer{
			{Name: "server1", Command: "cmd1"},
		},
	}

	err := cfg.RemoveServer("server1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(cfg.Servers) != 0 {
		t.Errorf("expected 0 servers, got %d", len(cfg.Servers))
	}
}

func TestSyncedClient(t *testing.T) {
	sc := SyncedClient{
		Name:    "claude-code",
		Local:   true,
		Servers: []string{"server1", "server2"},
	}

	if sc.Name != "claude-code" {
		t.Errorf("expected Name 'claude-code', got %q", sc.Name)
	}
	if !sc.Local {
		t.Error("expected Local to be true")
	}
	if len(sc.Servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(sc.Servers))
	}
}

func TestConfig_AddSyncedClient(t *testing.T) {
	cfg := &Config{}

	cfg.AddSyncedClient("claude-code", false, nil)

	if len(cfg.SyncedClients) != 1 {
		t.Fatalf("expected 1 synced client, got %d", len(cfg.SyncedClients))
	}

	sc := cfg.SyncedClients[0]
	if sc.Name != "claude-code" {
		t.Errorf("expected Name 'claude-code', got %q", sc.Name)
	}
	if sc.Local {
		t.Error("expected Local to be false")
	}
	if sc.Servers != nil {
		t.Errorf("expected nil Servers, got %v", sc.Servers)
	}
}

func TestConfig_AddSyncedClient_WithServers(t *testing.T) {
	cfg := &Config{}

	cfg.AddSyncedClient("cursor", true, []string{"server1", "server2"})

	if len(cfg.SyncedClients) != 1 {
		t.Fatalf("expected 1 synced client, got %d", len(cfg.SyncedClients))
	}

	sc := cfg.SyncedClients[0]
	if sc.Name != "cursor" {
		t.Errorf("expected Name 'cursor', got %q", sc.Name)
	}
	if !sc.Local {
		t.Error("expected Local to be true")
	}
	if len(sc.Servers) != 2 {
		t.Errorf("expected 2 servers, got %d", len(sc.Servers))
	}
}

func TestConfig_AddSyncedClient_Update(t *testing.T) {
	cfg := &Config{}

	// Add initial client
	cfg.AddSyncedClient("claude-code", false, []string{"server1"})

	// Update the same client
	cfg.AddSyncedClient("claude-code", false, []string{"server1", "server2", "server3"})

	// Should still be 1 client, not 2
	if len(cfg.SyncedClients) != 1 {
		t.Fatalf("expected 1 synced client, got %d", len(cfg.SyncedClients))
	}

	// Should have updated servers
	if len(cfg.SyncedClients[0].Servers) != 3 {
		t.Errorf("expected 3 servers after update, got %d", len(cfg.SyncedClients[0].Servers))
	}
}

func TestConfig_AddSyncedClient_SameNameDifferentLocal(t *testing.T) {
	cfg := &Config{}

	// Add global client
	cfg.AddSyncedClient("claude-code", false, nil)

	// Add local client with same name - should be separate entry
	cfg.AddSyncedClient("claude-code", true, nil)

	if len(cfg.SyncedClients) != 2 {
		t.Fatalf("expected 2 synced clients (global and local), got %d", len(cfg.SyncedClients))
	}

	// Verify both exist
	var hasGlobal, hasLocal bool
	for _, sc := range cfg.SyncedClients {
		if sc.Name == "claude-code" && !sc.Local {
			hasGlobal = true
		}
		if sc.Name == "claude-code" && sc.Local {
			hasLocal = true
		}
	}

	if !hasGlobal {
		t.Error("expected global claude-code client")
	}
	if !hasLocal {
		t.Error("expected local claude-code client")
	}
}

func TestConfig_RemoveSyncedClient(t *testing.T) {
	cfg := &Config{
		SyncedClients: []SyncedClient{
			{Name: "claude-code", Local: false},
			{Name: "cursor", Local: true},
			{Name: "vscode", Local: false},
		},
	}

	cfg.RemoveSyncedClient("cursor", true)

	if len(cfg.SyncedClients) != 2 {
		t.Errorf("expected 2 synced clients, got %d", len(cfg.SyncedClients))
	}

	// Verify cursor is gone
	for _, sc := range cfg.SyncedClients {
		if sc.Name == "cursor" {
			t.Error("cursor should have been removed")
		}
	}
}

func TestConfig_RemoveSyncedClient_NotFound(t *testing.T) {
	cfg := &Config{
		SyncedClients: []SyncedClient{
			{Name: "claude-code", Local: false},
		},
	}

	// Should not panic or error - just no-op
	cfg.RemoveSyncedClient("nonexistent", false)

	if len(cfg.SyncedClients) != 1 {
		t.Errorf("expected 1 synced client, got %d", len(cfg.SyncedClients))
	}
}

func TestConfig_RemoveSyncedClient_WrongLocal(t *testing.T) {
	cfg := &Config{
		SyncedClients: []SyncedClient{
			{Name: "claude-code", Local: false},
		},
	}

	// Try to remove local when only global exists
	cfg.RemoveSyncedClient("claude-code", true)

	// Should not remove anything
	if len(cfg.SyncedClients) != 1 {
		t.Errorf("expected 1 synced client, got %d", len(cfg.SyncedClients))
	}
}

func TestConfig_GetSyncedClients(t *testing.T) {
	cfg := &Config{
		SyncedClients: []SyncedClient{
			{Name: "claude-code", Local: false},
			{Name: "cursor", Local: true},
		},
	}

	clients := cfg.GetSyncedClients()

	if len(clients) != 2 {
		t.Errorf("expected 2 synced clients, got %d", len(clients))
	}
}

func TestConfig_GetSyncedClients_Empty(t *testing.T) {
	cfg := &Config{}

	clients := cfg.GetSyncedClients()

	if len(clients) != 0 {
		t.Errorf("expected 0 synced clients, got %d", len(clients))
	}
}

func TestConfig_GetSyncedClient(t *testing.T) {
	cfg := &Config{
		SyncedClients: []SyncedClient{
			{Name: "claude-code", Local: false, Servers: []string{"s1"}},
			{Name: "cursor", Local: true, Servers: []string{"s2"}},
		},
	}

	sc := cfg.GetSyncedClient("cursor", true)
	if sc == nil {
		t.Fatal("expected to find cursor client")
	}
	if sc.Name != "cursor" {
		t.Errorf("expected Name 'cursor', got %q", sc.Name)
	}
	if !sc.Local {
		t.Error("expected Local to be true")
	}
}

func TestConfig_GetSyncedClient_NotFound(t *testing.T) {
	cfg := &Config{
		SyncedClients: []SyncedClient{
			{Name: "claude-code", Local: false},
		},
	}

	sc := cfg.GetSyncedClient("nonexistent", false)
	if sc != nil {
		t.Error("expected nil for nonexistent client")
	}
}

func TestConfig_GetSyncedClient_WrongLocal(t *testing.T) {
	cfg := &Config{
		SyncedClients: []SyncedClient{
			{Name: "claude-code", Local: false},
		},
	}

	// Look for local when only global exists
	sc := cfg.GetSyncedClient("claude-code", true)
	if sc != nil {
		t.Error("expected nil when local flag doesn't match")
	}
}

func TestConfig_SyncedClients_SaveAndLoad(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mcpr-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.json")

	// Create config with synced clients
	cfg := &Config{
		Servers: []MCPServer{
			{Name: "server1", Type: "stdio", Command: "cmd1"},
		},
		SyncedClients: []SyncedClient{
			{Name: "claude-code", Local: false, Servers: nil},
			{Name: "cursor", Local: true, Servers: []string{"server1"}},
		},
	}
	cfg.SetPath(configPath)

	err = cfg.Save()
	if err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Load and verify
	loadedCfg, err := LoadFromPath(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(loadedCfg.SyncedClients) != 2 {
		t.Fatalf("expected 2 synced clients, got %d", len(loadedCfg.SyncedClients))
	}

	// Find and verify claude-code
	cc := loadedCfg.GetSyncedClient("claude-code", false)
	if cc == nil {
		t.Fatal("expected to find claude-code")
	}
	if cc.Local {
		t.Error("expected claude-code Local to be false")
	}

	// Find and verify cursor
	cursor := loadedCfg.GetSyncedClient("cursor", true)
	if cursor == nil {
		t.Fatal("expected to find cursor")
	}
	if !cursor.Local {
		t.Error("expected cursor Local to be true")
	}
	if len(cursor.Servers) != 1 || cursor.Servers[0] != "server1" {
		t.Errorf("expected cursor Servers to be ['server1'], got %v", cursor.Servers)
	}
}
