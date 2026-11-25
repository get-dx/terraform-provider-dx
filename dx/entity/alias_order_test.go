package entity

import (
	"context"
	"testing"

	"terraform-provider-dx/dx/dxapi"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResponseBodyToModelAliasOrder(t *testing.T) {
	ctx := context.Background()
	instance := "1"
	alias := func(id string, inst *string) AliasModel {
		m := AliasModel{Identifier: types.StringValue(id), InstanceIdentifier: types.StringNull()}
		if inst != nil {
			m.InstanceIdentifier = types.StringValue(*inst)
		}
		return m
	}
	// DX sorts alias identifiers as strings.
	apiResp := &dxapi.APIEntityResponse{Entity: dxapi.APIEntity{
		Identifier: "svc",
		Type:       "service",
		Aliases: map[string][]dxapi.APIAlias{
			"github_repo": {{Identifier: "1310008280"}, {Identifier: "209154742"}},
			"gitlab_repo": {{Identifier: "12", InstanceIdentifier: &instance}, {Identifier: "9", InstanceIdentifier: &instance}},
		},
	}}

	t.Run("same aliases keep the configured order", func(t *testing.T) {
		plan := EntityResourceModel{Aliases: map[string][]AliasModel{
			"github_repo": {alias("209154742", nil), alias("1310008280", nil)},
			"gitlab_repo": {alias("9", &instance), alias("12", &instance)},
		}}
		oldPlan := plan
		responseBodyToModel(ctx, apiResp, &plan, &oldPlan)

		want := map[string][]string{"github_repo": {"209154742", "1310008280"}, "gitlab_repo": {"9", "12"}}
		for aliasType, ids := range want {
			got := plan.Aliases[aliasType]
			if len(got) != len(ids) {
				t.Fatalf("%s: want %d aliases, got %d", aliasType, len(ids), len(got))
			}
			for i, id := range ids {
				if got[i].Identifier.ValueString() != id {
					t.Errorf("%s[%d]: want %s, got %s", aliasType, i, id, got[i].Identifier.ValueString())
				}
			}
		}
	})

	t.Run("different aliases use the API value", func(t *testing.T) {
		plan := EntityResourceModel{Aliases: map[string][]AliasModel{
			"github_repo": {alias("209154742", nil), alias("999", nil)},
		}}
		oldPlan := plan
		responseBodyToModel(ctx, apiResp, &plan, &oldPlan)

		got := plan.Aliases["github_repo"]
		if len(got) != 2 || got[0].Identifier.ValueString() != "1310008280" || got[1].Identifier.ValueString() != "209154742" {
			t.Errorf("want API aliases [1310008280 209154742], got %v", got)
		}
	})
}
