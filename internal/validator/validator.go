package validator

import (
	"fmt"
	"strings"

	"github.com/13rac1/go-wasm-ts-gen/internal/extractor"
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

// Validate runs all validation rules on extracted signatures, collecting all errors.
// Returns nil if validation passes, ValidationError with all issues if any fail.
func Validate(sigs []extractor.FunctionSignature) error {
	var errs []error

	for _, sig := range sigs {
		errs = append(errs, validateSignature(sig)...)
	}

	if len(errs) > 0 {
		return ValidationError{Errors: errs}
	}
	return nil
}

// validateSignature checks a single function signature for issues
func validateSignature(sig extractor.FunctionSignature) []error {
	var errs []error

	// Check for "any" return type (type inference failed)
	if sig.Returns.Type == "any" {
		errs = append(errs, fmt.Errorf("%s:%d: return type inferred as 'any' for %s\n"+
			"  hint: add result.Get(\"field\").String() or result.Bool() calls to infer type",
			sig.SourceFile, sig.Line, sig.Name))
	}

	return errs
}
