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
	"errors"
	"os"
	"../utils"
	"github.com/spf13/cobra"
)

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "A brief description of your command",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Please provide a service name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		denvFile := utils.LoadDenvFile("")
		service, definitionError := utils.GetDefinition(os.Args[2], denvFile)
		if definitionError != nil {
			os.Exit(1)
		}

		utils.StopService(service)
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
}
