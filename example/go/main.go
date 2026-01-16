package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type User struct {
	DisplayName string `json:"displayName"`
	Status      string `json:"status"`
}

type EmailResult struct {
	Valid bool   `json:"valid"`
	Error string `json:"error"`
}

// Greet returns a greeting message
func Greet(name string) string {
	return "Hello, " + name + "!"
}

// Calculate performs basic arithmetic operations
func Calculate(a int, b int, op string) int {
	switch op {
	case "add":
		return a + b
	case "sub":
		return a - b
	case "mul":
		return a * b
	case "div":
		if b == 0 {
			return 0
		}
		return a / b
	default:
		return 0
	}
}

// FormatUser creates a formatted user object
func FormatUser(name string, age int, active bool) User {
	status := "inactive"
	if active {
		status = "active"
	}
	return User{
		DisplayName: fmt.Sprintf("%s (%d)", name, age),
		Status:      status,
	}
}

// SumNumbers parses comma-separated numbers and returns their sum
func SumNumbers(input string) int {
	if input == "" {
		return 0
	}

	parts := strings.Split(input, ",")
	sum := 0

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if n, err := strconv.Atoi(p); err == nil {
			sum += n
		}
	}

	return sum
}

// ValidateEmail checks if an email address is valid
func ValidateEmail(email string) EmailResult {
	if !strings.Contains(email, "@") {
		return EmailResult{
			Valid: false,
			Error: "missing @ symbol",
		}
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return EmailResult{
			Valid: false,
			Error: "invalid email format",
		}
	}

	if !strings.Contains(parts[1], ".") {
		return EmailResult{
			Valid: false,
			Error: "domain must contain a dot",
		}
	}

	return EmailResult{
		Valid: true,
		Error: "",
	}
}

// Divide performs integer division with error handling
func Divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("division by zero")
	}
	return a / b, nil
}

func main() {
	// Keep the Go program running
	// bindings_gen.go will handle all WASM function registrations
	select {}
}
