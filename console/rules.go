package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func runRules(scanner *bufio.Scanner, client *Client) {
	fmt.Println("Rules subcommands: get <level_number>, set <level_number> <grammar.md>")
	fmt.Print("> ")
	if !scanner.Scan() {
		return
	}
	input := strings.TrimSpace(scanner.Text())
	parts := strings.Fields(input)
	if len(parts) < 2 {
		fmt.Println("Usage: get <level_number> | set <level_number> <grammar.md>")
		return
	}

	cmd := parts[0]
	numberStr := parts[1]

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		fmt.Printf("Invalid level number: %s\n", numberStr)
		return
	}

	switch cmd {
	case "get":
		rulesGet(client, number)
	case "set":
		if len(parts) < 3 {
			fmt.Println("Usage: set <level_number> <grammar.md>")
			return
		}
		rulesSet(client, number, parts[2])
	default:
		fmt.Printf("Unknown rules subcommand: %s\n", cmd)
	}
}

func rulesGet(client *Client, number int) {
	l, err := client.GetLevel(number)
	if err != nil {
		fmt.Printf("Error fetching level %d: %v\n", number, err)
		return
	}

	fmt.Printf("\nLevel %d — Phase %d\n", l.Number, l.PhaseNumber)
	fmt.Println(strings.Repeat("─", 40))
	fmt.Println("Grammar:")
	fmt.Println(l.GrammarMD)
	fmt.Println()
}

func rulesSet(client *Client, number int, grammarPath string) {
	grammarData, err := os.ReadFile(grammarPath)
	if err != nil {
		fmt.Printf("Error reading grammar file %q: %v\n", grammarPath, err)
		return
	}
	grammarMD := string(grammarData)

	if err := client.UpdateLevel(number, grammarMD); err != nil {
		fmt.Printf("Error updating level %d: %v\n", number, err)
		return
	}

	fmt.Printf("Grammar loaded from %s (%d bytes)\n", grammarPath, len(grammarData))
	fmt.Printf("Level %d rules updated successfully.\n", number)
}
