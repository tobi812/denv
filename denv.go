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
	"time"

	"gopkg.in/gookit/color.v1"
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

func main() {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	denvPath := addCmd.String("path", "", "path to denv-file")

	if len(os.Args) == 1 {
//		man, error := ioutil.ReadFile("docs/man.txt")
//		if error != nil {
//			log.Fatal(error)
//		}
		man := "Denv is a tool to manage and concert multiple docker-compose scripts.\n\nUsage:\n\n    denv <command> [arguments]\n\nCommands:\n    up          start a service\n    down        stop a service\n    boot        start a list of services\n    boot-down   stop a list of services\n    add         add new configuration file\n    switch      switch environment context"
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

		stopService(service)

	case "boot":
		denvFile := loadDenvFile("")
		serviceList := denvFile.Environment.Definitions

		if len(os.Args) > 2 {
			serviceList = getDefinitionList(os.Args[2], denvFile)
		}

		for _, service := range serviceList {
			startService(service)
		}

	case "boot-down":
		denvFile := loadDenvFile("")
		serviceList := denvFile.Environment.Definitions

		if len(os.Args) > 2 {
			serviceList = getDefinitionList(os.Args[2], denvFile)
		}

		for _, service := range serviceList {
			stopService(service)
		}

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
		cfg.SaveTo(os.Getenv("HOME") + "/.denv_config")

	case "switch":
		if len(os.Args) < 3 {
			fmt.Println("Expected environment name.")
			os.Exit(1)
		}

		environment := os.Args[2]
		switchEnvironment(environment)

	case "config":
		fmt.Println("Print current config")

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

	if service.Wait > 0 {
		fmt.Println("Waiting for ", service.Wait, " seconds for ", service.Name)
		time.Sleep(time.Duration(service.Wait) * time.Second)
	}

	execServiceCommands(service)
}

func stopService(service Definition) {
	args := extractArgsFromDenvFile(service)
	args = append(args, "down")

	execCommand("docker-compose", args...)
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
		definition, definitionError := getDefinition(definitionName, denvFile)

		if definitionError != nil {
			os.Exit(1)
		}
		definitionList = append(definitionList, definition)
	}

	return definitionList
}

func loadConfig() *ini.File {
	cfg, err := ini.Load(os.Getenv("HOME") + "/.denv_config")

	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	return cfg
}

func loadDenvFile(path string) DenvFile {
	if path == "" {
		config := loadConfig()
		currentEnvironment := config.Section("current").Key("environment").String()
		path = config.Section("Environments").Key("denv." + currentEnvironment).String()
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
	cfg.SaveTo(os.Getenv("HOME") + "/.denv_config")
}

func extractArgsFromDenvFile(service Definition) []string {
	args := []string{}

	for _, file := range service.Files {
		if !fileExists(file) {
			log.Fatalf("docker-compose %s file not found", file)
		}

		args = append(args, "-f")
		args = append(args, file)
	}

	return args
}

func execServiceCommands(service Definition) {
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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func GetenvStrict(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable $%s, that was defined in denv file, not set!", key)
	}

	return value
}
