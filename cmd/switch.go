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
	"fmt"
	"gopkg.in/ini.v1"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// switchCmd represents the switch command
var switchCmd = &cobra.Command{
	Use:   "switch",
	Short: "A brief description of your command",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Please provide a environment name")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		environment := args[0]
		switchEnvironment(environment)
	},
}

func loadConfig() *ini.File {
	cfg, err := ini.Load(os.Getenv("HOME") + "/.denv_config")

	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	return cfg
}

func switchEnvironment(environment string) {
	cfg := loadConfig()

	currentEnvironment := cfg.Section("current").Key("environment").String()
	if environment == currentEnvironment {
		log.Fatalf("Given environment %v is already selected.", environment)
	}

	if !cfg.Section("environments").HasKey("denv." + environment) {
		log.Fatalf("Environment %v not configured", environment)
	}

	cfg.Section("current").Key("environment").SetValue(environment)
	err := cfg.SaveTo(os.Getenv("HOME") + "/.denv_config")
	if err != nil {
		fmt.Printf("Failed to write denv file: %v", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
