package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nshekhawat/lighttr/internal/request"
	"github.com/nshekhawat/lighttr/internal/tui"
)

// For testing
var osExit = os.Exit

func main() {
	// Command line flags
	method := flag.String("method", "", "HTTP method (GET, POST, PUT, DELETE, etc.)")
	url := flag.String("url", "", "Target URL")
	headers := flag.String("headers", "", "Headers in key:value,key2:value2 format")
	body := flag.String("body", "", "Request body")
	flag.Parse()

	// If command line arguments are provided, execute request directly
	if *url != "" {
		executeDirectRequest(*method, *url, *headers, *body)
		return
	}

	// Otherwise, launch the TUI
	model := tui.NewModel()
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		osExit(1)
	}
}

func executeDirectRequest(method, url, headers, body string) {
	req := request.NewRequestData()
	req.Method = method
	if req.Method == "" {
		req.Method = "GET"
	}
	req.URL = url
	req.Body = body

	// Parse headers
	if headers != "" {
		for _, header := range strings.Split(headers, ",") {
			parts := strings.SplitN(header, ":", 2)
			if len(parts) == 2 {
				req.Headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Validate request
	if err := req.Validate(); err != nil {
		fmt.Printf("Error: %v\n", err)
		osExit(1)
	}

	// Execute request
	resp, err := req.Execute()
	if err != nil {
		fmt.Printf("Error executing request: %v\n", err)
		osExit(1)
	}

	if resp.Error != "" {
		fmt.Printf("Error: %s\n", resp.Error)
		osExit(1)
	}

	// Print response
	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Time: %v\n", resp.ResponseTime)

	if len(resp.Headers) > 0 {
		fmt.Println("\nHeaders:")
		for k, v := range resp.Headers {
			fmt.Printf("%s: %s\n", k, v)
		}
	}

	if resp.Body != "" {
		fmt.Println("\nBody:")
		fmt.Println(resp.Body)
	}
}
