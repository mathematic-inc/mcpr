package cmd

import (
	"fmt"
	"strings"

	"github.com/mathematic-inc/mcpr/clients"
	"github.com/mathematic-inc/mcpr/config"

	"github.com/spf13/cobra"
)

var (
	clientSyncServers []string
	clientSyncLocal   bool
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Manage synced clients",
	Long: `Manage which clients are synced with your MCP server configurations.

Subcommands:
  sync   - Sync servers to a client (or resync all)
  remove - Remove a client from the sync list`,
}

var clientSyncCmd = &cobra.Command{
	Use:   "sync [client-name]",
	Short: "Sync MCP servers to a client",
	Long: `Sync MCP server configurations to a specific client.

When called without a client name, it will resync all previously synced clients.

Supported clients:
  - claude-desktop  : Claude Desktop application
  - claude-code     : Claude Code CLI
  - cursor          : Cursor editor
  - windsurf        : Windsurf editor
  - zed             : Zed editor
  - opencode        : OpenCode CLI
  - cline           : Cline VS Code extension
  - vscode          : VS Code with GitHub Copilot
  - continue        : Continue (VS Code/JetBrains)
  - codex           : Codex (OpenAI)
  - gemini          : Gemini CLI (Google)
  - kilo-code       : Kilo Code VS Code extension
  - zencoder        : ZenCoder VS Code extension

The --local flag syncs to project-local config (if supported).

Examples:
  mcpr client sync claude-desktop
  mcpr client sync claude-code --local
  mcpr client sync cursor --servers my-server,another-server
  mcpr client sync  # resync all`,
	Args: cobra.MaximumNArgs(1),
	RunE: runClientSync,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return clients.ListClientNames(), cobra.ShellCompDirectiveNoFileComp
	},
}

var clientRemoveCmd = &cobra.Command{
	Use:   "remove [client-name]",
	Short: "Remove a client from the sync list",
	Long: `Remove a client from the list of synced clients.

This stops the client from being updated when servers are added or removed.
It does not modify the client's current configuration.

Examples:
  mcpr client remove claude-desktop
  mcpr client remove cursor --local`,
	Args: cobra.ExactArgs(1),
	RunE: runClientRemove,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return clients.ListClientNames(), cobra.ShellCompDirectiveNoFileComp
	},
}

func init() {
	clientCmd.AddCommand(clientSyncCmd)
	clientCmd.AddCommand(clientRemoveCmd)

	clientSyncCmd.Flags().StringSliceVarP(&clientSyncServers, "servers", "s", nil, "Specific servers to sync (comma-separated)")
	clientSyncCmd.Flags().BoolVarP(&clientSyncLocal, "local", "l", false, "Sync to project-local config instead of global")
	clientRemoveCmd.Flags().BoolVarP(&clientSyncLocal, "local", "l", false, "Remove project-local sync instead of global")
}

func runClientSync(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// If no client specified, resync all stored clients
	if len(args) == 0 {
		return resyncAll(cfg)
	}

	clientName := args[0]

	// Get the client
	client, err := clients.GetClient(clientName)
	if err != nil {
		return fmt.Errorf("%w\n\nSupported clients: %s", err, strings.Join(clients.ListClientNames(), ", "))
	}

	// Get servers to sync
	var serversToSync []config.MCPServer
	var serverNames []string

	if len(clientSyncServers) > 0 {
		// Sync specific servers
		for _, name := range clientSyncServers {
			server, err := cfg.GetServer(name)
			if err != nil {
				return err
			}
			serversToSync = append(serversToSync, *server)
			serverNames = append(serverNames, name)
		}
	} else {
		// Sync all servers
		serversToSync = cfg.ListServers()
		serverNames = nil // nil means all servers
	}

	if len(serversToSync) == 0 {
		return fmt.Errorf("no servers configured. Use 'mcpr add' to add a server first")
	}

	// Sync to client
	configPath, err := client.Sync(serversToSync, clientSyncLocal)
	if err != nil {
		return fmt.Errorf("failed to sync to %s: %w", client.DisplayName, err)
	}

	// Store synced client info
	cfg.AddSyncedClient(clientName, clientSyncLocal, serverNames)
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save synced client info: %w", err)
	}

	fmt.Printf("Synced %d server(s) to %s\n", len(serversToSync), client.DisplayName)
	fmt.Printf("Config location: %s\n", configPath)
	fmt.Println("\nSynced servers:")
	for _, server := range serversToSync {
		fmt.Printf("  - %s\n", server.Name)
	}

	return nil
}

func runClientRemove(cmd *cobra.Command, args []string) error {
	clientName := args[0]

	// Validate client name
	if _, err := clients.GetClient(clientName); err != nil {
		return fmt.Errorf("%w\n\nSupported clients: %s", err, strings.Join(clients.ListClientNames(), ", "))
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if client is synced
	if cfg.GetSyncedClient(clientName, clientSyncLocal) == nil {
		localStr := ""
		if clientSyncLocal {
			localStr = " (local)"
		}
		return fmt.Errorf("client %q%s is not in the sync list", clientName, localStr)
	}

	// Remove from synced clients
	cfg.RemoveSyncedClient(clientName, clientSyncLocal)
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	localStr := ""
	if clientSyncLocal {
		localStr = " (local)"
	}
	fmt.Printf("Removed %s%s from sync list\n", clientName, localStr)

	return nil
}

func resyncAll(cfg *config.Config) error {
	syncedClients := cfg.GetSyncedClients()
	if len(syncedClients) == 0 {
		fmt.Println("No synced clients. Use 'mcpr client sync <client-name>' to add one.")
		return nil
	}

	var errors []string
	successCount := 0

	for _, sc := range syncedClients {
		client, err := clients.GetClient(sc.Name)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", sc.Name, err))
			continue
		}

		// Get servers to sync
		var serversToSync []config.MCPServer
		if len(sc.Servers) > 0 {
			for _, name := range sc.Servers {
				server, err := cfg.GetServer(name)
				if err != nil {
					errors = append(errors, fmt.Sprintf("%s: server %q not found", sc.Name, name))
					continue
				}
				serversToSync = append(serversToSync, *server)
			}
		} else {
			serversToSync = cfg.ListServers()
		}

		if len(serversToSync) == 0 {
			errors = append(errors, fmt.Sprintf("%s: no servers to sync", sc.Name))
			continue
		}

		// Sync to client
		configPath, err := client.Sync(serversToSync, sc.Local)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", sc.Name, err))
			continue
		}

		localStr := ""
		if sc.Local {
			localStr = " (local)"
		}
		fmt.Printf("✓ %s%s: %d server(s) → %s\n", client.DisplayName, localStr, len(serversToSync), configPath)
		successCount++
	}

	fmt.Printf("\nSynced %d/%d client(s)\n", successCount, len(syncedClients))

	if len(errors) > 0 {
		fmt.Println("\nErrors:")
		for _, e := range errors {
			fmt.Printf("  - %s\n", e)
		}
		return fmt.Errorf("some clients failed to sync")
	}

	return nil
}
