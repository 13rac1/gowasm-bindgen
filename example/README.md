# gowasm-bindgen Example

Go WASM modules expose functions on `window` with no type information—TypeScript sees them as `any`. This tool extracts parameter names and types from your Go tests to generate `.d.ts` declarations, giving you type-safe WASM calls in TypeScript.

## What's Included

- **wasm/main.go** - Go WASM functions (greet, calculate, formatUser, sumNumbers, validateEmail)
- **wasm/main_test.go** - Table-driven tests that gowasm-bindgen parses to extract types
- **web/** - TypeScript browser demo using the generated types
- **verify_test.ts** - Deno test to verify generated types work correctly

## Quick Start

```bash
# Build everything (WASM binary + TypeScript types)
make all

# This runs:
# 1. setup     - Copies wasm_exec.js from TinyGo installation
# 2. build     - Compiles Go to WASM with TinyGo (example.wasm)
# 3. generate  - Runs gowasm-bindgen to create types.d.ts
# 4. verify    - Runs Deno tests to validate types
# 5. web       - Compiles TypeScript demo using the generated types
```

### Using Standard Go (Alternative)

Standard Go also works but produces a much larger binary (~2.4MB vs ~200KB with TinyGo):

```bash
make setup-go   # Copy wasm_exec.js from Go installation
make build-go   # Build with standard Go
make generate
make verify
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
    sumNumbers(input: string): number;
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
           name string  // <- gowasm-bindgen uses this as the parameter name
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

3. **Run gowasm-bindgen** to extract types from tests:
   ```bash
   gowasm-bindgen --tests "wasm/*_test.go" --output types.d.ts
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

- [TinyGo](https://tinygo.org/getting-started/install/) (recommended for small WASM binaries)
- Go 1.21+ (alternative, produces larger binaries)
- Deno (for type verification)
- Python 3 (for `make serve`)
