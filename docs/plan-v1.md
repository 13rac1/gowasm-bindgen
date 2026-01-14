# Go WASM TypeScript Declaration Generator - Specification

## Problem Statement

Go WASM modules export functions to JavaScript via `syscall/js` and `js.FuncOf()`, but there's no automated way to generate TypeScript declarations. Developers must manually write `.d.ts` files, leading to:
- Type mismatches between Go signatures and TS declarations
- Maintenance burden when Go functions change
- No compile-time validation of WASM API usage

## Goal

Create a tool that generates TypeScript declaration files (`.d.ts`) from Go **unit tests**, providing type-safe WASM interop with **zero annotations required**.

**Works with**: Standard Go (1.21+) and TinyGo (0.33+)

---

## Design: Extract Everything from Unit Tests

### Core Principle

**Good code requires tests anyway** → Parse tests to extract types, names, and examples.

**No annotations needed.** Just write normal Go tests.

### How It Works

#### Step 1: Write Normal Unit Tests

```go
func TestMerkleVerifyProofHex(t *testing.T) {
    tests := []struct{
        name      string
        leafHash  string
        proof     []string
        root      string
        index     int
        wantValid bool
    }{
        {
            name:      "valid single-leaf proof",
            leafHash:  "abc123",
            proof:     []string{"def456"},
            root:      "root789",
            index:     0,
            wantValid: true,
        },
        {
            name:      "invalid proof",
            leafHash:  "wrong",
            proof:     []string{},
            root:      "root",
            index:     0,
            wantValid: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := merkleVerifyProofHex(js.Null(), []js.Value{
                js.ValueOf(tt.leafHash),
                js.ValueOf(tt.proof),
                js.ValueOf(tt.root),
                js.ValueOf(tt.index),
            })

            if errMsg := result.Get("error"); !errMsg.IsUndefined() {
                assert.False(t, tt.wantValid, "got error: %s", errMsg.String())
            } else {
                assert.Equal(t, tt.wantValid, result.Get("valid").Bool())
            }
        })
    }
}
```

#### Step 2: Tool Extracts Everything

**Type inference from `js.ValueOf()` calls**:
```go
js.ValueOf(tt.leafHash)  // leafHash is string  → param: string
js.ValueOf(tt.proof)     // proof is []string   → param: string[]
js.ValueOf(tt.root)      // root is string      → param: string
js.ValueOf(tt.index)     // index is int        → param: number
```

**Variable names from test table struct**:
```go
struct{
    leafHash  string   // ← param name: "leafHash"
    proof     []string // ← param name: "proof"
    root      string   // ← param name: "root"
    index     int      // ← param name: "index"
}
```

**Return type from assertions**:
```go
result.Get("error")      // → has {error: string}
result.Get("valid").Bool() // → has {valid: boolean}
// Union type: {valid: boolean} | {error: string}
```

**Examples from test cases**:
```go
// test case: "valid single-leaf proof"
// → Example: merkleVerifyProofHex("abc123", ["def456"], "root789", 0)
```

#### Step 3: Generated TypeScript

```typescript
/**
 * Auto-generated from test: TestMerkleVerifyProofHex
 *
 * @example
 * // valid single-leaf proof
 * merkleVerifyProofHex("abc123", ["def456"], "root789", 0)
 *
 * @example
 * // invalid proof
 * merkleVerifyProofHex("wrong", [], "root", 0)
 */
declare global {
  interface Window {
    merkleVerifyProofHex(
      leafHash: string,
      proof: string[],
      root: string,
      index: number
    ): {valid: boolean} | {error: string};
  }
}

export {};
```

---

## Type Mapping Rules

Tool automatically infers TypeScript types from Go types in `js.ValueOf()`:

| Go Type (in test) | `js.ValueOf()` | TypeScript Type |
|-------------------|----------------|-----------------|
| `string` | `js.ValueOf("hello")` | `string` |
| `int`, `uint`, `int32`, etc. | `js.ValueOf(123)` | `number` |
| `float32`, `float64` | `js.ValueOf(3.14)` | `number` |
| `bool` | `js.ValueOf(true)` | `boolean` |
| `[]string` | `js.ValueOf([]string{"a"})` | `string[]` |
| `[]int` | `js.ValueOf([]int{1, 2})` | `number[]` |
| `map[string]T` | `js.ValueOf(map[string]T{})` | `{[key: string]: T}` |
| `nil` | `js.ValueOf(nil)` | `null` |

**Return types** inferred from result usage:
```go
result.Get("fieldName").Bool()    // → {fieldName: boolean}
result.Get("fieldName").String()  // → {fieldName: string}
result.Get("fieldName").Int()     // → {fieldName: number}
result.IsUndefined()              // → undefined (optional field)
```

**Union types** from conditional checks:
```go
if !result.Get("error").IsUndefined() {
    // error case
} else {
    // success case with result.Get("valid")
}
// → {valid: boolean} | {error: string}
```

---

## Implementation

### Tool: `go-wasm-ts-gen`

**CLI tool written in Go** using `go/parser` and `go/ast`:

```bash
# Scan tests and generate TypeScript declarations
# Works with both standard Go and TinyGo WASM
go-wasm-ts-gen \
  --tests cmd/merkle-wasm/*_test.go \
  --output web/static/merkle.wasm.d.ts
```

### Architecture

```go
package main

import (
    "go/ast"
    "go/parser"
    "go/token"
)

type FunctionSignature struct {
    Name     string
    Params   []Parameter
    Returns  string
    Examples []string
    Doc      string
}

type Parameter struct {
    Name string
    Type string
}

func main() {
    // 1. Parse test files
    testFiles := parseGoFiles(os.Args[1:])

    // 2. Extract function signatures
    signatures := extractSignatures(testFiles)

    // 3. Generate TypeScript
    dts := generateTypeScript(signatures)

    // 4. Write output
    writeFile(outputPath, dts)
}

func extractSignatures(files []*ast.File) []FunctionSignature {
    var signatures []FunctionSignature

    for _, file := range files {
        ast.Inspect(file, func(n ast.Node) bool {
            // Find test functions
            fn, ok := n.(*ast.FuncDecl)
            if !ok || !strings.HasPrefix(fn.Name.Name, "Test") {
                return true
            }

            // Find calls to WASM exports
            for _, call := range findWASMCalls(fn.Body) {
                sig := FunctionSignature{
                    Name:   call.FuncName,
                    Params: extractParams(call),
                    Returns: inferReturnType(call, fn.Body),
                    Examples: extractExamples(call),
                }
                signatures = append(signatures, sig)
            }
            return true
        })
    }

    return signatures
}

func extractParams(call *CallExpr) []Parameter {
    // Look for table-driven test struct
    if tableStruct := findTableStruct(call); tableStruct != nil {
        return extractParamsFromStruct(tableStruct)
    }

    // Fallback: generic arg0, arg1, arg2
    return extractGenericParams(call)
}

func extractParamsFromStruct(structType *ast.StructType) []Parameter {
    var params []Parameter

    for _, field := range structType.Fields.List {
        // Get field name and type
        name := field.Names[0].Name
        tsType := goTypeToTS(field.Type)

        params = append(params, Parameter{
            Name: name,
            Type: tsType,
        })
    }

    return params
}

func goTypeToTS(expr ast.Expr) string {
    switch t := expr.(type) {
    case *ast.Ident:
        switch t.Name {
        case "string":
            return "string"
        case "int", "int32", "int64", "uint", "uint32", "float32", "float64":
            return "number"
        case "bool":
            return "boolean"
        }
    case *ast.ArrayType:
        elemType := goTypeToTS(t.Elt)
        return elemType + "[]"
    case *ast.MapType:
        keyType := goTypeToTS(t.Key)
        valType := goTypeToTS(t.Value)
        return fmt.Sprintf("{[key: %s]: %s}", keyType, valType)
    }
    return "any"
}

func inferReturnType(call *CallExpr, body *ast.BlockStmt) string {
    // Find result.Get() calls
    fields := make(map[string]string)

    ast.Inspect(body, func(n ast.Node) bool {
        // Look for: result.Get("fieldName").Bool()
        if sel, ok := n.(*ast.SelectorExpr); ok {
            if call, ok := sel.X.(*ast.CallExpr); ok {
                if isGetCall(call) {
                    fieldName := getFieldName(call)
                    fieldType := inferTypeFromMethod(sel.Sel.Name)
                    fields[fieldName] = fieldType
                }
            }
        }
        return true
    })

    // Build object type: {field1: type1, field2: type2}
    return buildObjectType(fields)
}
```

---

## Build Integration

### Option 1: Manual Build Script

```bash
#!/bin/bash
# scripts/build-wasm-with-types.sh

echo "Building WASM..."

# Option A: TinyGo (smaller binaries)
tinygo build -o web/static/merkle.wasm \
  -target wasm \
  -opt=z \
  cmd/merkle-wasm/main.go

# Option B: Standard Go
# GOOS=js GOARCH=wasm go build -o web/static/merkle.wasm cmd/merkle-wasm/main.go

echo "Generating TypeScript declarations from tests..."
go-wasm-ts-gen \
  --tests cmd/merkle-wasm/*_test.go \
  --output web/static/merkle.wasm.d.ts

echo "✓ Build complete"
echo "  WASM: web/static/merkle.wasm"
echo "  Types: web/static/merkle.wasm.d.ts"
```

### Option 2: Makefile Integration

```makefile
.PHONY: wasm wasm-types wasm-tinygo wasm-go

# TinyGo (smaller binaries, faster compile)
wasm-tinygo:
	tinygo build -o web/static/merkle.wasm \
		-target wasm -opt=z \
		cmd/merkle-wasm/main.go

# Standard Go (if TinyGo not available)
wasm-go:
	GOOS=js GOARCH=wasm go build \
		-o web/static/merkle.wasm \
		cmd/merkle-wasm/main.go

# Default to TinyGo
wasm: wasm-tinygo

wasm-types:
	go-wasm-ts-gen \
		--tests cmd/merkle-wasm/*_test.go \
		--output web/static/merkle.wasm.d.ts

build-web: wasm wasm-types
```

### Option 3: Vite Plugin (Future)

```typescript
// vite.config.ts
import { tinygoTypes } from 'vite-plugin-tinygo-types'

export default {
  plugins: [
    tinygoTypes({
      tests: '../cmd/merkle-wasm/*_test.go',
      output: 'static/merkle.wasm.d.ts'
    })
  ]
}
```

---

## Optional: Enhanced Documentation

### Minimal (No Annotations)

Tests are sufficient for basic type generation:

```go
func TestMerkleHashLeaf(t *testing.T) {
    result := merkleHashLeaf(js.Null(), []js.Value{
        js.ValueOf("48656c6c6f"),
    })
    assert.Equal(t, "06b3df...", result.String())
}
```

**Generates**:
```typescript
merkleHashLeaf(arg0: string): string
```

### Enhanced (Optional Annotations)

Add doc comments for richer documentation:

```go
// merkleHashLeaf hashes leaf data using Keccak256 with domain separation.
// The input must be hex-encoded (no 0x prefix).
func TestMerkleHashLeaf(t *testing.T) {
    // ... same test
}
```

**Generates**:
```typescript
/**
 * merkleHashLeaf hashes leaf data using Keccak256 with domain separation.
 * The input must be hex-encoded (no 0x prefix).
 */
merkleHashLeaf(arg0: string): string
```

### With Parameter Names (Table-Driven)

Use struct field names for semantic parameter names:

```go
func TestMerkleHashLeaf(t *testing.T) {
    tests := []struct{
        data string  // ← param name: "data"
        want string
    }{
        {data: "48656c6c6f", want: "06b3df..."},
    }

    for _, tt := range tests {
        result := merkleHashLeaf(js.Null(), []js.Value{
            js.ValueOf(tt.data),
        })
        assert.Equal(t, tt.want, result.String())
    }
}
```

**Generates**:
```typescript
merkleHashLeaf(data: string): string  // ← uses field name!
```

---

## Example: Complete Workflow

### 1. Write Go WASM Code

```go
// cmd/merkle-wasm/main.go
package main

import "syscall/js"

//go:wasmexport merkleHashLeafHex
func hashLeafHex(this js.Value, args []js.Value) interface{} {
    data := args[0].String()
    hash := merkle.HashLeaf([]byte(data))
    return hex.EncodeToString(hash)
}

func main() {
    <-make(chan struct{})
}
```

### 2. Write Unit Tests

```go
// cmd/merkle-wasm/main_test.go
package main

import (
    "syscall/js"
    "testing"
)

func TestMerkleHashLeafHex(t *testing.T) {
    tests := []struct{
        name string
        data string
        want string
    }{
        {
            name: "hello world",
            data: "48656c6c6f20576f726c64",
            want: "06b3dfaec148fb1bb2b066f10ec285e7c9bf402ab32aa78a5d38e34566810cd2",
        },
        {
            name: "empty string",
            data: "",
            want: "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := hashLeafHex(js.Null(), []js.Value{
                js.ValueOf(tt.data),
            })

            if got := result.String(); got != tt.want {
                t.Errorf("got %s, want %s", got, tt.want)
            }
        })
    }
}
```

### 3. Build WASM + Types

```bash
# Build WASM (choose one)

# Option A: TinyGo (recommended - smaller binaries)
tinygo build -o web/static/merkle.wasm -target wasm cmd/merkle-wasm/main.go

# Option B: Standard Go
# GOOS=js GOARCH=wasm go build -o web/static/merkle.wasm cmd/merkle-wasm/main.go

# Generate types from tests (works with both compilers)
go-wasm-ts-gen --tests cmd/merkle-wasm/main_test.go --output web/static/merkle.wasm.d.ts
```

### 4. Generated TypeScript

```typescript
/**
 * Auto-generated from test: TestMerkleHashLeafHex
 *
 * @example
 * // hello world
 * merkleHashLeafHex("48656c6c6f20576f726c64")
 * // → "06b3dfaec148fb1bb2b066f10ec285e7c9bf402ab32aa78a5d38e34566810cd2"
 *
 * @example
 * // empty string
 * merkleHashLeafHex("")
 * // → "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
 */
declare global {
  interface Window {
    merkleHashLeafHex(data: string): string;
  }
}

export {};
```

### 5. Use in TypeScript

```typescript
import type {} from '/merkle.wasm';

// TypeScript now knows the types!
const hash = window.merkleHashLeafHex("48656c6c6f");
//    ^? string (auto-complete works!)
```

---

## Implementation Estimate

**Go CLI tool**:
- Go AST parsing: ~300 LOC
- Type inference: ~400 LOC
- TypeScript generation: ~200 LOC
- **Total**: ~900 LOC

**Development time**:
- Day 1: AST parsing + basic type extraction
- Day 2: Table-driven test support + parameter names
- Day 3: Return type inference + union types
- Day 4: TypeScript generation + formatting
- Day 5: Testing + edge cases
- **Total**: ~1 week for production-ready tool

---

## Benefits Summary

### ✅ Zero Annotations Required
- Just write normal Go tests
- No special comments or directives
- No compiler modifications

### ✅ Types Are Validated
- If tests pass, types work
- No drift between types and implementation
- Compiler catches type mismatches in tests

### ✅ Examples Included
- Test cases become usage examples
- Shows real-world usage patterns
- Documentation is always up-to-date

### ✅ Parameter Names from Tests
- Table-driven tests provide semantic names
- No generic `arg0`, `arg1`
- Self-documenting code

### ✅ Minimal Overhead
- WASM binary: unchanged size
- Build step: <1s to parse tests
- No runtime cost

### ✅ Standard Go Tooling
- Uses `go/parser` and `go/ast`
- No external dependencies
- Works with any Go version

---

## Success Criteria

- ✅ Generates valid `.d.ts` files from test code
- ✅ Infers all primitive types correctly
- ✅ Handles arrays, maps, and complex objects
- ✅ Extracts parameter names from table-driven tests
- ✅ Infers union return types from error handling
- ✅ Includes test examples in JSDoc comments
- ✅ Works with standard Go testing conventions
- ✅ <1 second generation time for typical projects

---

## Comparison to Alternatives

| Approach | Annotations | Compiler Changes | Binary Size | Type Validation |
|----------|-------------|-----------------|-------------|-----------------|
| **Manual `.d.ts`** | None | None | No change | None |
| **Option 1: Source annotations** | Required | None | No change | External tool |
| **Option 2: Compiler extension** | Required | Modify TinyGo | +sidecar file | Compiler |
| **Option 3: Unit tests** | **None** ✨ | **None** ✨ | **No change** ✨ | **Tests validate** ✨ |

**Recommended**: Option 3 (this spec)

---

## Future Enhancements

1. **IDE Integration**: VSCode extension for inline type hints
2. **Runtime Validation**: Generate Zod schemas for runtime checking
3. **Bidirectional Types**: Generate Go test scaffolding from TypeScript types
4. **Watch Mode**: Auto-regenerate on test file changes
5. **Coverage Analysis**: Warn about untested WASM exports
6. **Multi-language**: Support Rust/C++ WASM with similar patterns

---

## Related Work

- **wasm-bindgen** (Rust): Uses proc macros + custom sections (requires annotations)
- **AssemblyScript**: TypeScript-first (native types, but different language)
- **Emscripten** (C++): Uses WebIDL (complex annotation format)

**Innovation**: First tool to extract WASM types purely from unit tests, with zero annotations.
