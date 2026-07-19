package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Elydora-Infrastructure/Elydora-Go-SDK/cmd/elydora/plugins"
)

const version = "1.2.1"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		cmdInstall(os.Args[2:])
	case "uninstall":
		cmdUninstall(os.Args[2:])
	case "status":
		cmdStatus(os.Args[2:])
	case "agents":
		cmdAgents()
	case "version", "--version", "-v":
		fmt.Printf("elydora %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Elydora CLI — agent hook installer for tamper-evident audit logging.

Usage:
  elydora <command> [flags]

Commands:
  install      Install Elydora audit hook for an agent
  uninstall    Remove Elydora audit hook for an agent
  status       Show installation status for an agent (or all agents)
  agents       List all supported agents
  version      Print version
  help         Show this help

Run "elydora <command> -h" for details on each command.
`)
}

// ---------------------------------------------------------------------------
// install
// ---------------------------------------------------------------------------

func cmdInstall(args []string) {
	fs := flag.NewFlagSet("install", flag.ExitOnError)
	agent := fs.String("agent", "", "Agent name (required). Use 'elydora agents' to list.")
	orgID := fs.String("org-id", "", "Organization ID (required)")
	agentID := fs.String("agent-id", "", "Agent ID (required)")
	privateKey := fs.String("private-key", "", "Base64url-encoded Ed25519 private key seed (required)")
	kid := fs.String("kid", "", "Key ID (defaults to <agent-id>-key-1)")
	token := fs.String("token", "", "API token for authenticated requests (optional)")
	baseURL := fs.String("base-url", "https://api.elydora.com", "Elydora API base URL")

	fs.Parse(args)

	if *agent == "" || *orgID == "" || *agentID == "" || *privateKey == "" {
		fmt.Fprintln(os.Stderr, "Error: --agent, --org-id, --agent-id, and --private-key are required.")
		fmt.Fprintln(os.Stderr)
		fs.Usage()
		os.Exit(1)
	}

	if *kid == "" {
		*kid = *agentID + "-key-1"
	}

	plugin := plugins.NewPlugin(*agent)
	if plugin == nil {
		fmt.Fprintf(os.Stderr, "Error: unsupported agent %q. Run 'elydora agents' to see supported agents.\n", *agent)
		os.Exit(1)
	}

	// Generate guard script (PreToolUse — freeze enforcement)
	agentDirectory, err := plugins.PrepareAgentRuntimeDirectory(*agentID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	guardScriptPath := filepath.Join(agentDirectory, "guard.js")
	guardScript := plugins.GenerateGuardScript(*agent, *agentID)
	if err := os.WriteFile(guardScriptPath, []byte(guardScript), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing guard script: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Guard script: %s\n", guardScriptPath)

	config := plugins.InstallConfig{
		AgentName:       *agent,
		OrgID:           *orgID,
		AgentID:         *agentID,
		PrivateKey:      *privateKey,
		KID:             *kid,
		Token:           *token,
		BaseURL:         *baseURL,
		GuardScriptPath: guardScriptPath,
	}

	if err := plugin.Install(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// uninstall
// ---------------------------------------------------------------------------

func cmdUninstall(args []string) {
	fs := flag.NewFlagSet("uninstall", flag.ExitOnError)
	agent := fs.String("agent", "", "Agent name (required)")
	agentID := fs.String("agent-id", "", "Agent ID (if omitted, scans ~/.elydora/*/config.json for matching agent_name)")

	fs.Parse(args)

	if *agent == "" {
		fmt.Fprintln(os.Stderr, "Error: --agent is required.")
		fmt.Fprintln(os.Stderr)
		fs.Usage()
		os.Exit(1)
	}

	plugin := plugins.NewPlugin(*agent)
	if plugin == nil {
		fmt.Fprintf(os.Stderr, "Error: unsupported agent %q. Run 'elydora agents' to see supported agents.\n", *agent)
		os.Exit(1)
	}

	runtimeRoot, err := plugins.AgentRuntimeRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if _, err := plugins.RequirePhysicalDirectory(runtimeRoot); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Resolve agentID: if not given, scan ~/.elydora/*/config.json for matching agent_name
	resolvedAgentID := *agentID
	if resolvedAgentID == "" {
		resolvedAgentID, err = findAgentIDByName(*agent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if resolvedAgentID == "" {
			fmt.Fprintf(os.Stderr, "Error: could not find agent ID for %q in ~/.elydora/*/config.json; pass --agent-id explicitly\n", *agent)
			os.Exit(1)
		}
	}
	agentDirectory, err := plugins.ResolveAgentRuntimeDirectory(resolvedAgentID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	agentDirectoryExists, err := plugins.RequirePhysicalDirectory(agentDirectory)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := plugin.Uninstall(resolvedAgentID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Remove the per-agent directory
	if agentDirectoryExists {
		if err := os.RemoveAll(agentDirectory); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing agent directory %s: %v\n", agentDirectory, err)
			os.Exit(1)
		}
		fmt.Printf("  Removed agent directory: %s\n", agentDirectory)
	}
}

// findAgentIDByName scans ~/.elydora/*/config.json for a config whose agent_name
// matches the given name and returns the directory name (agent ID).
func findAgentIDByName(agentName string) (string, error) {
	elydoraDir, err := plugins.AgentRuntimeRoot()
	if err != nil {
		return "", err
	}
	exists, err := plugins.RequirePhysicalDirectory(elydoraDir)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", nil
	}
	entries, err := os.ReadDir(elydoraDir)
	if err != nil {
		return "", fmt.Errorf("read Elydora runtime root %s: %w", elydoraDir, err)
	}
	foundAgentID := ""
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 || !entry.IsDir() {
			continue
		}
		configPath := filepath.Join(elydoraDir, entry.Name(), "config.json")
		data, err := os.ReadFile(configPath)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("read agent config %s: %w", configPath, err)
		}
		var config map[string]interface{}
		if err := json.Unmarshal(data, &config); err != nil {
			return "", fmt.Errorf("parse agent config %s: %w", configPath, err)
		}
		if name, _ := config["agent_name"].(string); name == agentName {
			storedAgentID, ok := config["agent_id"].(string)
			if !ok || storedAgentID == "" {
				return "", fmt.Errorf("agent config %s has an invalid agent_id", configPath)
			}
			if storedAgentID != entry.Name() {
				return "", fmt.Errorf("agent config %s crosses its runtime directory", configPath)
			}
			if _, err := plugins.ResolveAgentRuntimeDirectory(storedAgentID); err != nil {
				return "", fmt.Errorf("validate agent config %s: %w", configPath, err)
			}
			if foundAgentID != "" {
				return "", fmt.Errorf(
					"multiple %q agent runtimes found; pass --agent-id explicitly",
					agentName,
				)
			}
			foundAgentID = storedAgentID
		}
	}
	return foundAgentID, nil
}

// ---------------------------------------------------------------------------
// status
// ---------------------------------------------------------------------------

func cmdStatus(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	agent := fs.String("agent", "", "Agent name (optional — omit to show all)")

	fs.Parse(args)

	if *agent != "" {
		showAgentStatus(*agent)
		return
	}

	// Show status for all agents
	names := sortedAgentNames()
	for _, name := range names {
		showAgentStatus(name)
	}
}

func showAgentStatus(name string) {
	plugin := plugins.NewPlugin(name)
	if plugin == nil {
		fmt.Fprintf(os.Stderr, "Unknown agent: %s\n", name)
		return
	}

	status, err := plugin.Status()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %s: error getting status: %v\n", name, err)
		return
	}

	state := "not installed"
	if status.Installed {
		state = "installed"
	} else if status.HookConfigured && !status.HookScriptExists {
		state = "configured (hook script missing)"
	} else if !status.HookConfigured && status.HookScriptExists {
		state = "hook script exists (not configured)"
	}

	fmt.Printf("  %-14s  %-20s  %s\n", name, status.DisplayName, state)
}

// ---------------------------------------------------------------------------
// agents
// ---------------------------------------------------------------------------

func cmdAgents() {
	fmt.Println("Supported agents:")
	fmt.Println()
	fmt.Printf("  %-14s  %-20s  %s\n", "ID", "Name", "Config")
	fmt.Printf("  %-14s  %-20s  %s\n", strings.Repeat("-", 14), strings.Repeat("-", 20), strings.Repeat("-", 30))

	names := sortedAgentNames()
	for _, name := range names {
		entry := plugins.SupportedAgents[name]
		configPath := entry.ConfigDir + "/" + entry.ConfigFile // display only, not a filesystem path
		fmt.Printf("  %-14s  %-20s  %s\n", name, entry.Name, configPath)
	}
}

func sortedAgentNames() []string {
	names := make([]string, 0, len(plugins.SupportedAgents))
	for name := range plugins.SupportedAgents {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// guardScriptPathForAgent returns the path to ~/.elydora/<agentId>/guard.js.
func guardScriptPathForAgent(agentId string) (string, error) {
	agentDirectory, err := plugins.ResolveAgentRuntimeDirectory(agentId)
	if err != nil {
		return "", err
	}
	return filepath.Join(agentDirectory, "guard.js"), nil
}
