package cmd

import (
	"fmt"
	"strings"

	"github.com/mathematic-inc/mcpr/clients"
	"github.com/mathematic-inc/mcpr/config"

	"github.com/spf13/cobra"
)

var listClients bool

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured MCP servers or supported clients",
	Long: `List all configured MCP servers or supported clients.

Examples:
  # List all configured servers
  mcpr list

  # List supported clients
  mcpr list --clients`,
	RunE: runList,
}

func init() {
	listCmd.Flags().BoolVarP(&listClients, "clients", "c", false, "List supported clients instead of servers")
}

func runList(cmd *cobra.Command, args []string) error {
	if listClients {
		return listSupportedClients()
	}
	return listServers()
}

func listServers() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	servers := cfg.ListServers()
	if len(servers) == 0 {
		fmt.Println("No servers configured.")
		fmt.Println("Use 'mcpr add' to add a server.")
		return nil
	}

	fmt.Printf("Configured servers (from %s):\n\n", cfg.Path())
	for _, server := range servers {
		fmt.Printf("  %s\n", server.Name)
		fmt.Printf("    Command: %s\n", server.Command)
		if len(server.Args) > 0 {
			fmt.Printf("    Args:    %s\n", strings.Join(server.Args, " "))
		}
		if len(server.Env) > 0 {
			envPairs := make([]string, 0, len(server.Env))
			for k, v := range server.Env {
				envPairs = append(envPairs, fmt.Sprintf("%s=%s", k, v))
			}
			fmt.Printf("    Env:     %s\n", strings.Join(envPairs, ", "))
		}
		fmt.Println()
	}

	return nil
}

func listSupportedClients() error {
	fmt.Println("Supported MCP clients:")
	fmt.Println()
	for name, client := range clients.GetClients() {
		path, _ := client.ConfigPath()
		fmt.Printf("  %s (%s)\n", name, client.DisplayName)
		fmt.Printf("    Config: %s\n", path)
		fmt.Println()
	}
	return nil
}
