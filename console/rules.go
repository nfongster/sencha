package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func runRules(scanner *bufio.Scanner, client *Client) {
	fmt.Println("Rules subcommands: get <level_number>, set <level_number> <grammar.md> [exceptions.md]")
	fmt.Print("> ")
	if !scanner.Scan() {
		return
	}
	input := strings.TrimSpace(scanner.Text())
	parts := strings.Fields(input)
	if len(parts) < 2 {
		fmt.Println("Usage: get <level_number> | set <level_number> <grammar.md> [exceptions.md]")
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
			fmt.Println("Usage: set <level_number> <grammar.md> [exceptions.md]")
			return
		}
		rulesSet(client, number, parts[2], parts[3])
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
	fmt.Println("Exceptions:")
	if l.ExceptionsMD == "" {
		fmt.Println("(none)")
	} else {
		fmt.Println(l.ExceptionsMD)
	}
	fmt.Println()
}

func rulesSet(client *Client, number int, grammarPath, exceptionsPath string) {
	grammarData, err := os.ReadFile(grammarPath)
	if err != nil {
		fmt.Printf("Error reading grammar file %q: %v\n", grammarPath, err)
		return
	}
	grammarMD := string(grammarData)

	var exceptionsMD string
	if exceptionsPath != "" {
		data, err := os.ReadFile(exceptionsPath)
		if err != nil {
			fmt.Printf("Error reading exceptions file %q: %v\n", exceptionsPath, err)
			return
		}
		exceptionsMD = string(data)
	}

	if err := client.UpdateLevel(number, grammarMD, exceptionsMD); err != nil {
		fmt.Printf("Error updating level %d: %v\n", number, err)
		return
	}

	fmt.Printf("Grammar loaded from %s (%d bytes)\n", grammarPath, len(grammarData))
	if exceptionsPath != "" {
		fmt.Printf("Exceptions loaded from %s (%d bytes)\n", exceptionsPath, len(exceptionsMD))
	}
	fmt.Printf("Level %d rules updated successfully.\n", number)
}
