// Package validator checks generated Go code for syntax errors.
package validator

import (
	"fmt"
	"strings"

	"github.com/13rac1/gowasm-bindgen/internal/parser"
)

// ValidationError holds multiple validation errors
type ValidationError struct {
	Errors []error
}

func (e ValidationError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "found %d validation error(s):\n", len(e.Errors))
	for _, err := range e.Errors {
		fmt.Fprintf(&b, "  %v\n", err)
	}
	return b.String()
}

// ValidateFunctions runs all validation rules on parsed functions
func ValidateFunctions(parsed *parser.ParsedFile) error {
	var errs []error

	for _, fn := range parsed.Functions {
		errs = append(errs, validateFunction(fn, parsed.Types)...)
	}

	if len(errs) > 0 {
		return ValidationError{Errors: errs}
	}
	return nil
}

// validateFunction checks a single function for unsupported features
func validateFunction(fn parser.GoFunction, _ map[string]*parser.GoType) []error {
	var errs []error

	// Check parameters for unsupported types
	for _, param := range fn.Params {
		if err := validateType(param.Type, fn.Name, "parameter "+param.Name); err != nil {
			errs = append(errs, err)
		}
	}

	// Check return types for unsupported types
	for i, ret := range fn.Returns {
		if ret.IsError && i != len(fn.Returns)-1 {
			errs = append(errs, fmt.Errorf(
				"function %s: error return type must be last", fn.Name))
		}
		if !ret.IsError {
			if err := validateType(ret, fn.Name, "return type"); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errs
}

// validateType checks if a type is supported for WASM bindings
func validateType(t parser.GoType, funcName, context string) error {
	switch t.Kind {
	case parser.KindPrimitive:
		// All primitives are supported
		return nil

	case parser.KindSlice, parser.KindArray:
		if t.Elem != nil {
			return validateType(*t.Elem, funcName, context+" element")
		}
		return nil

	case parser.KindMap:
		// Only map[string]T is supported
		if t.Key == nil || t.Key.Name != "string" {
			return fmt.Errorf(
				"function %s: %s uses unsupported map type %s (only map[string]T is supported)",
				funcName, context, t.Name)
		}
		if t.Value != nil {
			return validateType(*t.Value, funcName, context+" map value")
		}
		return nil

	case parser.KindStruct:
		// Structs are supported, validate fields
		for _, field := range t.Fields {
			if err := validateType(field.Type, funcName, context+" field "+field.Name); err != nil {
				return err
			}
		}
		return nil

	case parser.KindPointer:
		// Pointers are supported, validate underlying type
		if t.Elem != nil {
			return validateType(*t.Elem, funcName, context+" (pointer)")
		}
		return nil

	case parser.KindError:
		// Error is supported
		return nil

	case parser.KindFunction:
		// Callbacks are only supported as direct function parameters
		if !strings.HasPrefix(context, "parameter ") {
			return fmt.Errorf(
				"function %s: %s uses a function type (functions are only supported as callback parameters)",
				funcName, context)
		}

		// Reject nested callbacks (callback inside another callback)
		if strings.Contains(context, "callback param") {
			return fmt.Errorf(
				"function %s: %s is a nested callback (callbacks cannot take callbacks as parameters)",
				funcName, context)
		}

		// Reject callbacks with return values
		if !t.IsVoid {
			return fmt.Errorf(
				"function %s: %s has a return value (only void callbacks are supported)",
				funcName, context)
		}

		// Validate callback parameter types recursively
		for i, param := range t.CallbackParams {
			if err := validateType(param, funcName, fmt.Sprintf("%s callback param %d", context, i)); err != nil {
				return err
			}
		}
		return nil

	case parser.KindUnsupported:
		return fmt.Errorf(
			"function %s: %s uses unsupported type %q (channels, interfaces, and external types are not supported)",
			funcName, context, t.Name)

	default:
		return fmt.Errorf(
			"function %s: %s uses unknown type kind %v", funcName, context, t.Kind)
	}
}
