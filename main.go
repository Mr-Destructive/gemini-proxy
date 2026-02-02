package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	chatCmd := flag.NewFlagSet("chat", flag.ExitOnError)
	chatMessage := chatCmd.String("msg", "", "Message to send")
	timeout := flag.Duration("timeout", 60*time.Second, "Request timeout")
	retries := flag.Int("retries", 3, "Number of retries")

	if len(os.Args) < 2 {
		interactive(timeout, retries)
		return
	}

	switch os.Args[1] {
	case "chat":
		chatCmd.Parse(os.Args[2:])
		if *chatMessage == "" {
			fmt.Println("Usage: gemini-proxy chat -msg 'your message'")
			os.Exit(1)
		}
		api := NewGeminiAPIWithConfig(*timeout, *retries)
		response, err := api.Ask(*chatMessage)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(response)
	default:
		interactive(timeout, retries)
	}
}

func interactive(timeout *time.Duration, retries *int) {
	api := NewGeminiAPIWithConfig(*timeout, *retries)
	fmt.Println("=" + strings.Repeat("=", 58) + "=")
	fmt.Println("Gemini Proxy - Interactive Mode")
	fmt.Println("=" + strings.Repeat("=", 58) + "=")
	fmt.Println("Commands:")
	fmt.Println("  /clear - Clear conversation")
	fmt.Println("  /quit  - Exit")
	fmt.Println("=" + strings.Repeat("=", 58) + "=\n")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		if input == "" {
			continue
		}

		switch input {
		case "/quit":
			fmt.Println("Goodbye!")
			os.Exit(0)
		case "/clear":
			api.ClearConversation()
			fmt.Println("Cleared\n")
		default:
			response, err := api.Ask(input)
			if err != nil {
				fmt.Printf("Error: %v\n\n", err)
			} else {
				fmt.Printf("Gemini: %s\n\n", response)
			}
		}
	}
}
