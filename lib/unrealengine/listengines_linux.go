package unrealengine

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

func listEngines() (map[string]string, error) {
	userConfig, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user config dir: %w", err)
	}

	installsFile := filepath.Join(userConfig, "Epic", "UnrealEngine", "Install.ini")
	cfg, err := ini.Load(installsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load installs file: %w", err)
	}

	installationsSection := cfg.Section("Installations")
	if installationsSection == nil {
		return nil, fmt.Errorf("failed to find installations section")
	}

	engines := make(map[string]string)
	for _, key := range installationsSection.KeyStrings() {
		enginePath := installationsSection.Key(key).String()
		engines[key] = enginePath
	}
	return engines, nil
}
