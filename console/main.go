package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("🍵 Sencha 🍵")
	fmt.Println("Available commands: Start, Quit")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "Start":
			startSession()
		case "Quit":
			fmt.Println("Quitting the app. Goodbye!")
			return
		default:
			fmt.Println("Invalid command. Please try again.")
		}
	}
}

func startSession() {
	fmt.Println("Starting study session...")
	// Logic to initiate a study session will go here
}
