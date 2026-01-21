---
title: "Type Mapping"
weight: 11
---

# Type Mapping

How Go types map to TypeScript types.

## Primitives

| Go Type | TypeScript Type |
|---------|-----------------|
| `string` | `string` |
| `bool` | `boolean` |
| `int`, `int8`, `int16`, `int32`, `int64` | `number` |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | `number` |
| `float32`, `float64` | `number` |

## Typed Arrays

Numeric slices map to TypeScript typed arrays for efficient data transfer:

| Go Type | TypeScript Type | Performance |
|---------|-----------------|-------------|
| `[]byte`, `[]uint8` | `Uint8Array` | Bulk copy (~10-100x faster) |
| `[]int8` | `Int8Array` | Element iteration |
| `[]int16` | `Int16Array` | Element iteration |
| `[]uint16` | `Uint16Array` | Element iteration |
| `[]int32` | `Int32Array` | Element iteration |
| `[]uint32` | `Uint32Array` | Element iteration |
| `[]float32` | `Float32Array` | Element iteration |
| `[]float64` | `Float64Array` | Element iteration |

**Note**: Only `[]byte` uses efficient bulk copy via `js.CopyBytesToGo()` and `js.CopyBytesToJS()`. Other numeric types use element-by-element iteration.

## Collections

| Go Type | TypeScript Type |
|---------|-----------------|
| `[]T` | `T[]` |
| `map[string]T` | `{ [key: string]: T }` |

**Limitation**: Only `map[string]T` is supported. Maps with non-string keys are not supported.

## Structs

Go structs become TypeScript interfaces. Field names use JSON tags if present:

```go
type User struct {
    ID        int    `json:"id"`
    FirstName string `json:"firstName"`
    IsActive  bool   `json:"isActive"`
}
```

```typescript
interface User {
    id: number;
    firstName: string;
    isActive: boolean;
}
```

## Functions

### Return Types

| Go Return | TypeScript Return (Worker) | TypeScript Return (Sync) |
|-----------|---------------------------|-------------------------|
| `T` | `Promise<T>` | `T` |
| `(T, error)` | `Promise<T>` (throws on error) | `T` (throws on error) |
| `error` | `Promise<void>` (throws on error) | `void` (throws on error) |
| (none) | `Promise<void>` | `void` |

### Callbacks

Void callbacks (no return value) are supported:

| Go Callback | TypeScript Callback |
|-------------|---------------------|
| `func()` | `() => void` |
| `func(T)` | `(arg0: T) => void` |
| `func(T, U)` | `(arg0: T, arg1: U) => void` |

**Not supported**: Callbacks with return values like `func(T) bool`.

## Special Cases

### interface{}

Go `interface{}` becomes TypeScript `any`:

```go
func GetValue() interface{} { ... }
// → getValue(): Promise<any>
```

**Recommendation**: Use concrete types whenever possible.

### Pointers

Pointers are automatically dereferenced:

```go
func GetUser() *User { ... }
// → getUser(): Promise<User>
```

### Unsupported Types

The following Go types are not supported and will cause validation errors:

- Channels (`chan T`)
- Interfaces (except `error`)
- External package types (except standard library)
- Function types as return values
- Maps with non-string keys
