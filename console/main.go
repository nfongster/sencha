package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("🍵 Sencha 🍵")
	fmt.Println("Available commands: start, quit")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())

		switch input {
		case "quit":
			fmt.Println("Goodbye!")
			return
		case "start":
			runSession(scanner)
		default:
			fmt.Println("Unknown command. Available: start, quit")
		}
	}
}
