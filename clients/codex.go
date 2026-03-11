package clients

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mathematic-inc/mcpr/config"
)

// Path functions as variables for testing
var (
	getCodexConfigPath = getCodexConfigPathImpl
	getCodexLocalPath  = getCodexLocalPathImpl
)

func init() {
	RegisterClient(&Client{
		Name:          "codex",
		DisplayName:   "Codex (OpenAI)",
		GlobalPath:    func() (string, error) { return getCodexConfigPath() },
		LocalPath:     func() (string, error) { return getCodexLocalPath() },
		SupportsLocal: true,
		SyncFunc:      syncToCodex,
	})
}

func getCodexConfigPathImpl() (string, error) {
	codexHome := codexHomeDir()
	return filepath.Join(codexHome, "config.toml"), nil
}

func getCodexLocalPathImpl() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if err := validateCodexProjectTrust(cwd); err != nil {
		return "", err
	}
	return filepath.Join(cwd, ".codex", "config.toml"), nil
}

func codexHomeDir() string {
	if codexHome := os.Getenv("CODEX_HOME"); codexHome != "" {
		return codexHome
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ".codex"
	}
	return filepath.Join(home, ".codex")
}

func validateCodexProjectTrust(projectPath string) error {
	globalConfigPath := filepath.Join(codexHomeDir(), "config.toml")
	trustLevel, err := readCodexProjectTrustLevel(globalConfigPath, projectPath)
	if err != nil {
		return err
	}
	if trustLevel != "trusted" {
		return fmt.Errorf(
			"project %q is not trusted in Codex\n\nTo enable local configuration, add the following to %s:\n\n[projects.%q]\ntrust_level = \"trusted\"",
			projectPath, globalConfigPath, projectPath,
		)
	}
	return nil
}

// readCodexProjectTrustLevel parses configPath for the trust_level of the given project.
// Returns "" if the project section or key is not found.
func readCodexProjectTrustLevel(configPath, projectPath string) (string, error) {
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to read Codex config: %w", err)
	}

	// Build the expected section header, e.g. [projects."/home/user/project"]
	sectionHeader := fmt.Sprintf(`[projects.%q]`, projectPath)

	inSection := false
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == sectionHeader {
			inSection = true
			continue
		}
		if inSection {
			if strings.HasPrefix(trimmed, "[") {
				break
			}
			if strings.HasPrefix(trimmed, "trust_level") {
				parts := strings.SplitN(trimmed, "=", 2)
				if len(parts) == 2 {
					val := strings.TrimSpace(parts[1])
					val = strings.Trim(val, `"'`)
					return val, nil
				}
			}
		}
	}
	return "", nil
}

func syncToCodex(servers []config.MCPServer, path string) error {
	var existingContent string
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		existingContent = ""
	} else if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	} else {
		existingContent = string(data)
	}

	// Parse existing content and remove existing [mcp_servers.*] sections
	lines := tomlSplitLines(existingContent)
	var filteredLines []string
	inMcpSection := false

	for _, line := range lines {
		trimmed := tomlTrimWhitespace(line)
		if tomlHasPrefix(trimmed, "[mcp_servers.") {
			inMcpSection = true
			continue
		}
		if inMcpSection && tomlHasPrefix(trimmed, "[") {
			inMcpSection = false
		}
		if !inMcpSection {
			filteredLines = append(filteredLines, line)
		}
	}

	// Build new MCP servers sections
	var mcpSections []string
	for _, server := range servers {
		if server.Type == "http" {
			section := fmt.Sprintf("[mcp_servers.%s]\nurl = %q\n", server.Name, server.URL)
			if len(server.Headers) > 0 {
				section += "http_headers = { "
				headerKeys := make([]string, 0, len(server.Headers))
				for k := range server.Headers {
					headerKeys = append(headerKeys, k)
				}
				sort.Strings(headerKeys)
				for i, k := range headerKeys {
					if i > 0 {
						section += ", "
					}
					section += fmt.Sprintf("%q = %q", k, server.Headers[k])
				}
				section += " }\n"
			}
			mcpSections = append(mcpSections, section)
		} else {
			section := fmt.Sprintf("[mcp_servers.%s]\ncommand = %q\n", server.Name, server.Command)
			if len(server.Args) > 0 {
				section += "args = ["
				for i, arg := range server.Args {
					if i > 0 {
						section += ", "
					}
					section += fmt.Sprintf("%q", arg)
				}
				section += "]\n"
			}
			if len(server.Env) > 0 {
				section += "env = { "
				envKeys := make([]string, 0, len(server.Env))
				for k := range server.Env {
					envKeys = append(envKeys, k)
				}
				sort.Strings(envKeys)
				for i, k := range envKeys {
					if i > 0 {
						section += ", "
					}
					section += fmt.Sprintf("%q = %q", k, server.Env[k])
				}
				section += " }\n"
			}
			mcpSections = append(mcpSections, section)
		}
	}

	// Combine filtered content with new MCP sections
	result := tomlJoinLines(filteredLines)
	if len(mcpSections) > 0 {
		if result != "" && !tomlHasSuffix(result, "\n\n") {
			if tomlHasSuffix(result, "\n") {
				result += "\n"
			} else {
				result += "\n\n"
			}
		}
		for _, section := range mcpSections {
			result += section + "\n"
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return os.WriteFile(path, []byte(result), 0o644)
}

// TOML helper functions

func tomlSplitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func tomlJoinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}

func tomlTrimWhitespace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

func tomlHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func tomlHasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
