package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const serverURL = "http://localhost:8080"

type sessionResponse struct {
	SessionID       string `json:"session_id"`
	Direction       string `json:"direction"`
	TotalCards      int    `json:"total_cards"`
	CardsRemaining  int    `json:"cards_remaining"`
	SessionComplete bool   `json:"session_complete"`
}

type revealResponse struct {
	CardID int    `json:"card_id"`
	Front  string `json:"front"`
	Back   string `json:"back"`
}

type gradeResponse struct {
	CardsRemaining  int            `json:"cards_remaining"`
	SessionComplete bool           `json:"session_complete"`
	GradeCounts     map[string]int `json:"grade_counts"`
}

type apiError struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

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

func runSession(scanner *bufio.Scanner) {
	fmt.Print("Direction (korean-to-english/english-to-korean/mixed) [default: korean-to-english]: ")
	direction := "korean-to-english"
	if scanner.Scan() {
		d := strings.TrimSpace(scanner.Text())
		if d != "" {
			direction = d
		}
	}

	sess, err := createSession(direction)
	if err != nil {
		fmt.Printf("Error starting session: %v\n", err)
		return
	}

	fmt.Printf("\n=== Session started (%d cards) ===\n\n", sess.TotalCards)

	for {
		reveal, err := revealCard(sess.SessionID)
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

		result, err := submitGrade(sess.SessionID, gradeStr)
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
					runSession(scanner)
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

func createSession(direction string) (*sessionResponse, error) {
	body := bytes.NewBufferString(`{"direction":"` + direction + `"}`)
	resp, err := http.Post(serverURL+"/api/sessions", "application/json", body)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var apiErr apiError
		json.NewDecoder(resp.Body).Decode(&apiErr)
		return nil, fmt.Errorf("%s (%s)", apiErr.Error, apiErr.Code)
	}

	var sess sessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&sess); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	return &sess, nil
}

func revealCard(sessionID string) (*revealResponse, error) {
	resp, err := http.Post(serverURL+"/api/sessions/"+sessionID+"/reveal", "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		json.NewDecoder(resp.Body).Decode(&apiErr)
		return nil, fmt.Errorf("%s (%s)", apiErr.Error, apiErr.Code)
	}

	var reveal revealResponse
	if err := json.NewDecoder(resp.Body).Decode(&reveal); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	return &reveal, nil
}

func submitGrade(sessionID, grade string) (*gradeResponse, error) {
	body := bytes.NewBufferString(`{"grade":"` + grade + `"}`)
	resp, err := http.Post(serverURL+"/api/sessions/"+sessionID+"/grade", "application/json", body)
	if err != nil {
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		json.NewDecoder(resp.Body).Decode(&apiErr)
		return nil, fmt.Errorf("%s (%s)", apiErr.Error, apiErr.Code)
	}

	var result gradeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("invalid response: %w", err)
	}
	return &result, nil
}

func printSummary(counts map[string]int, total int) {
	pass := counts["pass"]
	hard := counts["hard"]
	fail := counts["fail"]
	fmt.Println("\n=== Session complete! ===")
	fmt.Printf("Pass: %d   Hard: %d   Fail: %d   Total: %d\n", pass, hard, fail, total)
}
