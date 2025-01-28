package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	for {
		fmt.Fprint(os.Stdout, "$ ")
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}
		command = command[:len(command)-1] // Remove the newline character
		if command == "exit 0" {
			os.Exit(0)
		} else if strings.Split(command, None)[0] == "echo" {
			fmt.Println(strings.Join(strings.Split(command, None)[1:], None))
		} else {
			fmt.Println(command + ": command not found")
		}

	}

}
