package workspacer

import (
	"fmt"
	"os"

	"github.com/JamesTiberiusKirk/workspacer/config"
	"github.com/JamesTiberiusKirk/workspacer/workspacer"
	"github.com/spf13/cobra"
)

var (
	GlobalConfig = config.DefaultGlobalConfig
	configFile   = ""
	workspace    = ""
)

var rootCmd = &cobra.Command{
	Use:   "workspacer",
	Short: "Workspace manager",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Workspace " + workspace)

		workspaceConfig := GlobalConfig.Workspaces[workspace]

		workspacer.StartOrSwitchToSession(workspace, workspaceConfig, GlobalConfig.SessionPresets, args[0])
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error while executing your CLI '%s'", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is $HOME/workspacer/workspacer.json)")

	rootCmd.Flags().StringVar(&workspace, "workspace", "", "Workspace to be used (the map value)")
	rootCmd.MarkFlagRequired("workspace")
}

func initConfig() {
	if configFile != "" {
		c, err := config.LoadGlobalConfig(configFile)
		if err != nil {
			panic(err)
		}
		if c == nil {
			fmt.Printf("No config in file")
		}
		GlobalConfig = *c
	}

}
