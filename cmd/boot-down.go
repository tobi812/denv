package cmd

import (
	"../utils"
	"github.com/spf13/cobra"
)

// bootDownCmd represents the boot-down command
var bootDownCmd = &cobra.Command{
	Use:   "boot-down",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		bootDown(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(bootDownCmd)
}

func bootDown(cmd *cobra.Command, args []string) {
	denvFile := utils.LoadDenvFile("")
	serviceList := denvFile.Environment.Definitions

	if len(args) > 2 {
		serviceList = utils.GetDefinitionList(args[0], denvFile)
	}

	for _, service := range serviceList {
		utils.StopService(service)
	}
}

