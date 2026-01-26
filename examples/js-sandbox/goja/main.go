//go:build js && wasm

package main

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/dop251/goja"
)

// JSResult contains the output from running JavaScript code in Goja.
type JSResult struct {
	Result   string // The final expression value
	Logs     string // Captured console.log output
	ErrorMsg string // Error message if execution failed
}

// RunJS executes JavaScript code in the Goja interpreter.
// The __goja__ global is injected to prove execution happens in Goja, not browser JS.
func RunJS(code string) JSResult {
	vm := goja.New()

	// Inject proof-of-goja globals
	vm.Set("__goja__", map[string]interface{}{
		"engine":    "goja",
		"goVersion": runtime.Version(),
		"goOS":      runtime.GOOS,
		"goArch":    runtime.GOARCH,
	})

	// Capture console output
	var logLines []string
	console := vm.NewObject()
	console.Set("log", func(call goja.FunctionCall) goja.Value {
		parts := make([]string, len(call.Arguments))
		for i, arg := range call.Arguments {
			parts[i] = arg.String()
		}
		logLines = append(logLines, strings.Join(parts, " "))
		return goja.Undefined()
	})
	vm.Set("console", console)

	// Run user code
	val, err := vm.RunString(code)
	if err != nil {
		return JSResult{
			Logs:     strings.Join(logLines, "\n"),
			ErrorMsg: err.Error(),
		}
	}

	return JSResult{
		Result: fmt.Sprintf("%v", val.Export()),
		Logs:   strings.Join(logLines, "\n"),
	}
}

func main() {
	select {} // Keep alive
}
