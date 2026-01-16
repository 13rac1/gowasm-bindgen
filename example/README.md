# gowasm-bindgen Example

Go WASM functions use an untyped signature (`[]js.Value` in, `interface{}` out)—TypeScript sees them as `any`. This tool extracts parameter names and types from your Go tests to generate a typed TypeScript class API.

## What's Included

- **wasm/main.go** - Go WASM functions (greet, calculate, formatUser, sumNumbers, validateEmail)
- **wasm/main_test.go** - Table-driven tests that gowasm-bindgen parses to extract types
- **web/** - TypeScript browser demo using the generated types (strict mode, zero `any`)
- **verify_test.ts** - TypeScript test to verify generated types work correctly
- **client.ts** - Generated TypeScript class with typed methods
- **worker.js** - Generated Web Worker that runs Go WASM in a separate thread

With `--sync` flag, additional files are generated:
- **client.ts** - Synchronous API (no worker.js, runs on main thread)

## Quick Start

```bash
# Build everything (WASM binary + TypeScript types)
make all

# This runs:
# 1. setup     - Copies wasm_exec.js from TinyGo installation
# 2. build     - Compiles Go to WASM with TinyGo (example.wasm)
# 3. generate  - Runs gowasm-bindgen to create client.ts + worker.js
# 4. verify    - Runs TypeScript tests to validate types
# 5. web       - Compiles TypeScript demo using the generated types
```

### Using Standard Go (Alternative)

Standard Go produces larger binaries (~2.4MB vs ~200KB) but has full language support. Use it if your code needs features TinyGo doesn't support (see [TinyGo Language Support](https://tinygo.org/docs/reference/lang-support/)):

```bash
make setup-go   # Copy wasm_exec.js from Go installation
make build-go   # Build with standard Go
make generate
make verify
```

### Worker Mode (Default - Non-blocking)

By default, gowasm-bindgen generates a Web Worker wrapper for non-blocking async calls:

```bash
make generate  # Generate client.ts + worker.js
```

Then use the class-based API:

```typescript
import { Main } from './client';

const wasm = await Main.init('./worker.js');

// Non-blocking! UI stays responsive
const greeting = await wasm.greet('World');
const result = await wasm.calculate(5, 3, 'add');

wasm.terminate();  // Clean up when done
```

### Sync Mode (Opt-in - Blocking)

Use `--sync` flag for synchronous calls that block the main thread:

```bash
make generate-sync  # Generate client.ts only (no worker.js)
```

```typescript
import { Main } from './client';

const wasm = await Main.init('./example.wasm');  // async load
const greeting = wasm.greet('World');  // sync call, no await
```

## Generated Output

After running `make generate`, you'll have `client.ts` and `worker.js`:

```typescript
// client.ts - Class-based API with typed methods
export interface FormatUserResult {
  displayName: string;
  status: string;
}
export interface ValidateEmailResult {
  valid: boolean;
  error: string;
}

export class Main {
  static async init(workerUrl: string): Promise<Main>;
  greet(name: string): Promise<string>;
  calculate(a: number, b: number, op: string): Promise<number>;
  formatUser(name: string, age: number, active: boolean): Promise<FormatUserResult>;
  sumNumbers(input: string): Promise<number>;
  validateEmail(email: string): Promise<ValidateEmailResult>;
  terminate(): void;
}
```

## Try the Web Demo

```bash
# Build and start local server
make serve

# Open http://localhost:8080 in your browser
```

The `make serve` command:
1. Builds everything (WASM, TypeScript, types)
2. Copies all artifacts to `dist/`
3. Starts a local server using `npx serve`

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
   gowasm-bindgen --tests "wasm/*_test.go" --output client.ts
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
- Node.js 18+ (for TypeScript tooling and local server)
