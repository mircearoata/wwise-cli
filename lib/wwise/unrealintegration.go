package wwise

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mircearoata/wwise-cli/lib/unrealengine"
	"github.com/mircearoata/wwise-cli/lib/wwise/client"
	"github.com/mircearoata/wwise-cli/lib/wwise/product"

	cp "github.com/otiai10/copy"
)

func IntegrateWwiseUnreal(uprojectFilePath string, integrationVersion string, wwiseClient *client.WwiseClient) error {
	if filepath.Ext(uprojectFilePath) != ".uproject" {
		return fmt.Errorf("invalid project path: %s", uprojectFilePath)
	}

	if _, err := os.Stat(uprojectFilePath); os.IsNotExist(err) {
		return fmt.Errorf("project path does not exist: %w", err)
	}

	ueIntegrationProduct := product.NewWwiseProduct(wwiseClient, "unrealintegration")

	ueIntegrationVersion, err := ueIntegrationProduct.GetVersion(integrationVersion)
	if err != nil {
		return fmt.Errorf("failed to get unreal integration version: %w", err)
	}

	versionInfo, err := ueIntegrationVersion.GetInfo()
	if err != nil {
		return fmt.Errorf("failed to get wwise manifest: %w", err)
	}

	// Get UE version from project file
	engineRoot, err := unrealengine.GetEngineRootFromProject(uprojectFilePath)
	if err != nil {
		return fmt.Errorf("failed to get engine root: %w", err)
	}

	engineBuild, err := unrealengine.GetEngineVersionData(engineRoot)
	if err != nil {
		return fmt.Errorf("failed to get engine build: %w", err)
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
		return fmt.Errorf("failed to download integration file: %w", err)
	}

	wwiseSDKVersion := fmt.Sprintf("%d.%d.%d.%d", versionInfo.Version.Year, versionInfo.Version.Major, versionInfo.Version.Minor, versionInfo.ProductDependentData.WwiseSdkBuild)

	sdkProduct := product.NewWwiseProduct(wwiseClient, "wwise")
	sdkProductVersion, err := sdkProduct.GetVersion(fmt.Sprintf("wwise.%s", wwiseSDKVersion))
	if err != nil {
		return fmt.Errorf("failed to get sdk version: %w", err)
	}

	type SDKIntegrationAsset struct {
		Source      string
		Destination string
	}

	sdkAssets := []SDKIntegrationAsset{}

	if versionInfo.ProductDependentData.PlatformFolders != nil {
		for _, folder := range versionInfo.ProductDependentData.PlatformFolders.Mandatory {
			if _, err := os.Stat(filepath.Join(sdkProductVersion.Dir, "SDK", folder)); os.IsNotExist(err) {
				return fmt.Errorf("failed to find mandatory folder: %s", folder)
			}
			sdkAssets = append(sdkAssets, SDKIntegrationAsset{
				Source:      folder,
				Destination: folder,
			})
		}

		for _, folder := range versionInfo.ProductDependentData.PlatformFolders.Optional {
			if _, err := os.Stat(filepath.Join(sdkProductVersion.Dir, "SDK", folder)); !os.IsNotExist(err) {
				sdkAssets = append(sdkAssets, SDKIntegrationAsset{
					Source:      folder,
					Destination: folder,
				})
			}
		}
	} else if versionInfo.ProductDependentData.SdkPlatformFolders != nil {
		for _, platformInfos := range *versionInfo.ProductDependentData.SdkPlatformFolders {
			for _, platformInfo := range platformInfos {
				// TODO: Use FileMatchExpression, but for now it seems like it's always "*"

				if platformInfo.SinceEngine != nil {
					major, err := strconv.Atoi(platformInfo.SinceEngine.Major)
					if err != nil {
						return fmt.Errorf("failed to parse major version: %w", err)
					}
					minor, err := strconv.Atoi(platformInfo.SinceEngine.Minor)
					if err != nil {
						return fmt.Errorf("failed to parse minor version: %w", err)
					}
					if engineBuild.MajorVersion < major || (engineBuild.MajorVersion == major && engineBuild.MinorVersion < minor) {
						continue
					}
				}

				if platformInfo.UntilEngine != nil {
					major, err := strconv.Atoi(platformInfo.UntilEngine.Major)
					if err != nil {
						return fmt.Errorf("failed to parse major version: %w", err)
					}
					minor, err := strconv.Atoi(platformInfo.UntilEngine.Minor)
					if err != nil {
						return fmt.Errorf("failed to parse minor version: %w", err)
					}
					if engineBuild.MajorVersion > major || (engineBuild.MajorVersion == major && engineBuild.MinorVersion > minor) {
						continue
					}
				}

				if platformInfo.Optional {
					if _, err := os.Stat(filepath.Join(sdkProductVersion.Dir, "SDK", platformInfo.Source)); !os.IsNotExist(err) {
						sdkAssets = append(sdkAssets, SDKIntegrationAsset{
							Source:      platformInfo.Source,
							Destination: platformInfo.Destination,
						})
					}
				} else {
					if _, err := os.Stat(filepath.Join(sdkProductVersion.Dir, "SDK", platformInfo.Source)); os.IsNotExist(err) {
						return fmt.Errorf("failed to find mandatory folder: %s", platformInfo.Source)
					}
					sdkAssets = append(sdkAssets, SDKIntegrationAsset{
						Source:      platformInfo.Source,
						Destination: platformInfo.Destination,
					})
				}
			}
		}
	} else {
		return errors.New("failed to find platform folders")
	}

	integrationAssets, err := os.ReadDir(ueIntegrationVersion.Dir)
	if err != nil {
		return fmt.Errorf("failed to read integration cache path: %w", err)
	}

	projectRoot := filepath.Dir(uprojectFilePath)
	for _, entry := range integrationAssets {
		if entry.IsDir() {
			err = cp.Copy(filepath.Join(ueIntegrationVersion.Dir, entry.Name()), filepath.Join(projectRoot, "Plugins", entry.Name()))
			if err != nil {
				return fmt.Errorf("failed to copy integration asset: %s: %w", entry.Name(), err)
			}
		}
	}

	mainWwisePlugin := filepath.Join(projectRoot, "Plugins", "Wwise")
	for _, sdkAsset := range sdkAssets {
		err = cp.Copy(filepath.Join(sdkProductVersion.Dir, "SDK", sdkAsset.Source), filepath.Join(mainWwisePlugin, "ThirdParty", sdkAsset.Destination))
		if err != nil {
			return fmt.Errorf("failed to copy third party files: %w", err)
		}
	}

	return nil
}
