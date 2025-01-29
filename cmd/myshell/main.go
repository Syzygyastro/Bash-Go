package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

func execInPath(execName string, basePaths []string) (string, error) {
	for _, basePath := range basePaths {
		fullPath := filepath.Join(basePath, execName)
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
	return cmd.Run()
}

// dirChanger changes the directory for a "cd" command.
func dirChanger(path string) (string, error) {
	expanded, err := tildaExpander(path)
	if err != nil {
		return "", err
	}
	newPath, err := filepath.Abs(expanded)
	if err != nil {
		return "", err
	}
	if err := os.Chdir(newPath); err != nil {
		return "", err
	}
	return newPath, nil
}

// tildaExpander expands '~' to the user's home directory.
func tildaExpander(path string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if len(path) > 0 && path[0] == '~' {
		return filepath.Join(homeDir, path[1:]), nil
	}
	return path, nil
}

// autoCompleter reads input in raw mode, handles tab completion for builtins,
// and returns the full command once Enter is pressed.
func autoCompleter(fd int, builtins []string) string {
	var input string
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
					completion = cmd
					break
				}
			}
			if completion != "" {
				// Overwrite current line, show completed command + space
				fmt.Print("\r$ " + completion + " ")
				input = completion
			}

		case 127: // Backspace
			if len(input) > 0 {
				// Erase one character on the screen
				fmt.Print("\b \b")
				input = input[:len(input)-1]
			}

		case 10: // Enter (\n)
			return input

		case 13: // Carriage return (\r)
			// Ignore carriage returns in raw mode
			// Some terminals send \r before \n
			continue

		case 3: // Ctrl+C
			os.Exit(0)

		default:
			fmt.Print(string(b[0]))
			input += string(b[0])
		}
	}
}

func main() {
	// Built-in commands you want to complete or handle specially
	builtins := []string{"echo", "type", "exit", "pwd"}

	// Keep track of which commands are builtin
	builtinSet := make(map[string]bool)
	for _, b := range builtins {
		builtinSet[b] = true
	}

	// Split $PATH so we can search executables
	pathVariable := os.Getenv("PATH")
	paths := strings.Split(pathVariable, ":")

	// Put terminal into raw mode
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error making terminal raw:", err)
		return
	}
	// Ensure we restore terminal on exit
	defer term.Restore(fd, oldState)

	//REPL:
	for {
		// Print prompt
		fmt.Fprint(os.Stdout, "$ ")

		// Read command using raw-mode input
		command := autoCompleter(fd, builtins)

		// Restore cooked mode before processing
		term.Restore(fd, oldState)
		// Print a newline so the prompt stands on its own line next time
		fmt.Println()

		// Trim & parse
		command = strings.TrimSpace(command)
		fields := strings.Fields(command)
		if len(fields) == 0 {
			// No input: go to the next prompt
			// Re-enter raw mode
			oldState, _ = term.MakeRaw(fd)
			continue
		}

		// Handle builtins and external commands
		switch fields[0] {
		case "exit":
			// (Optionally handle `exit <code>`)
			os.Exit(0)

		case "pwd":
			if wd, err := os.Getwd(); err == nil {
				fmt.Println(wd)
			}

		case "cd":
			// Handle "cd" with or without argument
			if len(fields) < 2 {
				// No arg => cd to home
				homeDir, e := os.UserHomeDir()
				if e != nil {
					fmt.Println("cd: error locating home directory:", e)
				} else if _, e := dirChanger(homeDir); e != nil {
					fmt.Println("cd:", e)
				}
			} else {
				fullPath, e := dirChanger(fields[1])
				if e != nil {
					fmt.Println("cd:", fullPath, ": No such file or directory")
				}
			}

		case "echo":
			// Print remaining fields
			fmt.Println(strings.Join(fields[1:], " "))

		case "type":
			if len(fields) < 2 {
				fmt.Println("type: usage: type <command>")
			} else {
				target := fields[1]
				if builtinSet[target] {
					fmt.Println(target, "is a shell builtin")
				} else if found, e := execInPath(target, paths); e == nil {
					fmt.Println(target, "is", found)
				} else {
					fmt.Println(target + ": not found")
				}
			}

		default:
			// Not a builtin, try external command
			if fullPath, e := execInPath(fields[0], paths); e == nil {
				err := executioner(fullPath, fields[1:]...)
				if err != nil {
					fmt.Println("Error:", err)
				}
			} else {
				fmt.Println(command + ": command not found")
			}
		}

		// Re-enter raw mode for next input loop
		oldState, _ = term.MakeRaw(fd)
	}
}
