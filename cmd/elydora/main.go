package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/elydora/sdk-go/cmd/elydora/plugins"
)

const version = "1.0.0"

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
	token := fs.String("token", "", "JWT token for authenticated requests (optional)")
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
	guardScriptPath, err := guardScriptPathForAgent(*agent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	guardScript := plugins.GenerateGuardScript(*agent)
	if err := os.MkdirAll(filepath.Dir(guardScriptPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating guard script directory: %v\n", err)
		os.Exit(1)
	}
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

	if err := plugin.Uninstall(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Also remove the guard script
	guardPath, _ := guardScriptPathForAgent(*agent)
	if guardPath != "" {
		os.Remove(guardPath)
	}
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
		configPath := entry.ConfigDir + "/" + entry.ConfigFile
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

// guardScriptPathForAgent returns the path to ~/.elydora/hooks/<agent>-guard.js.
func guardScriptPathForAgent(agentName string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".elydora", "hooks", agentName+"-guard.js"), nil
}
