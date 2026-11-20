package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mircearoata/wwise-cli/lib/wwise/product"
	"github.com/mircearoata/wwise-cli/utils"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a version of wwise",
	RunE: func(cmd *cobra.Command, args []string) error {
		sdkVersion := viper.GetString("sdk-version")

		filters := viper.GetStringSlice("filter")
		filterMap := make(map[string][]string)
		for _, filter := range filters {
			parts := strings.Split(filter, "=")
			if len(parts) == 1 {
				parts = append(parts, "")
			}
			if len(parts) != 2 {
				return errors.New("invalid filter format. use key=value")
			}
			filterMap[parts[0]] = append(filterMap[parts[0]], parts[1])
		}

		wwiseClient, ok := ClientFromContext(cmd.Context())
		if !ok {
			return errors.New("could not get Wwise client from context")
		}

		fmt.Printf("Fetching Wwise SDK %s manifest...\n", sdkVersion)

		sdk := product.NewWwiseProduct(wwiseClient, "wwise")
		sdkProductVersion, err := sdk.GetVersion(sdkVersion)
		if err != nil {
			return fmt.Errorf("could not get SDK version: %v", err)
		}

		sdkVersionInfo, err := sdkProductVersion.GetInfo()
		if err != nil {
			return fmt.Errorf("could not get SDK version info: %v", err)
		}

		groupFilter := make([]product.GroupFilter, 0)
		for key, values := range filterMap {
			groupFilter = append(groupFilter, product.GroupFilter{GroupID: key, GroupValues: values})
		}

		files := sdkVersionInfo.FindFilesByGroups(groupFilter)
		if len(files) == 0 {
			fmt.Println("No packages found matching the specified filters.")
			return nil
		}

		var pendingFiles []product.File
		for _, file := range files {
			if sdkProductVersion.IsFileDownloaded(file.Name) {
				fmt.Printf("[Cached] %s is already downloaded\n", file.Name)
			} else {
				pendingFiles = append(pendingFiles, file)
			}
		}

		if len(pendingFiles) == 0 {
			fmt.Printf("All %d packages are already cached. Nothing to download.\n", len(files))
			return nil
		}

		var totalBytes int64
		for _, file := range pendingFiles {
			if file.Size > 0 {
				totalBytes += int64(file.Size)
			}
		}

		fmt.Printf("Found %d packages matching filter (%d to download, %s total).\n", len(files), len(pendingFiles), utils.FormatBytes(totalBytes))

		isTTY := utils.IsTTY()
		var bar *progressbar.ProgressBar
		if isTTY {
			bar = utils.NewDownloadProgressBar(
				cmd.OutOrStdout(),
				totalBytes,
				fmt.Sprintf("Downloading [1/%d] %s", len(pendingFiles), pendingFiles[0].Name),
			)
		}

		for i, file := range pendingFiles {
			if bar != nil {
				bar.Describe(fmt.Sprintf("Downloading [%d/%d] %s", i+1, len(pendingFiles), file.Name))
			} else {
				fmt.Printf("Downloading [%d/%d] %s (%s)...\n", i+1, len(pendingFiles), file.Name, utils.FormatBytes(int64(file.Size)))
			}

			err = sdkProductVersion.DownloadOrCache(file, bar)
			if err != nil {
				if bar != nil {
					bar.Exit()
				}
				return fmt.Errorf("could not download file %v: %w", file.Name, err)
			}

			if bar == nil {
				fmt.Printf("Completed [%d/%d] %s\n", i+1, len(pendingFiles), file.Name)
			}
		}

		if bar != nil {
			bar.Finish()
		}

		fmt.Printf("Successfully downloaded and extracted %d packages (%s total).\n", len(pendingFiles), utils.FormatBytes(totalBytes))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().String("sdk-version", "", "Wwise SDK version to download")
	downloadCmd.MarkFlagRequired("sdk-version")
	downloadCmd.Flags().StringArray("filter", []string{"Packages=SDK"}, "Filters to apply to the downloaded files")

	_ = viper.BindPFlag("sdk-version", downloadCmd.Flags().Lookup("sdk-version"))
	_ = viper.BindPFlag("filter", downloadCmd.Flags().Lookup("filter"))
}
