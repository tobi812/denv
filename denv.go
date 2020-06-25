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

type Definition struct {
	Name  string
	Files []string
	path  string
}

type DenvFile struct {
	Version     string
	Environment struct {
		Name        string
		Definitions []Definition
		Boot        []string
	}
}

func main() {
	upCmd := flag.NewFlagSet("up", flag.ExitOnError)

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
		upCmd.Parse(os.Args[2:])

		denvFile := loadDenvFile("")
		if len(os.Args) < 3 {
			fmt.Println("Expected service name.")
			os.Exit(1)
		}

		service, definitionError := getDefinition(os.Args[2], denvFile)
		if definitionError != nil {
			os.Exit(1)
		}

		args := []string{}

		for _, file := range service.Files {
			args = append(args, "-f")
			args = append(args, file)
		}
		args = append(args, "up")
		args = append(args, "-d")

		execCommand("docker-compose", args...)
	case "add":
		addCmd.Parse(os.Args[2:])

		denvFile := loadDenvFile(*denvPath)

		cfg, err := ini.Load("denv_config")
		if err != nil {
			fmt.Printf("Fail to read file: %v", err)
			os.Exit(1)
		}

		path, err := os.Getwd()
		if err != nil {
			log.Println(err)
		}

		cfg.Section("environments").Key("denv." + denvFile.Environment.Name).SetValue(path)
		cfg.SaveTo("denv_config")

	default:
		fmt.Println("Unknown command.")
		os.Exit(1)
	}
	// fmt.Printf("--- t:\n%v\n\n", denvFile)
	// fmt.Println("Name: " + denvFile.Environment.Name)
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
	} else {
		fmt.Printf(outStr)
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

func loadDenvFile(path string) DenvFile {

	if !strings.HasSuffix(path, "/") && path != "" {
		path = path + "/"
	}

	file, error := os.Open(path + "denv.yaml")
	if error != nil {
		log.Fatal(error)
		os.Exit(1)
	}

	defer file.Close()

	data, error := ioutil.ReadAll(file)
	if error != nil {
		log.Fatal(error)
	}

	denvFile := DenvFile{}
	error1 := yaml.Unmarshal([]byte(data), &denvFile)
	if error1 != nil {
		log.Fatalf("error: %v", error1)
	}

	return denvFile
}
