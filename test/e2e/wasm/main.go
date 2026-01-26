//go:build js && wasm

package main

// Info represents information about something.
type Info struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
	Active  bool   `json:"active"`
}

// Greet returns a greeting message.
func Greet(name string) string {
	return "Hello, " + name + "!"
}

// Add returns the sum of two numbers.
func Add(a, b int) int {
	return a + b
}

// GetInfo returns information about the given name.
func GetInfo(name string) Info {
	return Info{
		Name:    name,
		Version: 1,
		Active:  true,
	}
}

func main() {
	// Keep the program running
	<-make(chan struct{})
}
