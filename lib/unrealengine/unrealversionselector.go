package unrealengine

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type UProject struct {
	EngineAssociation string `json:"EngineAssociation"`
}

func getEngineAssociationFromProject(projectPath string) (string, error) {
	projectData, err := os.ReadFile(projectPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to read project file")
	}

	var uproject UProject
	if err := json.Unmarshal(projectData, &uproject); err != nil {
		return "", errors.Wrap(err, "failed to unmarshal project file")
	}

	return uproject.EngineAssociation, nil
}

func GetEngineRootFromProject(projectPath string) (string, error) {
	engineAssociation, err := getEngineAssociationFromProject(projectPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to get engine association")
	}

	engines, err := listEngines()
	if err != nil {
		return "", errors.Wrap(err, "failed to list engines")
	}

	engineRoot, ok := engines[engineAssociation]
	if !ok {
		return "", errors.New("failed to find engine")
	}

	return engineRoot, nil
}

type EngineBuildFile struct {
	MajorVersion         int    `json:"MajorVersion"`
	MinorVersion         int    `json:"MinorVersion"`
	PatchVersion         int    `json:"PatchVersion"`
	Changelist           int    `json:"Changelist"`
	CompatibleChangelist int    `json:"CompatibleChangelist"`
	IsLicenseeVersion    int    `json:"IsLicenseeVersion"`
	IsPromotedBuild      int    `json:"IsPromotedBuild"`
	BranchName           string `json:"BranchName"`
}

func GetEngineVersionData(enginePath string) (EngineBuildFile, error) {
	buildVersionFilePath := filepath.Join(enginePath, "Engine", "Build", "Build.version")

	buildVersionData, err := os.ReadFile(buildVersionFilePath)
	if err != nil {
		return EngineBuildFile{}, errors.Wrap(err, "failed to read build version file")
	}

	reader := transform.NewReader(bytes.NewReader(buildVersionData), unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder())

	var buildVersion EngineBuildFile
	if err := json.NewDecoder(reader).Decode(&buildVersion); err != nil {
		return EngineBuildFile{}, errors.Wrap(err, "failed to unmarshal build version file")
	}

	return buildVersion, nil
}
