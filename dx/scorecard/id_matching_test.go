package scorecard

import (
	"context"
	"testing"

	"terraform-provider-dx/dx/dxapi"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestResponseBodyToModelMatchesByID verifies that responseBodyToModel maps
// API response items to state keys by ID rather than by position. This ensures
// that regardless of the order the API returns checks, levels, or check groups,
// they are always assigned to the correct state key.
func TestResponseBodyToModelMatchesByID(t *testing.T) {
	ctx := context.Background()

	strPtr := func(s string) *string { return &s }
	int32Ptr := func(i int32) *int32 { return &i }

	// Old plan has checks under specific keys with known IDs
	oldPlan := &ScorecardModel{
		Type:             types.StringValue("LEVEL"),
		EntityFilterType: types.StringValue("entity_types"),
		Levels: map[string]LevelModel{
			"bronze": {
				Id:    types.StringValue("level-1"),
				Name:  types.StringValue("Bronze"),
				Color: types.StringValue("#FB923C"),
				Rank:  types.Int32Value(1),
			},
		},
		Checks: map[string]CheckModel{
			"check_alpha": {
				Id:                types.StringValue("id-aaa"),
				Name:              types.StringValue("Check Alpha"),
				ScorecardLevelKey: types.StringValue("bronze"),
				Ordering:          types.Int32Value(0),
			},
			"check_beta": {
				Id:                types.StringValue("id-bbb"),
				Name:              types.StringValue("Check Beta"),
				ScorecardLevelKey: types.StringValue("bronze"),
				Ordering:          types.Int32Value(0),
			},
			"check_gamma": {
				Id:                types.StringValue("id-ccc"),
				Name:              types.StringValue("Check Gamma"),
				ScorecardLevelKey: types.StringValue("bronze"),
				Ordering:          types.Int32Value(0),
			},
		},
	}

	// API returns checks in a DIFFERENT order than the old plan keys
	apiResp := &dxapi.APIResponse{
		Scorecard: dxapi.APIScorecard{
			Id:                  "scorecard-1",
			Name:                "Test Scorecard",
			Type:                "LEVEL",
			EntityFilterType:    "entity_types",
			EvaluationFrequency: 2,
			Levels: []*dxapi.APILevel{
				{Id: strPtr("level-1"), Name: strPtr("Bronze"), Color: strPtr("#FB923C"), Rank: int32Ptr(1)},
			},
			Checks: []*dxapi.APICheck{
				// Note: returned in reverse order compared to alphabetical key order
				{Id: strPtr("id-ccc"), Name: strPtr("Check Gamma"), Ordering: 0, OutputType: strPtr("string"), Sql: strPtr("SELECT 'PASS' as status -- gamma")},
				{Id: strPtr("id-aaa"), Name: strPtr("Check Alpha"), Ordering: 0, OutputType: strPtr("string"), Sql: strPtr("SELECT 'PASS' as status -- alpha")},
				{Id: strPtr("id-bbb"), Name: strPtr("Check Beta"), Ordering: 0, OutputType: strPtr("string"), Sql: strPtr("SELECT 'PASS' as status -- beta")},
			},
		},
	}

	state := &ScorecardModel{}
	responseBodyToModel(ctx, apiResp, state, oldPlan)

	// Verify each check landed under the correct key by checking the ID
	expectations := map[string]string{
		"check_alpha": "id-aaa",
		"check_beta":  "id-bbb",
		"check_gamma": "id-ccc",
	}

	for key, expectedID := range expectations {
		check, ok := state.Checks[key]
		if !ok {
			t.Errorf("expected key %q in state.Checks, not found", key)
			continue
		}
		if check.Id.ValueString() != expectedID {
			t.Errorf("key %q: expected ID %q, got %q (name: %s)", key, expectedID, check.Id.ValueString(), check.Name.ValueString())
		}
	}

	// Run 100 times to verify determinism
	for i := 0; i < 100; i++ {
		s := &ScorecardModel{}
		responseBodyToModel(ctx, apiResp, s, oldPlan)
		for key, expectedID := range expectations {
			if s.Checks[key].Id.ValueString() != expectedID {
				t.Fatalf("iteration %d: key %q got ID %q, expected %q", i, key, s.Checks[key].Id.ValueString(), expectedID)
			}
		}
	}
}
