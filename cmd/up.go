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
	"errors"
	"github.com/spf13/cobra"
	"os"
)

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Start all containers of a service.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Please provide a service name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		denvFile := utils.LoadDenvFile("")
		service, definitionError := utils.GetDefinition(args[0], denvFile)
		if definitionError != nil {
			os.Exit(1)
		}

		utils.StartService(service)
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
}
