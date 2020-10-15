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
	"os"
	"../utils"
	"github.com/spf13/cobra"
)

// bootDownCmd represents the bootDown command
var bootDownCmd = &cobra.Command{
	Use:   "bootDown",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		denvFile := utils.LoadDenvFile("")
		serviceList := denvFile.Environment.Definitions

		if len(os.Args) > 2 {
			serviceList = utils.GetDefinitionList(os.Args[2], denvFile)
		}

		for _, service := range serviceList {
			utils.StopService(service)
		}
	},
}

func init() {
	rootCmd.AddCommand(bootDownCmd)
}

