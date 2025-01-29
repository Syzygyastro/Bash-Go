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

func executioner(fileName string, args ...string) error {
	cmd := exec.Command(fileName, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	return err
}

// Function to check if a path exists
func pathExists(path string) error {
	// Use os.Stat to get the file info
	_, err := os.Stat(path)

	// If an error occurs, check if it's "not found" (os.ErrNotExist)
	if os.IsNotExist(err) {
		return err // Path does not exist
	}

	// If there's no error, the path exists
	return nil
}

func main() {
	for {
		set := map[string]bool{}

		builtins := []string{"echo", "type", "exit", "type", "pwd"}
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

		} else if fields[0] == "pwd" {
			mydir, err := os.Getwd()
			if err == nil {
				fmt.Println(mydir)
			}
		} else if fields[0] == "cd" {
			path := fields[1]
			err := pathExists(path)
			if err != nil {
				fmt.Println("cd: " + path + ": No such file or directory")
			} else {
				os.Chdir(path)
			}

		} else if _, err := execInPath(fields[0], paths); err == nil {
			err := executioner(fields[0], fields[1:]...)
			if err != nil {
				fmt.Println("Error:", err)
			}

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
