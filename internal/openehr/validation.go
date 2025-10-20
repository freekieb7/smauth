package openehr

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/freekieb7/smauth/internal/util"
)

var (
	UUIDRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	// ISO/IEC 8824 compliant version
	ISOOIDRegex = regexp.MustCompile(`^(0|1)(\.(0|[1-9][0-9]*)){0,1}(\.(0|[1-9][0-9]*))*$|^2(\.(0|[1-9][0-9]*))(\.(0|[1-9][0-9]*))*$`)
	// RFC 1034 compliant regex
	InternetIDRegex = regexp.MustCompile(`^([A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?)(?:\.([A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?))*$`)
)

type ValidationError struct {
	Model          string
	Path           string
	Message        string
	Recommendation string
}

// Define the correct interface
type Validatable interface {
	Validate() []ValidationError
}

type Validator struct {
}

func NewValidator() Validator {
	return Validator{}
}

func (v *Validator) ValidateModel(model any) []ValidationError {
	if model == nil {
		return []ValidationError{{
			Model:   "unknown",
			Path:    "$",
			Message: "model cannot be nil",
		}}
	}

	return validate("$", reflect.ValueOf(model))
}

func validate(path string, v reflect.Value) []ValidationError {
	var errors []ValidationError

	if !v.IsValid() {
		return []ValidationError{{
			Model:   "unknown",
			Path:    path,
			Message: "invalid value",
		}}
	}

	// Handle pointers
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		return validate(path, v.Elem())
	}

	// Check if the value implements Validatable interface
	if v.CanInterface() {
		if optionalValue, ok := v.Interface().(util.OptionalValue); ok {
			if !optionalValue.IsSet() {
				return nil
			}
		}

		if validatable, ok := v.Interface().(Validatable); ok {
			validationErrors := validatable.Validate()
			// Update the path for each validation error
			for i := range validationErrors {
				validationErrors[i].Path = path + validationErrors[i].Path
			}
			errors = append(errors, validationErrors...)
		}
	}

	// Handle slices and arrays
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			arrayPath := fmt.Sprintf("%s[%d]", path, i)
			errors = append(errors, validate(arrayPath, elem)...)
		}
		return errors
	}

	// Handle maps
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			mapValue := v.MapIndex(key)
			mapPath := fmt.Sprintf("%s[%v]", path, key.Interface())
			errors = append(errors, validate(mapPath, mapValue)...)
		}
		return errors
	}

	// Handle structs
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldType := v.Type().Field(i)

			// Skip unexported fields
			if !fieldType.IsExported() {
				continue
			}

			// Get JSON field name from struct tag
			fieldPath := path
			jsonFieldName := getJSONFieldName(fieldType)
			if jsonFieldName != "" {
				fieldPath += fmt.Sprintf(".%s", jsonFieldName)
			}

			errors = append(errors, validate(fieldPath, field)...)
		}
		return errors
	}

	return errors
}

// Helper function to extract JSON field name from struct tag
func getJSONFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		// If no json tag, use the field name
		return ""
	}

	// Split by comma to handle options like "omitempty"
	parts := strings.Split(jsonTag, ",")
	fieldName := parts[0]

	// Handle special cases
	if fieldName == "-" {
		// Field is ignored in JSON
		return ""
	}

	return fieldName
}

func ValidateUID(uid string) error {
	if uid == "" {
		return errors.New("UID cannot be empty")
	}

	// Check for UUID format (8-4-4-4-12 hex digits)
	if ValidateUUID(uid) {
		return nil
	}

	// Check for ISO OID format (numbers separated by dots)
	if ValidateISOOID(uid) {
		return nil
	}

	// Check for Internet ID format (reverse domain name style)
	if ValidateInternetID(uid) {
		return nil
	}

	return fmt.Errorf("UID '%s' does not match any valid format (UUID, ISO OID, or Internet ID)", uid)
}

func ValidateUUID(uuid string) bool {
	return UUIDRegex.MatchString(uuid)
}

func ValidateISOOID(oid string) bool {
	return ISOOIDRegex.MatchString(oid)
}

func ValidateInternetID(internetID string) bool {
	// Check total length (RFC 1034: max 255 characters, but commonly 253)
	if len(internetID) == 0 || len(internetID) > 255 {
		return false
	}

	return InternetIDRegex.MatchString(internetID)
}
