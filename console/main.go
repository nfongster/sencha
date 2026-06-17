package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	client := NewClient("")

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
			runSession(scanner, client)
		default:
			fmt.Println("Unknown command. Available: start, quit")
		}
	}
}
