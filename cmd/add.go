package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mathematic-inc/mcpr/config"

	"github.com/spf13/cobra"
)

var addLocal bool

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new MCP server configuration",
	Long: `Add a new MCP server to your configuration.

The server configuration will be stored in:
  - Local mcpr.json (if found in current or parent directories, or with --local flag)
  - ~/.config/mcpr/config.json (global default)

Use one of the subcommands:
  mcpr add stdio  - Add a stdio-based MCP server
  mcpr add http   - Add an HTTP/SSE-based MCP server`,
}

// stdio subcommand
var (
	stdioName string
	stdioEnv  []string
)

var addStdioCmd = &cobra.Command{
	Use:   "stdio [command] [args...]",
	Short: "Add a stdio-based MCP server",
	Long: `Add a stdio-based MCP server that communicates via stdin/stdout.

Examples:
  # Add a server with npx
  mcpr add stdio npx -y @modelcontextprotocol/server-filesystem /path

  # Add with custom name
  mcpr add stdio --name my-server npx -y @modelcontextprotocol/server-filesystem

  # Add with environment variables
  mcpr add stdio --env API_KEY=xxx --env DEBUG=true node server.js

  # Add to local config
  mcpr add stdio --local ./my-server`,
	Args: cobra.MinimumNArgs(1),
	RunE: runAddStdio,
}

// http subcommand
var (
	httpName    string
	httpHeaders []string
)

var addHttpCmd = &cobra.Command{
	Use:   "http [url]",
	Short: "Add an HTTP/SSE-based MCP server",
	Long: `Add an HTTP/SSE-based MCP server that communicates over HTTP.

Examples:
  # Add a remote server
  mcpr add http https://example.com/mcp

  # Add with custom name
  mcpr add http --name my-api https://example.com/mcp

  # Add with headers
  mcpr add http --header Authorization=Bearer\ token https://example.com/mcp

  # Add to local config
  mcpr add http --local https://example.com/mcp`,
	Args: cobra.ExactArgs(1),
	RunE: runAddHttp,
}

func init() {
	// Parent add command
	addCmd.PersistentFlags().BoolVarP(&addLocal, "local", "l", false, "Save to local mcpr.json instead of global config")

	// stdio subcommand flags
	addStdioCmd.Flags().StringVarP(&stdioName, "name", "n", "", "Server name (defaults to command name)")
	addStdioCmd.Flags().StringSliceVarP(&stdioEnv, "env", "e", nil, "Environment variables (KEY=VALUE)")
	// Disable interspersed flags so args like "-y" aren't parsed as flags
	addStdioCmd.Flags().SetInterspersed(false)

	// http subcommand flags
	addHttpCmd.Flags().StringVarP(&httpName, "name", "n", "", "Server name (defaults to URL host)")
	addHttpCmd.Flags().StringSliceVarP(&httpHeaders, "header", "H", nil, "HTTP headers (Key=Value)")

	// Add subcommands
	addCmd.AddCommand(addStdioCmd)
	addCmd.AddCommand(addHttpCmd)
}

func runAddStdio(cmd *cobra.Command, args []string) error {
	command := args[0]
	serverArgs := args[1:]

	// Determine name
	name := stdioName
	if name == "" {
		name = filepath.Base(command)
	}

	// Parse environment variables
	env := make(map[string]string)
	for _, e := range stdioEnv {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			env[parts[0]] = parts[1]
		}
	}

	// Load config
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Create server
	server := config.MCPServer{
		Name:    name,
		Type:    "stdio",
		Command: command,
		Args:    serverArgs,
	}
	if len(env) > 0 {
		server.Env = env
	}

	// Add and save
	if err := cfg.AddServer(server); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Added stdio server %q to %s\n", name, cfg.Path())
	resyncAll(cfg)
	return nil
}

func runAddHttp(cmd *cobra.Command, args []string) error {
	url := args[0]

	// Determine name
	name := httpName
	if name == "" {
		// Extract host from URL as default name
		name = extractHostFromURL(url)
	}

	// Parse headers
	headers := make(map[string]string)
	for _, h := range httpHeaders {
		parts := strings.SplitN(h, "=", 2)
		if len(parts) == 2 {
			headers[parts[0]] = parts[1]
		}
	}

	// Load config
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Create server
	server := config.MCPServer{
		Name: name,
		Type: "http",
		URL:  url,
	}
	if len(headers) > 0 {
		server.Headers = headers
	}

	// Add and save
	if err := cfg.AddServer(server); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Added http server %q to %s\n", name, cfg.Path())
	resyncAll(cfg)
	return nil
}

func loadConfig() (*config.Config, error) {
	if addLocal {
		path, err := config.GetWriteConfigPath(true)
		if err != nil {
			return nil, fmt.Errorf("failed to get config path: %w", err)
		}
		cfg, err := config.LoadFromPath(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
		return cfg, nil
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return cfg, nil
}

func extractHostFromURL(url string) string {
	// Remove protocol
	s := url
	if strings.HasPrefix(s, "https://") {
		s = s[8:]
	} else if strings.HasPrefix(s, "http://") {
		s = s[7:]
	}

	// Take only host part (before first /)
	if idx := strings.Index(s, "/"); idx != -1 {
		s = s[:idx]
	}

	// Remove port
	if idx := strings.Index(s, ":"); idx != -1 {
		s = s[:idx]
	}

	return s
}
