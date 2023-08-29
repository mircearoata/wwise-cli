package cmd

import (
	"fmt"

	"github.com/mircearoata/wwise-cli/lib/wwise"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var integrateUECmd = &cobra.Command{
	Use:   "integrate-ue",
	Short: "Integrate a version of wwise to an Unreal Engine project",
	RunE: func(cmd *cobra.Command, args []string) error {
		integrationVersion := viper.GetString("integration-version")
		project := viper.GetString("project")

		wwiseClient, ok := ClientFromContext(cmd.Context())
		if !ok {
			return errors.New("could not get Wwise client from context")
		}

		fmt.Printf("Integrating Wwise %s to UE project...\n", integrationVersion)

		err := wwise.IntegrateWwiseUnreal(project, integrationVersion, wwiseClient)
		if err != nil {
			return errors.Wrap(err, "could not integrate Wwise")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(integrateUECmd)

	integrateUECmd.Flags().String("integration-version", "", "Wwise UE integration version to download")
	integrateUECmd.MarkFlagRequired("integration-version")
	integrateUECmd.Flags().String("project", "", "Unreal Engine project to integrate Wwise to")
	integrateUECmd.MarkFlagRequired("project")

	_ = viper.BindPFlag("integration-version", integrateUECmd.Flags().Lookup("integration-version"))
	_ = viper.BindPFlag("project", integrateUECmd.Flags().Lookup("project"))
}
