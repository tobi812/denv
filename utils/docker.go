package utils

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func StartService(service Definition) {
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

func StopService(service Definition) {
	args := extractArgsFromDenvFile(service)
	args = append(args, "down")

	execCommand("docker-compose", args...)
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

func extractArgsFromDenvFile(service Definition) []string {
	args := []string{}

	for _, file := range service.Files {
		file = GetEnvironmentPath() + "/" + file
		if !fileExists(file) {
			log.Fatalf("docker-compose %s file not found", file)
		}

		args = append(args, "-f")
		args = append(args, file)
	}

	return args
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

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}