package dx

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Converts a string pointer into a TF string value, or `StringNull` if the pointer is nil.
func StringOrNull(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}

	return types.StringValue(*s)
}

// Converts a string pointer into a TF string value, or `StringNull` if the pointer is nil.
// If the string is empty, it will return `StringNull` instead of `StringValue("")`.
func StringOrNullConvertEmpty(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}

	stringVal := *s
	if stringVal == "" {
		return types.StringNull()
	}

	return types.StringValue(stringVal)
}

// Converts a boolean value from the API into a TF boolean value.
// If the boolean value is false and the plan value is null, it will return `BoolNull`.
func BoolApiToTF(apiVal bool, planVal types.Bool) types.Bool {
	if planVal.IsNull() && !apiVal {
		return types.BoolNull()
	}
	return types.BoolValue(apiVal)
}

// Converts a `*float32` into a TF float value, or `Float32Null` if the pointer is nil.
func Float32OrNull(f *float32) types.Float32 {
	if f != nil {
		return types.Float32Value(*f)
	}
	return types.Float32Null()
}

// Converts a `*int32` into a TF int value, or `Int32Null` if the pointer is nil.
func Int32OrNull(i *int32) types.Int32 {
	if i != nil {
		return types.Int32Value(*i)
	}
	return types.Int32Null()
}

// Converts a `*int64` into a TF int value, or `Int64Null` if the pointer is nil.
func Int64OrNull(i *int64) types.Int64 {
	if i != nil {
		return types.Int64Value(*i)
	}
	return types.Int64Null()
}
