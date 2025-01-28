package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func execInPath(exec string, basePaths []string) (string, error) {
	for _, basePath := range basePaths {
		if p, err := filepath.Rel(basePath, exec); err == nil {
			return basePath + p, err
		}
	}
	return "", errors.New("")
}

func main() {
	for {
		// set := map[string]bool{}

		// builtins := []string{"echo", "type", "exit"}
		// for _, builtin := range builtins {
		// 	set[builtin] = true
		// }

		pathVariable := os.Getenv("PATH")
		paths := strings.Split(pathVariable, ":")
		for i, path := range paths {
			paths[i] = strings.ReplaceAll(path, " ", "") // Clean up each path
		}
		fmt.Fprint(os.Stdout, "$ ")

		command, err := bufio.NewReader(os.Stdin).ReadString('\n')

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		command = command[:len(command)-1] // Remove the newline character
		fields := strings.Fields(command)
		fmt.Println(filepath.Rel(paths[0], fields[1]))
		// fmt.Println(paths)
		if command == "exit 0" {
			os.Exit(0)
		} else if fields[0] == "echo" {
			fmt.Println(strings.Join(strings.Fields(command)[1:], " "))
		} else if fields[0] == "type" {
			if v, err := execInPath(fields[1], paths); err != nil {
				fmt.Println(fields[1], "is ", v)
			} else {
				fmt.Println(fields[1] + ": not found")
			}
		} else {
			fmt.Println(command + ": command not found")
		}

	}

}
