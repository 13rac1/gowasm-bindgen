# go-wasm-ts-gen Example

This example demonstrates how to use `go-wasm-ts-gen` to generate TypeScript declarations for a Go WASM module.

## What's Included

- **wasm/main.go** - Go WASM functions (greet, calculate, formatUser, parseNumbers, validateEmail)
- **wasm/main_test.go** - Table-driven tests that go-wasm-ts-gen parses to extract types
- **web/** - Browser demo to interact with the WASM functions
- **verify_test.ts** - Deno test to verify generated types work correctly

## Quick Start

```bash
# Build everything (WASM binary + TypeScript types)
make all

# This runs:
# 1. setup     - Copies wasm_exec.js from Go installation
# 2. build     - Compiles Go to WASM (example.wasm)
# 3. generate  - Runs go-wasm-ts-gen to create types.d.ts
# 4. verify    - Runs Deno tests to validate types
```

## Generated Output

After running `make generate`, you'll have `types.d.ts`:

```typescript
declare global {
  interface Window {
    greet(name: string): string;
    calculate(a: number, b: number, op: string): number;
    formatUser(name: string, age: number, active: boolean): {
      displayName: string;
      status: string;
    };
    parseNumbers(input: string): number[];
    validateEmail(email: string): {
      valid: boolean;
      error: string;
    };
  }
}
export {};
```

## Try the Web Demo

```bash
# Start a local server
make serve

# Open http://localhost:8080/web/ in your browser
```

## How It Works

1. **Write Go WASM functions** with the standard signature:
   ```go
   func greet(this js.Value, args []js.Value) interface{} {
       name := args[0].String()
       return "Hello, " + name + "!"
   }
   ```

2. **Write tests** that call your functions:
   ```go
   func TestGreet(t *testing.T) {
       tests := []struct {
           name string  // <- go-wasm-ts-gen uses this as the parameter name
           want string
       }{
           {name: "World", want: "Hello, World!"},
       }
       for _, tt := range tests {
           result := greet(js.Null(), []js.Value{
               js.ValueOf(tt.name),  // <- type inferred from js.ValueOf()
           })
           // ...
       }
   }
   ```

3. **Run go-wasm-ts-gen** to extract types from tests:
   ```bash
   go-wasm-ts-gen --tests "wasm/*_test.go" --output types.d.ts
   ```

## WASM Functions in This Example

| Function | Parameters | Return Type | Description |
|----------|------------|-------------|-------------|
| `greet` | `name: string` | `string` | Returns a greeting |
| `calculate` | `a: number, b: number, op: string` | `number` | Basic arithmetic |
| `formatUser` | `name: string, age: number, active: boolean` | `object` | Format user data |
| `sumNumbers` | `input: string` | `number` | Sum comma-separated numbers |
| `validateEmail` | `email: string` | `object` | Validate email address |

## Requirements

- Go 1.21+
- Deno (for type verification)
- Python 3 (for `make serve`)
