package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/term"
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

// Changes the directory in cd command
func dirChanger(path string) (string, error) {
	path, err := tildaExpander(path)
	if err != nil {
		return "", err
	}
	newPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	} else {
		err := os.Chdir(newPath)
		return newPath, err
	}
}

// Expands Tilda in cd command
func tildaExpander(path string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if len(path) > 0 && path[0] == '~' {
		return filepath.Join(homeDir, path[1:]), nil
	} else {
		return path, nil
	}
}

func autoCompleter(fd int, builtins []string) string {
	input := ""
	for {
		b := make([]byte, 1)
		_, err := os.Stdin.Read(b)
		if err != nil {
			fmt.Println(err)
			return input
		}

		switch b[0] {

		case 9: // Tab key (ASCII 9)
			// Handle completion for built-in commands
			var completion string
			for _, cmd := range builtins {
				if strings.HasPrefix(cmd, input) {
					// If there's a match, complete the command
					completion = cmd
					break
				}
			}

			if completion != "" {
				// Print the completed command, including the space after it
				fmt.Print("\r$ " + completion + " ")
				input = completion
			}

		case 127: //Backspace
			if len(input) > 0 {
				fmt.Print("\b \b")
				input = input[:len(input)-1]
			}

		case 10: //Enter
			//Enter
			return input

		case 3: //Ctrl+C
			os.Exit(0)

		default:
			fmt.Print(string(b[0]))
			input += string(b[0])
		}

	}
}

func main() {
	set := map[string]bool{}

	builtins := []string{"echo", "type", "exit", "type", "pwd"}
	for _, builtin := range builtins {
		set[builtin] = true
	}

	pathVariable := os.Getenv("PATH")
	paths := strings.Split(pathVariable, ":")

	// Convert the terminal into RAW mode from COOKED MODE
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	// Defer the terminal to COOKED MODE to restore previous state
	defer term.Restore(fd, oldState)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error making terminal raw:", err)
		return
	}

	// REPL
	for {

		fmt.Fprint(os.Stdout, "$ ")

		command := autoCompleter(fd, builtins)

		term.Restore(fd, oldState)
		command = strings.TrimSpace(command)
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
			fullPath, err := dirChanger(path)
			if err != nil {
				fmt.Println("cd: " + fullPath + ": No such file or directory")
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
