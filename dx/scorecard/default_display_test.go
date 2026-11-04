package scorecard

import (
	"context"
	"strings"
	"testing"

	"terraform-provider-dx/dx/dxapi"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func basePlan() ScorecardModel {
	return ScorecardModel{
		Name:                               types.StringValue("Test"),
		Type:                               types.StringValue("POINTS"),
		EntityFilterType:                   types.StringValue("entity_types"),
		DefaultDisplayEntityTypeIdentifier: types.StringNull(),
		DefaultDisplayGrouping:             types.StringNull(),
	}
}

func TestModelToRequestBodyDefaultDisplay(t *testing.T) {
	ctx := context.Background()
	keys := []string{"default_display_entity_type_identifier", "default_display_grouping"}

	t.Run("set values are sent", func(t *testing.T) {
		plan := basePlan()
		plan.DefaultDisplayEntityTypeIdentifier = types.StringValue("service")
		plan.DefaultDisplayGrouping = types.StringValue("owner_team")

		for _, setIds := range []bool{false, true} {
			payload, err := modelToRequestBody(ctx, plan, setIds)
			if err != nil {
				t.Fatal(err)
			}
			if payload[keys[0]] != "service" || payload[keys[1]] != "owner_team" {
				t.Errorf("setIds=%v: unexpected payload values: %v, %v", setIds, payload[keys[0]], payload[keys[1]])
			}
		}
	})

	t.Run("null is omitted on create", func(t *testing.T) {
		payload, err := modelToRequestBody(ctx, basePlan(), false)
		if err != nil {
			t.Fatal(err)
		}
		for _, k := range keys {
			if _, ok := payload[k]; ok {
				t.Errorf("expected %s to be omitted on create", k)
			}
		}
	})

	t.Run("null is sent as explicit null on update", func(t *testing.T) {
		payload, err := modelToRequestBody(ctx, basePlan(), true)
		if err != nil {
			t.Fatal(err)
		}
		for _, k := range keys {
			v, ok := payload[k]
			if !ok || v != nil {
				t.Errorf("expected %s to be an explicit nil on update, got %v (present=%v)", k, v, ok)
			}
		}
	})
}

func TestResponseBodyToModelDefaultDisplay(t *testing.T) {
	ctx := context.Background()
	strPtr := func(s string) *string { return &s }

	cases := []struct {
		name         string
		entityType   *string
		grouping     *string
		wantEntity   types.String
		wantGrouping types.String
	}{
		{"values", strPtr("service"), strPtr("owner_team"), types.StringValue("service"), types.StringValue("owner_team")},
		{"null", nil, nil, types.StringNull(), types.StringNull()},
		{"empty", strPtr(""), strPtr(""), types.StringNull(), types.StringNull()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			apiResp := &dxapi.APIResponse{Scorecard: dxapi.APIScorecard{
				Id:                                 "sc-1",
				Name:                               "Test",
				Type:                               "POINTS",
				EntityFilterType:                   "entity_types",
				DefaultDisplayEntityTypeIdentifier: tc.entityType,
				DefaultDisplayGrouping:             tc.grouping,
			}}
			state := basePlan()
			old := basePlan()
			responseBodyToModel(ctx, apiResp, &state, &old)

			if !state.DefaultDisplayEntityTypeIdentifier.Equal(tc.wantEntity) {
				t.Errorf("entity type: got %v, want %v", state.DefaultDisplayEntityTypeIdentifier, tc.wantEntity)
			}
			if !state.DefaultDisplayGrouping.Equal(tc.wantGrouping) {
				t.Errorf("grouping: got %v, want %v", state.DefaultDisplayGrouping, tc.wantGrouping)
			}
		})
	}
}

func TestValidateModelRelationGroupingRequiresEntityType(t *testing.T) {
	plan := basePlan()
	plan.DefaultDisplayGrouping = types.StringValue("relation:abc123")

	var diags diag.Diagnostics
	ValidateModel(plan, &diags)
	if !diags.HasError() {
		t.Fatal("expected an error for relation grouping without an entity type")
	}

	plan.DefaultDisplayEntityTypeIdentifier = types.StringValue("service")
	diags = diag.Diagnostics{}
	ValidateModel(plan, &diags)
	for _, d := range diags {
		if strings.Contains(d.Detail(), "default_display") {
			t.Errorf("unexpected default_display error: %s", d.Detail())
		}
	}
}
