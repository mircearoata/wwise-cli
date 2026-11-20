package unrealengine

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

type UProject struct {
	EngineAssociation string `json:"EngineAssociation"`
}

func getEngineAssociationFromProject(projectPath string) (string, error) {
	projectData, err := os.ReadFile(projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to read project file: %w", err)
	}

	var uproject UProject
	if err := json.Unmarshal(projectData, &uproject); err != nil {
		return "", fmt.Errorf("failed to unmarshal project file: %w", err)
	}

	return uproject.EngineAssociation, nil
}

func GetEngineRootFromProject(projectPath string) (string, error) {
	engineAssociation, err := getEngineAssociationFromProject(projectPath)
	if err != nil {
		return "", fmt.Errorf("failed to get engine association: %w", err)
	}

	engines, err := listEngines()
	if err != nil {
		return "", fmt.Errorf("failed to list engines: %w", err)
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
		return EngineBuildFile{}, fmt.Errorf("failed to read build version file: %w", err)
	}
	encoding, _, _, err := determineEncodingFromReader(bytes.NewReader(buildVersionData), len(buildVersionData))
	if err != nil {
		return EngineBuildFile{}, fmt.Errorf("failed to determine encoding: %w", err)
	}

	reader := transform.NewReader(bytes.NewReader(buildVersionData), encoding.NewDecoder())

	var buildVersion EngineBuildFile
	if err := json.NewDecoder(reader).Decode(&buildVersion); err != nil {
		return EngineBuildFile{}, fmt.Errorf("failed to unmarshal build version file: %w", err)
	}

	return buildVersion, nil
}

func determineEncodingFromReader(r io.Reader, size int) (e encoding.Encoding, name string, certain bool, err error) {
	peekSize := 1024
	if size < peekSize {
		peekSize = size
	}
	b, err := bufio.NewReader(r).Peek(peekSize)
	if err != nil {
		return
	}

	e, name, certain = charset.DetermineEncoding(b, "")
	return
}
