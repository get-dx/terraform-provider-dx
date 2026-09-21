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
	propertyTypeSlackChannel     = "slack_channel"
	propertyTypeMSTeamsChannel   = "msteams_channel"
	propertyTypeEmail            = "email"
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
		propertyTypeSlackChannel,
		propertyTypeMSTeamsChannel,
		propertyTypeEmail,
		propertyTypeFileMatchingRule,
	}
	optionablePropertyTypes = []string{
		propertyTypeSelect,
		propertyTypeMultiSelect,
	}
)
