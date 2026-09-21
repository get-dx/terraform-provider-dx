package entitytype

const (
	propertyTypeText             = "text"
	propertyTypeUser             = "user"
	propertyTypeBoolean          = "boolean"
	propertyTypeMultiSelect      = "multi_select"
	propertyTypeNumber           = "number"
	propertyTypeDate             = "date"
	propertyTypeJSON             = "json"
	propertyTypeList             = "list"
	propertyTypeOpenAPI          = "openapi"
	propertyTypeSelect           = "select"
	propertyTypeComputed         = "computed"
	propertyTypeFileMatchingRule = "file_matching_rule"
	propertyTypeURL              = "url"
)

var (
	propertyTypes = []string{
		propertyTypeText,
		propertyTypeUser,
		propertyTypeURL,
		propertyTypeSelect,
		propertyTypeMultiSelect,
		propertyTypeBoolean,
		propertyTypeNumber,
		propertyTypeComputed,
		propertyTypeDate,
		propertyTypeJSON,
		propertyTypeList,
		propertyTypeOpenAPI,
		propertyTypeFileMatchingRule,
	}
	optionablePropertyTypes = []string{
		propertyTypeSelect,
		propertyTypeMultiSelect,
	}
)
