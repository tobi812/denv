
package cmd

import (
	"fmt"
	"log"
	"os"
	"../utils"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		add(path)
	},
}

func init() {
	addCmd.PersistentFlags().StringVarP(&path, "path", "p", getCurrentDir(), "define path")
	rootCmd.AddCommand(addCmd)
}

func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	return dir
}

func add(denvPath string)  {
	denvFile := utils.LoadDenvFile(denvPath)
	cfg := loadConfig()

	absolutePath := denvPath

	cfg.Section("environments").Key("denv." + denvFile.Environment.Name).SetValue(absolutePath)

	err := cfg.SaveTo(os.Getenv("HOME") + "/.denv_config")

	if err != nil {
		fmt.Printf("Failed to write denv file: %v", err)
		os.Exit(1)
	}
}