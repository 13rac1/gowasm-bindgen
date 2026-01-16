# TypeScript Configuration Explained

Our `tsconfig.json` uses the strictest TypeScript settings to catch bugs at compile time. Here's what each option does and why you should use it.

## Core Strict Mode

```json
"strict": true
```

**What it does:** Enables ALL strict type checking options at once (noImplicitAny, strictNullChecks, strictFunctionTypes, etc.)

**Why:** Catches 90% of potential runtime errors at compile time.

**Example it catches:**
```typescript
// ❌ Error: Parameter 'name' implicitly has an 'any' type
function greet(name) {
  return `Hello ${name}`;
}

// ✅ Correct
function greet(name: string): string {
  return `Hello ${name}`;
}
```

---

## Array/Index Access Safety

```json
"noUncheckedIndexedAccess": true
```

**What it does:** Array access and object key lookups return `T | undefined` instead of just `T`.

**Why:** Arrays can be empty. Objects can be missing keys. This catches index-out-of-bounds errors.

**Example it catches:**
```typescript
const users = ["Alice", "Bob"];

// ❌ Type error: users[5] is string | undefined, not string
const userName: string = users[5];

// ✅ Correct: handle the undefined case
const userName = users[5] ?? "Unknown";

// ✅ Or check first
if (users[5] !== undefined) {
  const userName: string = users[5];  // Now TypeScript knows it's string
}
```

---

## Optional Property Exactness

```json
"exactOptionalPropertyTypes": true
```

**What it does:** Optional properties can be *missing* or their declared type, but NOT explicitly `undefined`.

**Why:** In JavaScript, `{ name: undefined }` and `{}` behave differently for `"name" in obj` checks.

**Example it catches:**
```typescript
interface User {
  name: string;
  age?: number;  // Can be missing or a number
}

// ❌ Error: can't explicitly set optional property to undefined
const user: User = { name: "Alice", age: undefined };

// ✅ Correct: omit it entirely
const user: User = { name: "Alice" };

// ✅ Or include it with a value
const user: User = { name: "Alice", age: 30 };
```

---

## All Code Paths Must Return

```json
"noImplicitReturns": true
```

**What it does:** Functions with a return type must return a value in ALL code paths.

**Why:** Catches logic errors where some branches forget to return.

**Example it catches:**
```typescript
// ❌ Error: not all code paths return a value
function getStatus(active: boolean): string {
  if (active) {
    return "active";
  }
  // Oops, forgot the else case!
}

// ✅ Correct
function getStatus(active: boolean): string {
  if (active) {
    return "active";
  }
  return "inactive";
}
```

---

## No Switch Fallthrough

```json
"noFallthroughCasesInSwitch": true
```

**What it does:** Prevents accidental fallthrough in switch statements (missing `break`).

**Why:** Switch fallthrough is almost always a bug.

**Example it catches:**
```typescript
// ❌ Error: fallthrough case in switch
switch (op) {
  case "add":
    result = a + b;  // Missing break! Falls into "sub"
  case "sub":
    result = a - b;
    break;
}

// ✅ Correct
switch (op) {
  case "add":
    result = a + b;
    break;  // Explicit break
  case "sub":
    result = a - b;
    break;
}

// ✅ Or use early return
switch (op) {
  case "add":
    return a + b;
  case "sub":
    return a - b;
}
```

---

## Dead Code Detection

```json
"noUnusedLocals": true,
"noUnusedParameters": true
```

**What it does:** Errors on variables/parameters that are declared but never used.

**Why:** Catches typos, refactoring mistakes, and dead code.

**Example it catches:**
```typescript
// ❌ Error: 'userName' is declared but never used
function greet(userName: string, age: number): string {
  return `Hello ${name}!`;  // Typo! Should be userName
}

// ✅ Correct
function greet(userName: string): string {
  return `Hello ${userName}!`;
}

// ✅ If you intentionally don't use a param, prefix with underscore
function handleEvent(_event: Event): void {
  console.log("Something happened");
}
```

---

## Error Handling Type Safety

```json
"useUnknownInCatchVariables": true
```

**What it does:** Catch variables are typed as `unknown` instead of `any`.

**Why:** In JavaScript, you can throw anything (strings, numbers, objects). Forcing type checks prevents accessing properties that don't exist.

**Example it catches:**
```typescript
try {
  riskyOperation();
} catch (err) {  // err is 'unknown', not 'any'
  // ❌ Error: 'err' is of type 'unknown', can't access .message
  console.log(err.message);

  // ✅ Correct: check the type first
  if (err instanceof Error) {
    console.log(err.message);
  } else {
    console.log(String(err));
  }
}
```

---

## Explicit Dynamic Property Access

```json
"noPropertyAccessFromIndexSignature": true
```

**What it does:** Forces bracket notation for properties from index signatures.

**Why:** Makes it visually clear when you're accessing dynamic vs. static properties.

**Example it catches:**
```typescript
interface Config {
  [key: string]: string;
}

const config: Config = { host: "localhost" };

// ❌ Error: Property 'host' comes from index signature, use bracket notation
const host = config.host;

// ✅ Correct: bracket notation makes dynamic access explicit
const host = config["host"];
```

---

## Browser vs Node.js Differences

### Browser (`example/web/tsconfig.json`)
```json
{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ES2020",
    "moduleResolution": "bundler",
    "types": []  // Don't include Node.js types
  }
}
```

### Node.js (`example/tsconfig.json`)
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "types": ["node"]  // Include Node.js types
  }
}
```

**Key differences:**
- **Browser:** Uses ES modules, bundler handles imports, no Node.js globals
- **Node.js:** Uses NodeNext for ESM/CommonJS interop, includes `fs`, `path`, etc.

---

## Quick Setup for New Projects

Start with maximum strictness:

```bash
npx tsc --init
```

Then add these flags to `compilerOptions`:

```json
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "exactOptionalPropertyTypes": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "useUnknownInCatchVariables": true,
    "noPropertyAccessFromIndexSignature": true
  }
}
```

These settings will feel strict at first, but they prevent entire categories of bugs from reaching production.
