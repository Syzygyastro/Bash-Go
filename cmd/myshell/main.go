package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	for {
		set := map[string]bool{}

		builtins := []string{"echo", "type", "exit"}
		for _, builtin := range builtins {
			set[builtin] = true
		}

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
		} else if fields[0] == "echo" {
			fmt.Println(strings.Join(strings.Fields(command)[1:], " "))
		} else if fields[0] == "type" {
			if val, exists := set[fields[1]]; exists {
				fmt.Println(fields[1], "is a shell builtin")
			} else {
				fmt.Println(fields[1] + ": not found")
			}
		} else {
			fmt.Println(command + ": command not found")
		}

	}

}
