package dx

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Helper checks for and handles nil strings
func StringOrNull(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}

	return types.StringValue(*s)
}

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

// Helper preserves the value of a bool field if it's null in the plan
func BoolApiToTF(apiVal bool, planVal types.Bool) types.Bool {
	if planVal.IsNull() && !apiVal {
		return types.BoolNull()
	}
	return types.BoolValue(apiVal)
}

// Helper checks for and handles nil ints
func Float32OrNull(f *float32) types.Float32 {
	if f != nil {
		return types.Float32Value(*f)
	}
	return types.Float32Null()
}

func Int32OrNull(i *int32) types.Int32 {
	if i != nil {
		return types.Int32Value(*i)
	}
	return types.Int32Null()
}
