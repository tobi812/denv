/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"../utils"
	"github.com/spf13/cobra"
)

// bootCmd represents the boot command
var bootCmd = &cobra.Command{
	Use:   "boot",
	Short: "Boot set of services.",
	Run: func(cmd *cobra.Command, args []string) {
		boot(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(bootCmd)
}

func boot(cmd *cobra.Command, args []string)  {
	denvFile := utils.LoadDenvFile("")
	serviceList := denvFile.Environment.Definitions

	if len(args[0]) > 0 {
		serviceList = utils.GetDefinitionList(args[0], denvFile)
	}

	for _, service := range serviceList {
		utils.StartService(service)
	}
}
