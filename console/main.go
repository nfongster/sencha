package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)


func sendCommand(command string) (string, error) {
	urlStr := "http://localhost:8080/"
	data := url.Values{}
	data.Set("input", command)

	req, err := http.NewRequest("POST", urlStr, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("🍵 Sencha 🍵")
	fmt.Println("Available commands: Start, Quit")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		response, err := sendCommand(input)
		if err != nil {
			fmt.Println("error sending command:", err)
		} else {
			fmt.Println(response)
		}
	}
}

func startSession() {
	fmt.Println("Starting study session...")
	// Logic to initiate a study session will go here
}
