package unrealengine

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func listEngines() (map[string]string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Epic Games\Unreal Engine\Builds`, registry.QUERY_VALUE)
	if err != nil {
		return nil, fmt.Errorf("failed to open registry key: %w", err)
	}
	defer k.Close()

	engines := make(map[string]string)

	values, err := k.ReadValueNames(0)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry values: %w", err)
	}

	for _, value := range values {
		path, _, err := k.GetStringValue(value)
		if err != nil {
			return nil, fmt.Errorf("failed to get registry value: %w", err)
		}
		engineVersion := value
		if strings.HasPrefix(engineVersion, "UE_") {
			engineVersion = strings.Replace(engineVersion, "UE_", "", 1)
		}
		engines[engineVersion] = path
	}

	return engines, nil
}
