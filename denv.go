package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

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

	type DenvFile struct {
		Version     string
		Environment struct {
			Name        string
			Definitions []string
			Boot        []string
		}
	}

	denvFile := DenvFile{}
	error1 := yaml.Unmarshal([]byte(data), &denvFile)
	if error1 != nil {
		log.Fatalf("error: %v", error1)
	}

	fmt.Printf("--- t:\n%v\n\n", denvFile)
	fmt.Println("Name: " + denvFile.Environment.Name)
}
