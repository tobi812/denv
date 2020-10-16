package utils

import (
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"gopkg.in/gookit/color.v1"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Command struct {
	Container string
	Exec      string
}

type Definition struct {
	Name     string
	Files    []string
	Wait     int
	Commands []Command
	Path     string
}

type Boot struct {
	Name        string
	Definitions []string
}

type DenvFile struct {
	Version     string
	Environment struct {
		Name        string
		Definitions []Definition
		BootGroups  []Boot
	}
}

func GetDefinition(name string, denvFile DenvFile) (Definition, error) {
	for _, definition := range denvFile.Environment.Definitions {
		if definition.Name == name {
			return definition, nil
		}
	}

	errorMessage := "Definition for service '" + name + "' not found!"
	fmt.Println(errorMessage)

	return Definition{}, errors.New(errorMessage)
}

func GetDefinitionList(bootName string, denvFile DenvFile) []Definition {
	definitionNames := []string{}
	definitionList := []Definition{}

	for _, bootGroup := range denvFile.Environment.BootGroups {
		if bootGroup.Name == bootName {
			definitionNames = append(definitionNames, bootGroup.Definitions...)
		}
	}

	if len(definitionNames) < 1 {
		fmt.Println("Definition for bootGroup '" + bootName + "' not found!")
		os.Exit(1)
	}

	for _, definitionName := range definitionNames {
		definition, definitionError := GetDefinition(definitionName, denvFile)

		if definitionError != nil {
			os.Exit(1)
		}
		definitionList = append(definitionList, definition)
	}

	return definitionList
}

func GetEnvironmentPath() string {
	currentEnvironment := viper.GetString("current.environment")

	return viper.GetString("environments.denv." + currentEnvironment)
}

func LoadDenvFile(path string) DenvFile {
	if path == "" {
		path = GetEnvironmentPath()
	}

	if !strings.HasSuffix(path, "/") && path != "" {
		path = path + "/"
	}

	file, error1 := os.Open(path + "denv.yaml")
	if error1 != nil {
		log.Fatal(error1)
	}

	defer file.Close()

	data, error2 := ioutil.ReadAll(file)
	if error2 != nil {
		log.Fatal(error2)
	}

	data = []byte(os.Expand(string(data), GetenvStrict))

	denvFile := DenvFile{}
	error3 := yaml.Unmarshal(data, &denvFile)
	if error3 != nil {
		log.Fatalf("error: %v", error3)
	}

	green := color.FgGreen.Render
	fmt.Printf("Environment: [%s]\n", green(denvFile.Environment.Name))

	return denvFile
}

func GetenvStrict(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable $%s, that was defined in denv file, not set!", key)
	}

	return value
}