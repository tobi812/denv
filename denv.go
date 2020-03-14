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
	file, error := os.Open("denv.yaml")
	if error != nil {
		log.Fatal(error)
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

	flag.Parse()

	switch os.Args[1] {
	case "up":
		if len(os.Args) < 3 {
			fmt.Println("Expected service name.")
			os.Exit(1)
		}

		service, definitionError := getDefinition(os.Args[2], denvFile)
		if definitionError != nil {
			os.Exit(1)
		}

		args := []string{"up"}

		for _, file := range service.Files {
			args = append(args, "-f")
			args = append(args, file)
		}

		execCommand("docker-compose", args...)
	default:
		fmt.Println("Expected subcommand.")
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
