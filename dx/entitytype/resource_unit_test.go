package entitytype

import (
	"context"
	"testing"

	"terraform-provider-dx/dx/dxapi"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestModelToRequestBodySupportsSelectProperties verifies that select properties
// use the same option definition as multi_select properties, including defaults.
func TestModelToRequestBodySupportsSelectProperties(t *testing.T) {
	plan := EntityTypeModel{
		Identifier: types.StringValue("service"),
		Name:       types.StringValue("Service"),
		Properties: map[string]PropertyModel{
			"status": {
				Name: types.StringValue("Status"),
				Type: types.StringValue("select"),
				Options: []PropertyOptionModel{
					{Value: types.StringValue("active")},
					{Value: types.StringValue("inactive"), Color: types.StringValue("#ef4444")},
				},
			},
		},
	}

	payload := modelToRequestBody(context.Background(), plan, false)
	properties, ok := payload["properties"].([]map[string]interface{})
	if !ok {
		t.Fatalf("expected properties to be []map[string]interface{}, got %T", payload["properties"])
	}
	if len(properties) != 1 {
		t.Fatalf("expected one property, got %d", len(properties))
	}

	property := properties[0]
	if property["type"] != "select" {
		t.Errorf("expected property type select, got %v", property["type"])
	}

	definition, ok := property["definition"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected definition to be map[string]interface{}, got %T", property["definition"])
	}
	options, ok := definition["options"].([]map[string]interface{})
	if !ok {
		t.Fatalf("expected options to be []map[string]interface{}, got %T", definition["options"])
	}
	if len(options) != 2 {
		t.Fatalf("expected two options, got %d", len(options))
	}

	if options[0]["value"] != "active" {
		t.Errorf("expected first option value active, got %v", options[0]["value"])
	}
	if options[0]["color"] != DEFAULT_OPTION_COLOR {
		t.Errorf("expected default option color %q, got %v", DEFAULT_OPTION_COLOR, options[0]["color"])
	}
	if options[1]["value"] != "inactive" {
		t.Errorf("expected second option value inactive, got %v", options[1]["value"])
	}
	if options[1]["color"] != "#ef4444" {
		t.Errorf("expected explicit option color #ef4444, got %v", options[1]["color"])
	}
}

// TestResponseBodyToModelMapsSelectOptions verifies that select options returned
// by the API are restored to the Terraform property model.
func TestResponseBodyToModelMapsSelectOptions(t *testing.T) {
	apiResp := &dxapi.APIEntityTypeResponse{
		EntityType: dxapi.APIEntityType{
			Identifier: "service",
			Name:       "Service",
			Properties: []*dxapi.APIProperty{
				{
					Identifier: "status",
					Name:       "Status",
					Type:       "select",
					Definition: &dxapi.APIPropertyDefinition{
						Options: []dxapi.APIPropertyOption{
							{Value: "active", Color: "#3b82f6"},
							{Value: "inactive", Color: "#ef4444"},
						},
					},
				},
			},
		},
	}
	oldPlan := &EntityTypeModel{Properties: map[string]PropertyModel{}}
	state := &EntityTypeModel{}

	responseBodyToModel(context.Background(), apiResp, state, oldPlan)

	property, ok := state.Properties["status"]
	if !ok {
		t.Fatal("expected status property in state")
	}
	if property.Type.ValueString() != "select" {
		t.Errorf("expected property type select, got %q", property.Type.ValueString())
	}
	if len(property.Options) != 2 {
		t.Fatalf("expected two options, got %d", len(property.Options))
	}

	expected := []struct {
		value string
		color string
	}{
		{value: "active", color: "#3b82f6"},
		{value: "inactive", color: "#ef4444"},
	}
	for i, want := range expected {
		if got := property.Options[i].Value.ValueString(); got != want.value {
			t.Errorf("option %d: expected value %q, got %q", i, want.value, got)
		}
		if got := property.Options[i].Color.ValueString(); got != want.color {
			t.Errorf("option %d: expected color %q, got %q", i, want.color, got)
		}
	}
}
