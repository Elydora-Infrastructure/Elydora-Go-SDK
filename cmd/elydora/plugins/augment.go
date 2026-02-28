package plugins

import (
	"fmt"
	"os"
)

// AugmentPlugin manages the Elydora audit hook for Augment Code.
// It merges a PostToolUse hook into ~/.augment/settings.json.
type AugmentPlugin struct{}

func (p *AugmentPlugin) Install(config InstallConfig) error {
	scriptPath, err := hookScriptPath("augment")
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
		guardPath, err = guardScriptPath("augment")
		if err != nil {
			return err
		}
	}

	configDir, err := expandHome("~/.augment")
	if err != nil {
		return err
	}
	configPath := configDir + "/settings.json"

	settings, err := readJSONFile(configPath)
	if err != nil {
		return err
	}

	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		hooks = make(map[string]interface{})
	}

	// --- PreToolUse (guard — freeze enforcement) ---
	preToolUse, _ := hooks["PreToolUse"].([]interface{})
	var preFiltered []interface{}
	for _, entry := range preToolUse {
		if m, ok := entry.(map[string]interface{}); ok {
			if isElydoraHookEntry(m) {
				continue
			}
		}
		preFiltered = append(preFiltered, entry)
	}
	guardEntry := map[string]interface{}{
		"hooks": []interface{}{
			map[string]interface{}{
				"type":    "command",
				"command": "node " + guardPath,
			},
		},
	}
	preFiltered = append(preFiltered, guardEntry)
	hooks["PreToolUse"] = preFiltered

	// --- PostToolUse (audit logging) ---
	postToolUse, _ := hooks["PostToolUse"].([]interface{})
	var postFiltered []interface{}
	for _, entry := range postToolUse {
		if m, ok := entry.(map[string]interface{}); ok {
			if isElydoraHookEntry(m) {
				continue
			}
		}
		postFiltered = append(postFiltered, entry)
	}
	hookEntry := map[string]interface{}{
		"hooks": []interface{}{
			map[string]interface{}{
				"type":    "command",
				"command": "node " + scriptPath,
			},
		},
	}
	postFiltered = append(postFiltered, hookEntry)
	hooks["PostToolUse"] = postFiltered

	settings["hooks"] = hooks

	if err := writeJSONFile(configPath, settings); err != nil {
		return err
	}
	fmt.Printf("Installed Elydora hook for Augment Code at %s\n", configPath)
	return nil
}

func (p *AugmentPlugin) Uninstall() error {
	configDir, err := expandHome("~/.augment")
	if err != nil {
		return err
	}
	configPath := configDir + "/settings.json"

	settings, err := readJSONFile(configPath)
	if err != nil {
		return err
	}

	hooks, _ := settings["hooks"].(map[string]interface{})
	if hooks == nil {
		fmt.Println("No Augment Code hooks found.")
		return nil
	}

	// Remove PreToolUse Elydora entries
	preToolUse, _ := hooks["PreToolUse"].([]interface{})
	var preFiltered []interface{}
	for _, entry := range preToolUse {
		if m, ok := entry.(map[string]interface{}); ok {
			if isElydoraHookEntry(m) {
				continue
			}
		}
		preFiltered = append(preFiltered, entry)
	}
	if len(preFiltered) == 0 {
		delete(hooks, "PreToolUse")
	} else {
		hooks["PreToolUse"] = preFiltered
	}

	// Remove PostToolUse Elydora entries
	postToolUse, _ := hooks["PostToolUse"].([]interface{})
	var postFiltered []interface{}
	for _, entry := range postToolUse {
		if m, ok := entry.(map[string]interface{}); ok {
			if isElydoraHookEntry(m) {
				continue
			}
		}
		postFiltered = append(postFiltered, entry)
	}
	if len(postFiltered) == 0 {
		delete(hooks, "PostToolUse")
	} else {
		hooks["PostToolUse"] = postFiltered
	}

	if len(hooks) == 0 {
		delete(settings, "hooks")
	} else {
		settings["hooks"] = hooks
	}

	if err := writeJSONFile(configPath, settings); err != nil {
		return err
	}

	scriptPath, _ := hookScriptPath("augment")
	if scriptPath != "" {
		os.Remove(scriptPath)
	}
	gPath, _ := guardScriptPath("augment")
	if gPath != "" {
		os.Remove(gPath)
	}
	fmt.Println("Uninstalled Elydora hook for Augment Code.")
	return nil
}

func (p *AugmentPlugin) Status() (PluginStatus, error) {
	scriptPath, err := hookScriptPath("augment")
	if err != nil {
		return PluginStatus{}, err
	}

	configDir, err := expandHome("~/.augment")
	if err != nil {
		return PluginStatus{}, err
	}
	configPath := configDir + "/settings.json"

	status := PluginStatus{
		AgentName:   "augment",
		DisplayName: "Augment Code",
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
		preConfigured := hasElydoraEntry(hooks["PreToolUse"])
		postConfigured := hasElydoraEntry(hooks["PostToolUse"])
		status.HookConfigured = preConfigured && postConfigured
	}

	status.Installed = status.HookConfigured && status.HookScriptExists
	return status, nil
}
