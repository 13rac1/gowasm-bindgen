package parser

import (
	"fmt"
	"strings"
)

// GoTypeToTS converts a GoType to TypeScript type string
func GoTypeToTS(t GoType) string {
	switch t.Kind {
	case KindPrimitive:
		return primitiveToTS(t.Name)

	case KindSlice, KindArray:
		if t.Elem != nil && t.Elem.Kind == KindPrimitive {
			if tsType := goElemToTypedArray(t.Elem.Name); tsType != "" {
				return tsType
			}
		}
		if t.Elem != nil {
			return GoTypeToTS(*t.Elem) + "[]"
		}
		return "any[]"

	case KindMap:
		if t.Key != nil && t.Value != nil {
			keyType := GoTypeToTS(*t.Key)
			valueType := GoTypeToTS(*t.Value)
			if keyType == "string" {
				return fmt.Sprintf("{[key: string]: %s}", valueType)
			}
			return "Record<" + keyType + ", " + valueType + ">"
		}
		return "any"

	case KindStruct:
		// Generate inline interface
		if len(t.Fields) == 0 {
			return "any"
		}
		var b strings.Builder
		b.WriteString("{")
		for i, field := range t.Fields {
			if i > 0 {
				b.WriteString(", ")
			}
			fieldName := field.JSONTag
			if fieldName == "" {
				fieldName = field.Name
			}
			b.WriteString(fieldName)
			b.WriteString(": ")
			b.WriteString(GoTypeToTS(field.Type))
		}
		b.WriteString("}")
		return b.String()

	case KindPointer:
		if t.Elem != nil {
			return GoTypeToTS(*t.Elem)
		}
		return "any"

	case KindError:
		return "string"

	case KindFunction:
		// Generate TypeScript callback type: (arg0: T, arg1: U) => void
		var params []string
		for i, p := range t.CallbackParams {
			params = append(params, fmt.Sprintf("arg%d: %s", i, GoTypeToTS(p)))
		}
		return "(" + strings.Join(params, ", ") + ") => void"

	default:
		return "any"
	}
}

// primitiveToTS converts Go primitive type names to TypeScript
func primitiveToTS(name string) string {
	switch name {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64",
		"byte", "rune":
		return "number"
	case "bool":
		return "boolean"
	default:
		return "any"
	}
}

// goElemToTypedArray maps Go slice element types to TypeScript typed array names.
// Returns empty string if the element type doesn't map to a typed array.
func goElemToTypedArray(elemName string) string {
	switch elemName {
	case "byte", "uint8":
		return "Uint8Array"
	case "int8":
		return "Int8Array"
	case "int16":
		return "Int16Array"
	case "int32":
		return "Int32Array"
	case "uint16":
		return "Uint16Array"
	case "uint32":
		return "Uint32Array"
	case "float32":
		return "Float32Array"
	case "float64":
		return "Float64Array"
	}
	return ""
}

// isByteSlice returns true if the type is []byte or []uint8.
func isByteSlice(t GoType) bool {
	if t.Kind != KindSlice || t.Elem == nil {
		return false
	}
	return t.Elem.Kind == KindPrimitive && (t.Elem.Name == "byte" || t.Elem.Name == "uint8")
}

// GoTypeToJSExtraction generates JavaScript code to extract a value from js.Value
// argExpr is the expression representing the js.Value argument (e.g., "args[0]")
// workerMode determines whether to generate worker-compatible callback code
func GoTypeToJSExtraction(t GoType, argExpr string, workerMode bool) string {
	switch t.Kind {
	case KindPrimitive:
		return primitiveExtraction(t.Name, argExpr)

	case KindSlice, KindArray:
		return sliceExtraction(t, argExpr, workerMode)

	case KindMap:
		return mapExtraction(t, argExpr, workerMode)

	case KindStruct:
		return structExtraction(t, argExpr, workerMode)

	case KindPointer:
		if t.Elem != nil {
			return GoTypeToJSExtraction(*t.Elem, argExpr, workerMode)
		}
		return argExpr

	case KindFunction:
		if workerMode {
			return workerCallbackCode(t, argExpr)
		}
		return callbackWrapperCode(t, argExpr)

	default:
		return argExpr
	}
}

// primitiveExtraction generates extraction code for primitive types
func primitiveExtraction(typeName, argExpr string) string {
	switch typeName {
	case "string":
		return argExpr + ".String()"
	case "int":
		return argExpr + ".Int()"
	case "int64":
		return "int64(" + argExpr + ".Float())"
	case "int32":
		return "int32(" + argExpr + ".Int())"
	case "int16":
		return "int16(" + argExpr + ".Int())"
	case "int8":
		return "int8(" + argExpr + ".Int())"
	case "uint":
		return "uint(" + argExpr + ".Int())"
	case "uint64":
		return "uint64(" + argExpr + ".Float())"
	case "uint32":
		return "uint32(" + argExpr + ".Int())"
	case "uint16":
		return "uint16(" + argExpr + ".Int())"
	case "uint8":
		return "uint8(" + argExpr + ".Int())"
	case "float64":
		return argExpr + ".Float()"
	case "float32":
		return "float32(" + argExpr + ".Float())"
	case "bool":
		return argExpr + ".Bool()"
	default:
		return argExpr
	}
}

// sliceExtraction generates extraction code for slices
func sliceExtraction(t GoType, argExpr string, workerMode bool) string {
	if t.Elem == nil {
		return "nil"
	}

	// Use js.CopyBytesToGo for byte slices (efficient bulk copy)
	if isByteSlice(t) {
		return byteSliceExtraction(argExpr)
	}

	// Element-by-element extraction for other types
	elemType := t.Elem
	var b strings.Builder

	b.WriteString("func() []")
	b.WriteString(elemType.Name)
	b.WriteString(" {\n")
	b.WriteString("\t\tlength := ")
	b.WriteString(argExpr)
	b.WriteString(".Length()\n")
	b.WriteString("\t\tresult := make([]")
	b.WriteString(elemType.Name)
	b.WriteString(", length)\n")
	b.WriteString("\t\tfor i := 0; i < length; i++ {\n")
	b.WriteString("\t\t\tresult[i] = ")
	b.WriteString(GoTypeToJSExtraction(*elemType, argExpr+".Index(i)", workerMode))
	b.WriteString("\n\t\t}\n")
	b.WriteString("\t\treturn result\n")
	b.WriteString("\t}()")

	return b.String()
}

// byteSliceExtraction generates extraction code for byte slices using js.CopyBytesToGo.
// This is ~10-100x faster than element-by-element extraction for large arrays.
func byteSliceExtraction(argExpr string) string {
	return `func() []byte {
		length := ` + argExpr + `.Length()
		result := make([]byte, length)
		js.CopyBytesToGo(result, ` + argExpr + `)
		return result
	}()`
}

// mapExtraction generates extraction code for maps
func mapExtraction(t GoType, argExpr string, workerMode bool) string {
	if t.Key == nil || t.Value == nil {
		return "nil"
	}

	// For now, only support map[string]T
	if t.Key.Name != "string" {
		return "nil"
	}

	var b strings.Builder
	b.WriteString("func() map[string]")
	b.WriteString(t.Value.Name)
	b.WriteString(" {\n")
	b.WriteString("\t\tresult := make(map[string]")
	b.WriteString(t.Value.Name)
	b.WriteString(")\n")
	b.WriteString("\t\tkeys := js.Global().Get(\"Object\").Call(\"keys\", ")
	b.WriteString(argExpr)
	b.WriteString(")\n")
	b.WriteString("\t\tfor i := 0; i < keys.Length(); i++ {\n")
	b.WriteString("\t\t\tkey := keys.Index(i).String()\n")
	b.WriteString("\t\t\tresult[key] = ")
	b.WriteString(GoTypeToJSExtraction(*t.Value, argExpr+".Get(key)", workerMode))
	b.WriteString("\n\t\t}\n")
	b.WriteString("\t\treturn result\n")
	b.WriteString("\t}()")

	return b.String()
}

// structExtraction generates extraction code for structs
func structExtraction(t GoType, argExpr string, workerMode bool) string {
	var b strings.Builder

	b.WriteString("func() ")
	b.WriteString(t.Name)
	b.WriteString(" {\n")
	b.WriteString("\t\treturn ")
	b.WriteString(t.Name)
	b.WriteString("{\n")

	for _, field := range t.Fields {
		fieldKey := field.JSONTag
		if fieldKey == "" {
			fieldKey = field.Name
		}

		b.WriteString("\t\t\t")
		b.WriteString(field.Name)
		b.WriteString(": ")
		b.WriteString(GoTypeToJSExtraction(field.Type, argExpr+".Get(\""+fieldKey+"\")", workerMode))
		b.WriteString(",\n")
	}

	b.WriteString("\t\t}\n")
	b.WriteString("\t}()")

	return b.String()
}

// callbackWrapperCode generates sync-mode callback wrapper (direct JS function invocation).
// If the JavaScript callback throws an error, it panics in Go, which is caught
// by the WASM error boundary and returned to TypeScript as a rejected Promise.
func callbackWrapperCode(t GoType, argExpr string) string {
	var goParams []string
	var jsArgs []string

	for i, p := range t.CallbackParams {
		paramName := fmt.Sprintf("arg%d", i)
		goParams = append(goParams, fmt.Sprintf("%s %s", paramName, p.Name))
		// Convert Go value to JS value using existing helper
		jsArgs = append(jsArgs, GoTypeToJSReturn(p, paramName))
	}

	return "func(" + strings.Join(goParams, ", ") + ") { " +
		argExpr + ".Invoke(" + strings.Join(jsArgs, ", ") + ") }"
}

// workerCallbackCode generates worker-mode callback wrapper (postMessage-based invocation).
// The callback ID is passed as an integer, and arguments are marshaled to a JS array.
// Panics if invokeCallback is not defined in the global scope (set by worker.js).
// NOTE: Callbacks are only valid during the function's execution - they are unregistered
// when the Go function returns, so callbacks must not be invoked from goroutines.
func workerCallbackCode(t GoType, argExpr string) string {
	var params, pushes strings.Builder

	for i, p := range t.CallbackParams {
		if i > 0 {
			params.WriteString(", ")
		}
		fmt.Fprintf(&params, "arg%d %s", i, p.Name)
		fmt.Fprintf(&pushes, "\t\tcbArgs.Call(\"push\", %s)\n",
			GoTypeToJSReturn(p, fmt.Sprintf("arg%d", i)))
	}

	return fmt.Sprintf(`func(%s) {
		cbArgs := js.Global().Get("Array").New()
%s		js.Global().Call("invokeCallback", %s.Int(), cbArgs)
	}`, params.String(), pushes.String(), argExpr)
}

// GoTypeToJSReturn generates JavaScript return conversion code
// valueExpr is the Go expression to convert (e.g., "result")
func GoTypeToJSReturn(t GoType, valueExpr string) string {
	switch t.Kind {
	case KindPrimitive:
		return primitiveReturn(t.Name, valueExpr)

	case KindSlice, KindArray:
		return sliceReturn(t, valueExpr)

	case KindMap:
		return mapReturn(t, valueExpr)

	case KindStruct:
		return structReturn(t, valueExpr)

	case KindPointer:
		if t.Elem != nil {
			return GoTypeToJSReturn(*t.Elem, valueExpr)
		}
		return valueExpr

	case KindError:
		return valueExpr + ".Error()"

	default:
		return valueExpr
	}
}

// primitiveReturn generates return conversion for primitives
func primitiveReturn(typeName, valueExpr string) string {
	// Most primitives can be returned directly in Go WASM
	return valueExpr
}

// sliceReturn generates return conversion for slices
func sliceReturn(t GoType, valueExpr string) string {
	if t.Elem == nil {
		return "nil"
	}

	// Use js.CopyBytesToJS for byte slices (efficient bulk copy)
	if isByteSlice(t) {
		return byteSliceReturn(valueExpr)
	}

	// For typed array element types (int32, float64, etc.), create JS typed array
	if jsTypedArray := goElemToTypedArray(t.Elem.Name); jsTypedArray != "" {
		return typedArrayReturn(jsTypedArray, valueExpr)
	}

	// For other primitive element types (int, string, bool), return directly
	if t.Elem.Kind == KindPrimitive {
		return valueExpr
	}

	// For complex types, need to convert each element
	var b strings.Builder
	b.WriteString("func() []interface{} {\n")
	b.WriteString("\t\tresult := make([]interface{}, len(")
	b.WriteString(valueExpr)
	b.WriteString("))\n")
	b.WriteString("\t\tfor i, v := range ")
	b.WriteString(valueExpr)
	b.WriteString(" {\n")
	b.WriteString("\t\t\tresult[i] = ")
	b.WriteString(GoTypeToJSReturn(*t.Elem, "v"))
	b.WriteString("\n\t\t}\n")
	b.WriteString("\t\treturn result\n")
	b.WriteString("\t}()")

	return b.String()
}

// typedArrayReturn generates return code for typed arrays (Int32Array, Float64Array, etc.).
// Creates a JS typed array and copies elements one by one.
func typedArrayReturn(jsTypedArray, valueExpr string) string {
	return `func() js.Value {
		slice := ` + valueExpr + `
		arr := js.Global().Get("` + jsTypedArray + `").New(len(slice))
		for i, v := range slice {
			arr.SetIndex(i, v)
		}
		return arr
	}()`
}

// byteSliceReturn generates return code for byte slices using js.CopyBytesToJS.
// This is ~10-100x faster than returning a Go slice directly for large arrays.
func byteSliceReturn(valueExpr string) string {
	return `func() js.Value {
		arr := js.Global().Get("Uint8Array").New(len(` + valueExpr + `))
		js.CopyBytesToJS(arr, ` + valueExpr + `)
		return arr
	}()`
}

// mapReturn generates return conversion for maps
func mapReturn(t GoType, valueExpr string) string {
	return "map[string]interface{}(" + valueExpr + ")"
}

// structReturn generates return conversion for structs
func structReturn(t GoType, valueExpr string) string {
	var b strings.Builder

	b.WriteString("map[string]interface{}{\n")
	for _, field := range t.Fields {
		fieldKey := field.JSONTag
		if fieldKey == "" {
			// Use lowercase first letter for JSON key
			fieldKey = strings.ToLower(field.Name[:1]) + field.Name[1:]
		}

		b.WriteString("\t\t\"")
		b.WriteString(fieldKey)
		b.WriteString("\": ")
		b.WriteString(GoTypeToJSReturn(field.Type, valueExpr+"."+field.Name))
		b.WriteString(",\n")
	}
	b.WriteString("\t}")

	return b.String()
}
