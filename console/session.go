package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func runSession(scanner *bufio.Scanner, client *Client) {
	fmt.Print("Direction (korean-to-english/english-to-korean/mixed) [default: korean-to-english]: ")
	direction := "korean-to-english"
	if scanner.Scan() {
		d := strings.TrimSpace(scanner.Text())
		if d != "" {
			direction = d
		}
	}

	sess, err := client.CreateSession(direction)
	if err != nil {
		fmt.Printf("Error starting session: %v\n", err)
		return
	}

	fmt.Printf("\n=== Session started (%d cards) ===\n\n", sess.TotalCards)

	for {
		reveal, err := client.RevealCard(sess.SessionID)
		if err != nil {
			fmt.Printf("Error revealing card: %v\n", err)
			return
		}

		fmt.Println(reveal.Front)
		fmt.Println("(Press ENTER to reveal answer)")
		scanner.Scan()

		fmt.Println()
		fmt.Println("→", reveal.Back)
		fmt.Println("Grade: [p]ass / [h]ard / [f]ail")
		fmt.Print("> ")

		if !scanner.Scan() {
			return
		}
		grade := strings.TrimSpace(scanner.Text())

		var gradeStr string
		switch grade {
		case "p", "pass":
			gradeStr = "pass"
		case "h", "hard":
			gradeStr = "hard"
		case "f", "fail":
			gradeStr = "fail"
		default:
			fmt.Println("Invalid grade, skipping this card.")
			continue
		}

		result, err := client.SubmitGrade(sess.SessionID, gradeStr)
		if err != nil {
			fmt.Printf("Error submitting grade: %v\n", err)
			return
		}

		if result.SessionComplete {
			printSummary(result.GradeCounts, sess.TotalCards)
			fmt.Println("\n[s]tart a new session / [q]uit")
			fmt.Print("> ")
			if scanner.Scan() {
				choice := strings.TrimSpace(scanner.Text())
				switch choice {
				case "s", "start":
					runSession(scanner, client)
					return
				default:
					fmt.Println("Goodbye!")
					os.Exit(0)
				}
			}
			return
		}

		fmt.Printf("✅ %d cards remaining\n\n", result.CardsRemaining)
	}
}

func printSummary(counts map[string]int, total int) {
	pass := counts["pass"]
	hard := counts["hard"]
	fail := counts["fail"]
	fmt.Println("\n=== Session complete! ===")
	fmt.Printf("Pass: %d   Hard: %d   Fail: %d   Total: %d\n", pass, hard, fail, total)
}
