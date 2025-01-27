package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	// Uncomment this block to pass the first stage
	fmt.Fprint(os.Stdout, "$ ")

	// Wait for user input
	text, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	fmt.Fprint(os.Stdout, text+": Command not found")
}
