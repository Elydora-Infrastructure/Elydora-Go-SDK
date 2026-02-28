package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CursorPlugin manages the Elydora audit hook for Cursor.
// It writes/merges into ~/.cursor/hooks.json using nested settings.hooks.postToolUse[].
type CursorPlugin struct{}

func (p *CursorPlugin) configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".cursor", "hooks.json"), nil
}

func (p *CursorPlugin) Install(config InstallConfig) error {
	scriptPath, err := hookScriptPath("cursor")
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
		guardPath, err = guardScriptPath("cursor")
		if err != nil {
			return err
		}
	}

	configPath, err := p.configPath()
	if err != nil {
		return err
	}

	settings, err := readJSONFile(configPath)
	if err != nil {
		return err
	}

	// Ensure hooks object exists
	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = make(map[string]interface{})
	}

	// --- preToolUse (guard — freeze enforcement) ---
	preToolUse, _ := hooks["preToolUse"].([]interface{})
	var preFiltered []interface{}
	for _, entry := range preToolUse {
		if m, ok := entry.(map[string]interface{}); ok {
			if cmd, _ := m["command"].(string); strings.Contains(cmd, "elydora") {
				continue
			}
		}
		preFiltered = append(preFiltered, entry)
	}
	guardEntry := map[string]interface{}{
		"command": "node " + guardPath,
	}
	preFiltered = append(preFiltered, guardEntry)
	hooks["preToolUse"] = preFiltered

	// --- postToolUse (audit logging) ---
	postToolUse, _ := hooks["postToolUse"].([]interface{})
	var postFiltered []interface{}
	for _, entry := range postToolUse {
		if m, ok := entry.(map[string]interface{}); ok {
			if cmd, _ := m["command"].(string); strings.Contains(cmd, "elydora") {
				continue
			}
		}
		postFiltered = append(postFiltered, entry)
	}
	hookEntry := map[string]interface{}{
		"command": "node " + scriptPath,
	}
	postFiltered = append(postFiltered, hookEntry)
	hooks["postToolUse"] = postFiltered

	settings["hooks"] = hooks

	if err := writeJSONFile(configPath, settings); err != nil {
		return err
	}
	fmt.Printf("Installed Elydora hook for Cursor at %s\n", configPath)
	return nil
}

func (p *CursorPlugin) Uninstall() error {
	scriptPath, err := hookScriptPath("cursor")
	if err != nil {
		return err
	}

	configPath, err := p.configPath()
	if err != nil {
		return err
	}

	settings, err := readJSONFile(configPath)
	if err != nil {
		return err
	}

	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		fmt.Println("No Cursor hooks found.")
		return nil
	}

	// Remove preToolUse Elydora entries
	preToolUse, _ := hooks["preToolUse"].([]interface{})
	var preFiltered []interface{}
	for _, entry := range preToolUse {
		if m, ok := entry.(map[string]interface{}); ok {
			if cmd, _ := m["command"].(string); strings.Contains(cmd, "elydora") {
				continue
			}
		}
		preFiltered = append(preFiltered, entry)
	}
	if len(preFiltered) == 0 {
		delete(hooks, "preToolUse")
	} else {
		hooks["preToolUse"] = preFiltered
	}

	// Remove postToolUse Elydora entries
	postToolUse, _ := hooks["postToolUse"].([]interface{})
	var postFiltered []interface{}
	for _, entry := range postToolUse {
		if m, ok := entry.(map[string]interface{}); ok {
			if cmd, _ := m["command"].(string); strings.Contains(cmd, "elydora") {
				continue
			}
		}
		postFiltered = append(postFiltered, entry)
	}
	if len(postFiltered) == 0 {
		delete(hooks, "postToolUse")
	} else {
		hooks["postToolUse"] = postFiltered
	}

	if len(hooks) == 0 {
		delete(settings, "hooks")
	} else {
		settings["hooks"] = hooks
	}

	if err := writeJSONFile(configPath, settings); err != nil {
		return err
	}

	os.Remove(scriptPath)
	gPath, _ := guardScriptPath("cursor")
	if gPath != "" {
		os.Remove(gPath)
	}
	fmt.Println("Uninstalled Elydora hook for Cursor.")
	return nil
}

func (p *CursorPlugin) Status() (PluginStatus, error) {
	scriptPath, err := hookScriptPath("cursor")
	if err != nil {
		return PluginStatus{}, err
	}

	configPath, err := p.configPath()
	if err != nil {
		return PluginStatus{}, err
	}

	status := PluginStatus{
		AgentName:   "cursor",
		DisplayName: "Cursor",
		ConfigPath:  configPath,
	}

	if _, err := os.Stat(scriptPath); err == nil {
		status.HookScriptExists = true
	}

	settings, err := readJSONFile(configPath)
	if err != nil {
		return status, nil
	}

	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks != nil {
		preConfigured := hasCursorElydoraEntry(hooks["preToolUse"])
		postConfigured := hasCursorElydoraEntry(hooks["postToolUse"])
		status.HookConfigured = preConfigured && postConfigured
	}

	status.Installed = status.HookConfigured && status.HookScriptExists
	return status, nil
}

// hasCursorElydoraEntry checks if a Cursor hook array (flat format) contains an Elydora entry.
func hasCursorElydoraEntry(hookArray interface{}) bool {
	arr, _ := hookArray.([]interface{})
	for _, entry := range arr {
		if m, ok := entry.(map[string]interface{}); ok {
			if cmd, _ := m["command"].(string); strings.Contains(cmd, "elydora") {
				return true
			}
		}
	}
	return false
}
