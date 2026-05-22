package domain

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {

	// Register validation function to ensure no duplicate body parts in complaints
	validate.RegisterValidation("unique_complaints", func(fl validator.FieldLevel) bool {
		field := fl.Field()

		// 1. Ensure the field being validated is a collection slice
		if field.Kind() != reflect.Slice && field.Kind() != reflect.Array {
			return false
		}

		seen := make(map[string]bool)

		// 2. Loop through every complaint object in the collection
		for i := 0; i < field.Len(); i++ {
			item := field.Index(i)

			// Handle structural pointers (e.g., []*ReferralComplaint) cleanly
			if item.Kind() == reflect.Pointer {
				if item.IsNil() {
					continue
				}
				item = item.Elem()
			}

			// 3. Dynamic Lookup: Extract ONLY the "BodyPart" property value
			bodyPartField := item.FieldByName("BodyPart")

			// Structural safeguard check
			if !bodyPartField.IsValid() {
				continue
			}

			// 4. Standarize the string token (e.g., "KNEE", "SHOULDER")
			bodyPart := strings.ToUpper(strings.TrimSpace(bodyPartField.String()))
			if bodyPart == "" {
				continue
			}

			// 5. Enforce Rule: If this specific structural part was already registered, fail
			if seen[bodyPart] {
				return false
			}
			seen[bodyPart] = true
		}

		return true
	})
}

func ValidateStruct(s any) error {
	return validate.Struct(s)
}
