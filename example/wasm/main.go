package main

import (
	"fmt"
	"strconv"
	"strings"
	"syscall/js"
)

// greet returns a greeting message
func greet(this js.Value, args []js.Value) interface{} {
	name := args[0].String()
	return js.ValueOf("Hello, " + name + "!")
}

// calculate performs basic arithmetic operations
func calculate(this js.Value, args []js.Value) interface{} {
	a := args[0].Int()
	b := args[1].Int()
	op := args[2].String()

	var result int
	switch op {
	case "add":
		result = a + b
	case "sub":
		result = a - b
	case "mul":
		result = a * b
	case "div":
		if b == 0 {
			result = 0
		} else {
			result = a / b
		}
	default:
		result = 0
	}
	return js.ValueOf(result)
}

// formatUser creates a formatted user object
func formatUser(this js.Value, args []js.Value) interface{} {
	name := args[0].String()
	age := args[1].Int()
	active := args[2].Bool()

	status := "inactive"
	if active {
		status = "active"
	}

	return js.ValueOf(map[string]interface{}{
		"displayName": fmt.Sprintf("%s (%d)", name, age),
		"status":      status,
	})
}

// sumNumbers parses comma-separated numbers and returns their sum
func sumNumbers(this js.Value, args []js.Value) interface{} {
	input := args[0].String()
	if input == "" {
		return js.ValueOf(0)
	}

	parts := strings.Split(input, ",")
	sum := 0

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if n, err := strconv.Atoi(p); err == nil {
			sum += n
		}
	}

	return js.ValueOf(sum)
}

// validateEmail checks if an email address is valid
func validateEmail(this js.Value, args []js.Value) interface{} {
	email := args[0].String()

	if !strings.Contains(email, "@") {
		return js.ValueOf(map[string]interface{}{
			"valid": false,
			"error": "missing @ symbol",
		})
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return js.ValueOf(map[string]interface{}{
			"valid": false,
			"error": "invalid email format",
		})
	}

	if !strings.Contains(parts[1], ".") {
		return js.ValueOf(map[string]interface{}{
			"valid": false,
			"error": "domain must contain a dot",
		})
	}

	return js.ValueOf(map[string]interface{}{
		"valid": true,
	})
}

func main() {
	js.Global().Set("greet", js.FuncOf(greet))
	js.Global().Set("calculate", js.FuncOf(calculate))
	js.Global().Set("formatUser", js.FuncOf(formatUser))
	js.Global().Set("sumNumbers", js.FuncOf(sumNumbers))
	js.Global().Set("validateEmail", js.FuncOf(validateEmail))

	// Keep the Go program running
	select {}
}
