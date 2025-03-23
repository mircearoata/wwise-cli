package wwise

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mircearoata/wwise-cli/lib/unrealengine"
	"github.com/mircearoata/wwise-cli/lib/wwise/client"
	"github.com/mircearoata/wwise-cli/lib/wwise/product"
	"github.com/pkg/errors"

	cp "github.com/otiai10/copy"
)

func IntegrateWwiseUnreal(uprojectFilePath string, integrationVersion string, wwiseClient *client.WwiseClient) error {
	if filepath.Ext(uprojectFilePath) != ".uproject" {
		return errors.New("invalid project path: " + uprojectFilePath)
	}

	if _, err := os.Stat(uprojectFilePath); os.IsNotExist(err) {
		return errors.Wrap(err, "project path does not exist")
	}

	ueIntegrationProduct := product.NewWwiseProduct(wwiseClient, "unrealintegration")

	ueIntegrationVersion, err := ueIntegrationProduct.GetVersion(integrationVersion)
	if err != nil {
		return errors.Wrap(err, "failed to get unreal integration version")
	}

	versionInfo, err := ueIntegrationVersion.GetInfo()
	if err != nil {
		return errors.Wrap(err, "failed to get wwise manifest")
	}

	// Get UE version from project file
	engineRoot, err := unrealengine.GetEngineRootFromProject(uprojectFilePath)
	if err != nil {
		return errors.Wrap(err, "failed to get engine root")
	}

	engineBuild, err := unrealengine.GetEngineVersionData(engineRoot)
	if err != nil {
		return errors.Wrap(err, "failed to get engine build")
	}

	wwiseUEDeploymentPlatform := fmt.Sprintf("UE%d%d", engineBuild.MajorVersion, engineBuild.MinorVersion)

	integrationFiles := versionInfo.FindFilesByGroups([]product.GroupFilter{
		{GroupID: "DeploymentPlatforms", GroupValues: []string{wwiseUEDeploymentPlatform}},
		{GroupID: "Packages", GroupValues: []string{"Unreal"}},
	})

	if len(integrationFiles) == 0 {
		return errors.New("failed to find integration file")
	}

	if len(integrationFiles) > 1 {
		return errors.New("found more than one integration file")
	}

	err = ueIntegrationVersion.DownloadOrCache(integrationFiles[0])
	if err != nil {
		return errors.Wrap(err, "failed to download integration file")
	}

	wwiseSDKVersion := fmt.Sprintf("%d.%d.%d.%d", versionInfo.Version.Year, versionInfo.Version.Major, versionInfo.Version.Minor, versionInfo.ProductDependentData.WwiseSdkBuild)

	sdkProduct := product.NewWwiseProduct(wwiseClient, "wwise")
	sdkProductVersion, err := sdkProduct.GetVersion(fmt.Sprintf("wwise.%s", wwiseSDKVersion))
	if err != nil {
		return errors.Wrap(err, "failed to get sdk version")
	}

	foldersToCopyThirdParty := []string{}

	for _, folder := range versionInfo.ProductDependentData.PlatformFolders.Mandatory {
		if _, err := os.Stat(filepath.Join(sdkProductVersion.Dir, "SDK", folder)); os.IsNotExist(err) {
			return errors.New("failed to find mandatory folder: " + folder)
		}
		foldersToCopyThirdParty = append(foldersToCopyThirdParty, folder)
	}

	for _, folder := range versionInfo.ProductDependentData.PlatformFolders.Optional {
		if _, err := os.Stat(filepath.Join(sdkProductVersion.Dir, "SDK", folder)); !os.IsNotExist(err) {
			foldersToCopyThirdParty = append(foldersToCopyThirdParty, folder)
		}
	}

	integrationAssets, err := os.ReadDir(ueIntegrationVersion.Dir)
	if err != nil {
		return errors.Wrap(err, "failed to read integration cache path")
	}

	projectRoot := filepath.Dir(uprojectFilePath)
	for _, entry := range integrationAssets {
		if entry.IsDir() {
			err = cp.Copy(filepath.Join(ueIntegrationVersion.Dir, entry.Name()), filepath.Join(projectRoot, "Plugins", entry.Name()))
			if err != nil {
				return errors.Wrapf(err, "failed to copy integration asset: %s", entry.Name())
			}
		}
	}

	mainWwisePlugin := filepath.Join(projectRoot, "Plugins", "Wwise")
	for _, folder := range foldersToCopyThirdParty {
		err = cp.Copy(filepath.Join(sdkProductVersion.Dir, "SDK", folder), filepath.Join(mainWwisePlugin, "ThirdParty", folder))
		if err != nil {
			return errors.Wrap(err, "failed to copy third party files")
		}
	}

	return nil
}
