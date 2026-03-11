package cmd

import (
	"fmt"

	"github.com/mathematic-inc/mcpr/config"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove [server-name]",
	Aliases: []string{"rm"},
	Short:   "Remove an MCP server from configuration",
	Long: `Remove an MCP server from your configuration.

This removes the server from your mcpr config. If the daemon is running,
it will automatically resync all clients to reflect the change.

Examples:
  # Remove a server
  mcpr remove my-server

  # Using the alias
  mcpr rm my-server`,
	Args: cobra.ExactArgs(1),
	RunE: runRemove,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		// Load config and return server names for completion
		cfg, err := config.Load()
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		var names []string
		for _, s := range cfg.ListServers() {
			names = append(names, s.Name)
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	},
}

func runRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Remove server
	if err := cfg.RemoveServer(name); err != nil {
		return err
	}

	// Save config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Removed server %q from %s\n", name, cfg.Path())
	resyncAll(cfg)
	return nil
}
