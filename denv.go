package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/ini.v1"
	"gopkg.in/yaml.v2"
)

type Command struct {
	Container string
	Exec      string
}

type Definition struct {
	Name     string
	Files    []string
	Commands []Command
	path     string
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

func main() {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	denvPath := addCmd.String("path", "", "path to denv-file")

	if len(os.Args) == 1 {
		man, error := ioutil.ReadFile("docs/man.txt")
		if error != nil {
			log.Fatal(error)
		}

		fmt.Print(string(man))
		os.Exit(1)
	}

	switch os.Args[1] {
	case "up":
		if len(os.Args) < 3 {
			fmt.Println("Expected service name.")
			os.Exit(1)
		}

		denvFile := loadDenvFile("")
		service, definitionError := getDefinition(os.Args[2], denvFile)
		if definitionError != nil {
			os.Exit(1)
		}

		startService(service)

	case "down":
		if len(os.Args) < 3 {
			fmt.Println("Expected service name.")
			os.Exit(1)
		}

		denvFile := loadDenvFile("")
		service, definitionError := getDefinition(os.Args[2], denvFile)
		if definitionError != nil {
			os.Exit(1)
		}

		args := extractArgsFromDenvFile(service)
		args = append(args, "down")

		execCommand("docker-compose", args...)

	case "add":
		addCmd.Parse(os.Args[2:])

		denvFile := loadDenvFile(*denvPath)
		cfg := loadConfig()

		absolutePath := *denvPath
		if *denvPath == "" {
			path, err := os.Getwd()
			if err != nil {
				log.Println(err)
			}
			absolutePath = path
		}

		cfg.Section("environments").Key("denv." + denvFile.Environment.Name).SetValue(absolutePath)
		cfg.SaveTo("denv_config")

	case "boot":
		if len(os.Args) < 3 {
			fmt.Println("Expected service name.")
			os.Exit(1)
		}

		denvFile := loadDenvFile("")
		serviceList := getDefinitionList(os.Args[2], denvFile)

		for _, service := range serviceList {
			startService(service)
		}

	case "switch":
		if len(os.Args) < 3 {
			fmt.Println("Expected environment name.")
			os.Exit(1)
		}

		environment := os.Args[2]
		switchEnvironment(environment)

	default:
		fmt.Println("Unknown command.")
		os.Exit(1)
	}
}

func startService(service Definition) {
	args := extractArgsFromDenvFile(service)
	args = append(args, "up")
	args = append(args, "-d")

	execCommand("docker-compose", args...)

	for _, command := range service.Commands {
		containerArgs := []string{}
		containerArgs = append(containerArgs, "exec")
		containerArgs = append(containerArgs, command.Container)

		for _, execArg := range strings.Split(command.Exec, " ") {
			containerArgs = append(containerArgs, execArg)
		}

		execCommand("docker", containerArgs...)
	}
}

func execCommand(name string, args ...string) {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())

	if err != nil {
		fmt.Printf(errStr)
		fmt.Printf(stdout.String())
	} else {
		fmt.Printf(outStr)
		fmt.Printf(stderr.String())
	}
}

func getDefinition(name string, denvFile DenvFile) (Definition, error) {
	for _, definition := range denvFile.Environment.Definitions {
		if definition.Name == name {
			return definition, nil
		}
	}

	errorMessage := "Definition for service '" + name + "' not found!"
	fmt.Println(errorMessage)

	return Definition{}, errors.New(errorMessage)
}

func getDefinitionList(bootName string, denvFile DenvFile) []Definition {
	definitionNames := []string{}
	definitionList  := []Definition{}

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
		definition, definitionError := getDefinition(definitionName, denvFile)

		if definitionError != nil {
			os.Exit(1)
		}
		definitionList = append(definitionList, definition)
	}

	return definitionList
}

func loadConfig() *ini.File {
	cfg, err := ini.Load("denv_config")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	return cfg
}

func loadDenvFile(path string) DenvFile {

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

	denvFile := DenvFile{}
	error3 := yaml.Unmarshal([]byte(data), &denvFile)
	if error3 != nil {
		log.Fatalf("error: %v", error3)
	}

	return denvFile
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
	cfg.SaveTo("denv_config")
}

func extractArgsFromDenvFile(service Definition) []string {
	args := []string{}

	for _, file := range service.Files {
		args = append(args, "-f")
		args = append(args, file)
	}

	return args
}
