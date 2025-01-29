package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

// ----------------- Executable Discovery -----------------

// gatherExecutables collects all executable file names in the user's PATH.
// We store only one absolute path per name (the first found in PATH).
func gatherExecutables(paths []string) map[string]string {
	execMap := make(map[string]string)

	for _, dir := range paths {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			// Skip directories
			if e.IsDir() {
				continue
			}
			name := e.Name()
			fullPath := filepath.Join(dir, name)
			// Already stored a path for this name, skip
			if _, exists := execMap[name]; exists {
				continue
			}
			// Check if actually executable (on Unix)
			fi, err := os.Stat(fullPath)
			if err == nil && fi.Mode().IsRegular() && (fi.Mode().Perm()&0111) != 0 {
				// Store the name -> absolute path
				execMap[name] = fullPath
			}
		}
	}
	return execMap
}

// execInPath looks up the absolute path for an executable 'execName' from our map.
func execInPath(execName string, execMap map[string]string) (string, error) {
	if full, ok := execMap[execName]; ok {
		return full, nil
	}
	return "", fmt.Errorf("not found")
}

// ----------------- Shell Helpers -----------------

func executioner(fileName string, args ...string) error {
	cmd := exec.Command(fileName, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func dirChanger(path string) (string, error) {
	expanded, err := tildaExpander(path)
	if err != nil {
		return path, err
	}
	newPath, err := filepath.Abs(expanded)
	if err != nil {
		return path, err
	}
	if err := os.Chdir(newPath); err != nil {
		return path, err
	}
	return newPath, nil
}

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

// ----------------- Auto-Completer -----------------

// autoCompleter attempts to complete either a builtin or an executable name
// when the user presses <TAB>. If there's exactly one match, we complete it;
// otherwise, we beep.
func autoCompleter(builtins []string, execMap map[string]string) string {
	var input string
	for {
		b := make([]byte, 1)
		_, err := os.Stdin.Read(b)
		if err != nil {
			fmt.Println(err)
			return input
		}

		switch b[0] {
		case 9: // Tab key
			input = tryCompletion(input, builtins, execMap)

		case 127: // Backspace
			if len(input) > 0 {
				fmt.Print("\b \b")
				input = input[:len(input)-1]
			}

		case 10: // Enter (\n)
			return input

		case 3: // Ctrl+C
			os.Exit(0)

		case 13: // Carriage return (\r) - ignore or treat like Enter
			// Some terminals send \r\n on Enter.
			// We can ignore \r if \n is coming next.
			continue

		default:
			fmt.Print(string(b[0]))
			input += string(b[0])
		}
	}
}

// tryCompletion tries to complete 'input' as either a builtin or an external command.
// If exactly one match is found, it completes; otherwise, it beeps.
func tryCompletion(input string, builtins []string, execMap map[string]string) string {
	// If there's already a space, user is typing arguments, so skip
	// Or you can do more advanced logic to complete file paths, etc.
	if strings.ContainsRune(input, ' ') {
		fmt.Print("\a") // beep
		return input
	}

	// Gather possible matches for builtins
	builtinMatches := []string{}
	for _, b := range builtins {
		if strings.HasPrefix(b, input) {
			builtinMatches = append(builtinMatches, b)
		}
	}

	// Gather possible matches for executables
	execMatches := []string{}
	for cmdName := range execMap {
		if strings.HasPrefix(cmdName, input) {
			execMatches = append(execMatches, cmdName)
		}
	}

	// Decide priority: builtins first, then external
	matches := append(builtinMatches, execMatches...)

	if len(matches) == 1 {
		// Exactly one match; complete it
		completion := matches[0]
		// Print on the current line + space
		fmt.Print("\r$ " + completion + " ")
		// Return "completion + space" so the user can continue typing arguments
		return completion + " "
	}

	// If multiple or none, just beep
	fmt.Print("\a")
	return input
}

// ----------------- Main REPL -----------------

func main() {
	// Builtin commands
	builtins := []string{"echo", "type", "exit", "pwd"}

	// Create a set for quick "is-builtin" checks
	builtinSet := make(map[string]bool)
	for _, b := range builtins {
		builtinSet[b] = true
	}

	// Build a map of name -> absolute path for all executables in PATH
	pathVariable := os.Getenv("PATH")
	pathDirs := strings.Split(pathVariable, string(os.PathListSeparator))
	execMap := gatherExecutables(pathDirs)

	// Put terminal into raw mode
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error making terminal raw:", err)
		return
	}
	defer term.Restore(fd, oldState)

	// REPL
	for {
		fmt.Fprint(os.Stdout, "$ ")

		// Read user input in raw-mode
		command := autoCompleter(builtins, execMap)

		// Restore cooked mode so we can do normal prints
		term.Restore(fd, oldState)
		fmt.Println()

		fields := strings.Fields(command)
		if len(fields) == 0 {
			// No input => just re-enter raw mode and continue
			oldState, _ = term.MakeRaw(fd)
			continue
		}

		switch fields[0] {
		case "exit":
			os.Exit(0)

		case "pwd":
			if wd, e := os.Getwd(); e == nil {
				fmt.Println(wd)
			}

		case "cd":
			if len(fields) > 1 {
				fullPath, e := dirChanger(fields[1])
				if e != nil {
					fmt.Println("cd:", fullPath, ": No such file or directory")
				}
			} else {
				// cd with no args => go home
				home, e := os.UserHomeDir()
				if e != nil {
					fmt.Println("Error getting home:", e)
				} else {
					if _, e := dirChanger(home); e != nil {
						fmt.Println("cd:", e)
					}
				}
			}

		case "echo":
			fmt.Println(strings.Join(fields[1:], " "))

		case "type":
			if len(fields) < 2 {
				fmt.Println("type: usage: type <command>")
			} else {
				target := fields[1]
				if builtinSet[target] {
					fmt.Println(target, "is a shell builtin")
				} else if fullPath, e := execInPath(target, execMap); e == nil {
					fmt.Println(target, "is", fullPath)
				} else {
					fmt.Println(target + ": not found")
				}
			}

		default:
			// External command?
			if fullPath, e := execInPath(fields[0], execMap); e == nil {
				if err := executioner(fullPath, fields[1:]...); err != nil {
					fmt.Println("Error:", err)
				}
			} else {
				fmt.Println(command + ": command not found")
			}
		}

		// Re-enter raw mode for the next prompt
		oldState, _ = term.MakeRaw(fd)
	}
}
