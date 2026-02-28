package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// KiroPlugin manages the Elydora audit hook for Kiro.
// It writes a .kiro/hooks/elydora-audit.kiro.hook JSON file in the home directory.
type KiroPlugin struct{}

func (p *KiroPlugin) configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".kiro", "hooks", "elydora-audit.kiro.hook"), nil
}

func (p *KiroPlugin) Install(config InstallConfig) error {
	scriptPath, err := hookScriptPath("kiro")
	if err != nil {
		return err
	}
	if config.HookScript != "" {
		scriptPath = config.HookScript
	}

	if err := GenerateHookScript(scriptPath, config); err != nil {
		return fmt.Errorf("generate hook script: %w", err)
	}

	guardPath := config.GuardScriptPath
	if guardPath == "" {
		guardPath, err = guardScriptPath("kiro")
		if err != nil {
			return err
		}
	}

	hookFile, err := p.configPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(hookFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	content := buildKiroHookFile(scriptPath, guardPath)
	if err := os.WriteFile(hookFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("write %s: %w", hookFile, err)
	}

	fmt.Printf("Installed Elydora hook for Kiro at %s\n", hookFile)
	return nil
}

func buildKiroHookFile(scriptPath, guardPath string) string {
	hookConfig := map[string]interface{}{
		"name": "Elydora Audit",
		"hooks": map[string]interface{}{
			"pre_tool_use": map[string]interface{}{
				"command":    "node " + guardPath,
				"timeout_ms": 5000,
			},
			"post_tool_use": map[string]interface{}{
				"command":    "node " + scriptPath,
				"timeout_ms": 5000,
			},
		},
	}
	encoded, _ := json.MarshalIndent(hookConfig, "", "  ")
	return string(encoded) + "\n"
}

func (p *KiroPlugin) Uninstall() error {
	hookFile, err := p.configPath()
	if err != nil {
		return err
	}
	if err := os.Remove(hookFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove %s: %w", hookFile, err)
	}

	scriptPath, _ := hookScriptPath("kiro")
	if scriptPath != "" {
		os.Remove(scriptPath)
	}
	gPath, _ := guardScriptPath("kiro")
	if gPath != "" {
		os.Remove(gPath)
	}

	fmt.Println("Uninstalled Elydora hook for Kiro.")
	return nil
}

func (p *KiroPlugin) Status() (PluginStatus, error) {
	scriptPath, err := hookScriptPath("kiro")
	if err != nil {
		return PluginStatus{}, err
	}

	hookFile, err := p.configPath()
	if err != nil {
		return PluginStatus{}, err
	}

	status := PluginStatus{
		AgentName:   "kiro",
		DisplayName: "Kiro",
		ConfigPath:  hookFile,
	}

	if _, err := os.Stat(scriptPath); err == nil {
		status.HookScriptExists = true
	}

	// Check if hook file exists and contains both pre_tool_use and post_tool_use
	data, err := os.ReadFile(hookFile)
	if err == nil {
		var config map[string]interface{}
		if json.Unmarshal(data, &config) == nil {
			hooks, _ := config["hooks"].(map[string]interface{})
			if hooks != nil {
				_, hasPre := hooks["pre_tool_use"]
				_, hasPost := hooks["post_tool_use"]
				status.HookConfigured = hasPre && hasPost
			}
		}
	}

	status.Installed = status.HookConfigured && status.HookScriptExists
	return status, nil
}
