package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func execInPath(exec string, basePaths []string) (string, error) {
	for _, basePath := range basePaths {
		// Join the base path and the executable name
		fullPath := filepath.Join(basePath, exec)

		// Check if the file exists and is executable
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}
	return "", fmt.Errorf("not found")
}

func executioner(filepath string, args ...string) error {
	cmd := exec.Command(filepath, args...)
	err := cmd.Run()
	return err
}

func main() {
	for {
		set := map[string]bool{}

		builtins := []string{"echo", "type", "exit", "type"}
		for _, builtin := range builtins {
			set[builtin] = true
		}

		pathVariable := os.Getenv("PATH")
		paths := strings.Split(pathVariable, ":")

		fmt.Fprint(os.Stdout, "$ ")

		command, err := bufio.NewReader(os.Stdin).ReadString('\n')

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		command = command[:len(command)-1] // Remove the newline character
		fields := strings.Fields(command)

		if command == "exit 0" {
			os.Exit(0)

		} else if v, err := execInPath(fields[0], paths); err == nil {
			executioner(v, fields[1:]...)
		} else if fields[0] == "echo" {
			fmt.Println(strings.Join(strings.Fields(command)[1:], " "))

		} else if fields[0] == "type" {
			if set[fields[1]] {
				fmt.Println(fields[1], "is a shell builtin")
			} else if v, err := execInPath(fields[1], paths); err == nil {
				fmt.Println(fields[1], "is", v)

			} else {
				fmt.Println(fields[1] + ": not found")
			}
		} else {
			fmt.Println(command + ": command not found")
		}

	}

}
