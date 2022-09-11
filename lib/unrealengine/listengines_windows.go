package unrealengine

import (
	"github.com/pkg/errors"
	"golang.org/x/sys/windows/registry"
)

func listEngines() (map[string]string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Epic Games\Unreal Engine\Builds`, registry.QUERY_VALUE)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open registry key")
	}
	defer k.Close()

	engines := make(map[string]string)

	values, err := k.ReadValueNames(0)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read registry values")
	}

	for _, value := range values {
		path, _, err := k.GetStringValue(value)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get registry value")
		}
		engines[value] = path
	}

	return engines, nil
}
