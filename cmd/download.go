package cmd

import (
	"fmt"
	"strings"

	"github.com/mircearoata/wwise-cli/lib/wwise/product"
	"github.com/pkg/errors"
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

		fmt.Printf("Downloading Wwise sdk %s...\n", sdkVersion)

		sdk := product.NewWwiseProduct(wwiseClient, "wwise")
		sdkProductVersion, err := sdk.GetVersion(sdkVersion)
		if err != nil {
			return errors.Wrap(err, "could not get SDK version")
		}

		sdkVersionInfo, err := sdkProductVersion.GetInfo()
		if err != nil {
			return errors.Wrap(err, "could not get SDK version info")
		}

		groupFilter := make([]product.GroupFilter, 0)
		for key, values := range filterMap {
			groupFilter = append(groupFilter, product.GroupFilter{GroupID: key, GroupValues: values})
		}

		files := sdkVersionInfo.FindFilesByGroups(groupFilter)

		for _, file := range files {
			fmt.Printf("Downloading %v from %v\n", file.Name, file.URL)
			err = sdkProductVersion.DownloadOrCache(file)
			if err != nil {
				return errors.Wrapf(err, "could not download file %v", file.Name)
			}
		}

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
