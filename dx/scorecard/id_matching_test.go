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

// TestResponseBodyToModelPreservesGroupingKeysOnCreate verifies that on Create,
// when oldPlan checks have no IDs yet, scorecard_level_key and
// scorecard_check_group_key are still preserved from the plan.
func TestResponseBodyToModelPreservesGroupingKeysOnCreate(t *testing.T) {
	ctx := context.Background()

	strPtr := func(s string) *string { return &s }
	int32Ptr := func(i int32) *int32 { return &i }

	// On Create, oldPlan is a copy of plan — checks have NO IDs yet
	oldPlan := &ScorecardModel{
		Type:             types.StringValue("LEVEL"),
		EntityFilterType: types.StringValue("entity_types"),
		Levels: map[string]LevelModel{
			"bronze": {
				Name:  types.StringValue("Bronze"),
				Color: types.StringValue("#FB923C"),
				Rank:  types.Int32Value(1),
			},
			"silver": {
				Name:  types.StringValue("Silver"),
				Color: types.StringValue("#C0C0C0"),
				Rank:  types.Int32Value(2),
			},
		},
		Checks: map[string]CheckModel{
			"test_check": {
				Name:              types.StringValue("Test Check"),
				ScorecardLevelKey: types.StringValue("bronze"),
				Ordering:          types.Int32Value(0),
			},
			"another_check": {
				Name:              types.StringValue("Another Check"),
				ScorecardLevelKey: types.StringValue("bronze"),
				Ordering:          types.Int32Value(1),
			},
			"neat_silver_check": {
				Name:              types.StringValue("Neat Silver Check"),
				ScorecardLevelKey: types.StringValue("silver"),
				Ordering:          types.Int32Value(0),
			},
		},
	}

	// API response after Create — checks now have IDs assigned
	apiResp := &dxapi.APIResponse{
		Scorecard: dxapi.APIScorecard{
			Id:                  "scorecard-1",
			Name:                "Test Scorecard",
			Type:                "LEVEL",
			EntityFilterType:    "entity_types",
			EvaluationFrequency: 2,
			Levels: []*dxapi.APILevel{
				{Id: strPtr("level-1"), Name: strPtr("Bronze"), Color: strPtr("#FB923C"), Rank: int32Ptr(1)},
				{Id: strPtr("level-2"), Name: strPtr("Silver"), Color: strPtr("#C0C0C0"), Rank: int32Ptr(2)},
			},
			Checks: []*dxapi.APICheck{
				{Id: strPtr("id-111"), Name: strPtr("Test Check"), Ordering: 0, OutputType: strPtr("string"), Sql: strPtr("SELECT 1")},
				{Id: strPtr("id-222"), Name: strPtr("Another Check"), Ordering: 1, OutputType: strPtr("string"), Sql: strPtr("SELECT 2")},
				{Id: strPtr("id-333"), Name: strPtr("Neat Silver Check"), Ordering: 0, OutputType: strPtr("string"), Sql: strPtr("SELECT 3")},
			},
		},
	}

	state := &ScorecardModel{}
	responseBodyToModel(ctx, apiResp, state, oldPlan)

	// Verify scorecard_level_key is preserved from the plan
	expectedLevelKeys := map[string]string{
		"test_check":        "bronze",
		"another_check":     "bronze",
		"neat_silver_check": "silver",
	}

	for key, expectedLevelKey := range expectedLevelKeys {
		check, ok := state.Checks[key]
		if !ok {
			t.Errorf("expected key %q in state.Checks, not found", key)
			continue
		}
		if check.ScorecardLevelKey.IsNull() {
			t.Errorf("key %q: scorecard_level_key is null, expected %q", key, expectedLevelKey)
			continue
		}
		if check.ScorecardLevelKey.ValueString() != expectedLevelKey {
			t.Errorf("key %q: scorecard_level_key = %q, expected %q", key, check.ScorecardLevelKey.ValueString(), expectedLevelKey)
		}
	}
}
