package plugins

// AgentRegistryEntry describes a supported agent integration.
type AgentRegistryEntry struct {
	Name       string
	ConfigDir  string // ~ means home dir; no ~ means project-relative
	ConfigFile string
}

// SupportedAgents is the canonical registry of all supported agent integrations.
var SupportedAgents = map[string]AgentRegistryEntry{
	"claudecode": {Name: "Claude Code", ConfigDir: "~/.claude", ConfigFile: "settings.json"},
	"cursor":     {Name: "Cursor", ConfigDir: ".cursor", ConfigFile: "hooks.json"},
	"gemini":     {Name: "Gemini CLI", ConfigDir: "~/.gemini", ConfigFile: "settings.json"},
	"augment":    {Name: "Augment Code", ConfigDir: "~/.augment", ConfigFile: "settings.json"},
	"kiro":       {Name: "Kiro", ConfigDir: ".kiro/hooks", ConfigFile: "elydora-audit.kiro.hook"},
	"opencode":   {Name: "OpenCode", ConfigDir: ".opencode/plugins", ConfigFile: "elydora-audit.js"},
}

// NewPlugin creates a plugin instance for the given agent name.
// Returns nil if the agent is not supported.
func NewPlugin(agentName string) AgentPlugin {
	switch agentName {
	case "claudecode":
		return &ClaudeCodePlugin{}
	case "cursor":
		return &CursorPlugin{}
	case "gemini":
		return &GeminiPlugin{}
	case "augment":
		return &AugmentPlugin{}
	case "kiro":
		return &KiroPlugin{}
	case "opencode":
		return &OpenCodePlugin{}
	default:
		return nil
	}
}
