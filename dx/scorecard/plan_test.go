package scorecard_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"terraform-provider-dx/dx/scorecard"
	"terraform-provider-dx/internal/provider"
)

// TestPlanNewNestedElementsHaveUnknownIds checks that adding a level, check group,
// or check to an existing scorecard plans its id as unknown, not null.
func TestPlanNewNestedElementsHaveUnknownIds(t *testing.T) {
	ctx := context.Background()

	var schemaResp resource.SchemaResponse
	scorecard.NewScorecardResource().Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatal("schema type is not an object")
	}

	attrType := func(name string) tftypes.Type { return objType.AttributeTypes[name] }
	elemType := func(name string) tftypes.Object {
		m, ok := attrType(name).(tftypes.Map)
		if !ok {
			t.Fatalf("%s is not a map", name)
		}
		o, ok := m.ElementType.(tftypes.Object)
		if !ok {
			t.Fatalf("%s elements are not objects", name)
		}
		return o
	}
	// object builds a value of typ with the given attributes and nulls elsewhere.
	object := func(typ tftypes.Object, attrs map[string]tftypes.Value) tftypes.Value {
		vals := map[string]tftypes.Value{}
		for name, at := range typ.AttributeTypes {
			if v, ok := attrs[name]; ok {
				vals[name] = v
			} else {
				vals[name] = tftypes.NewValue(at, nil)
			}
		}
		return tftypes.NewValue(typ, vals)
	}
	str := func(s string) tftypes.Value { return tftypes.NewValue(tftypes.String, s) }
	num := func(n int) tftypes.Value { return tftypes.NewValue(tftypes.Number, n) }
	nullStr := tftypes.NewValue(tftypes.String, nil)

	level := func(id tftypes.Value, name string, rank int) tftypes.Value {
		return object(elemType("levels"), map[string]tftypes.Value{
			"id": id, "name": str(name), "color": str("#cccccc"), "rank": num(rank),
		})
	}
	group := func(id tftypes.Value, name string, ordering int) tftypes.Value {
		return object(elemType("check_groups"), map[string]tftypes.Value{
			"id": id, "name": str(name), "ordering": num(ordering),
		})
	}
	check := func(id tftypes.Value, name string, ordering int) tftypes.Value {
		return object(elemType("checks"), map[string]tftypes.Value{
			"id": id, "name": str(name), "ordering": num(ordering), "sql": str("select 1"),
			"output_enabled": tftypes.NewValue(tftypes.Bool, false),
		})
	}
	scorecardValue := func(id tftypes.Value, levels, groups, checks map[string]tftypes.Value) tftypes.Value {
		return object(objType, map[string]tftypes.Value{
			"id":                 id,
			"name":               str("Security"),
			"type":               str("POINTS"),
			"entity_filter_type": str("entity_types"),
			"levels":             tftypes.NewValue(attrType("levels"), levels),
			"check_groups":       tftypes.NewValue(attrType("check_groups"), groups),
			"checks":             tftypes.NewValue(attrType("checks"), checks),
		})
	}

	prior := scorecardValue(str("sc1"),
		map[string]tftypes.Value{"bronze": level(str("l1"), "Bronze", 1)},
		map[string]tftypes.Value{"existing": group(str("g1"), "Existing", 0)},
		map[string]tftypes.Value{"old_check": check(str("c1"), "Old check", 0)},
	)
	config := scorecardValue(nullStr,
		map[string]tftypes.Value{"bronze": level(nullStr, "Bronze", 1), "silver": level(nullStr, "Silver", 2)},
		map[string]tftypes.Value{"existing": group(nullStr, "Existing", 0), "new_group": group(nullStr, "New group", 1)},
		map[string]tftypes.Value{"old_check": check(nullStr, "Old check", 0), "new_check": check(nullStr, "New check", 1)},
	)
	// Terraform core proposes prior values for computed attributes and null for new elements.
	proposed := scorecardValue(str("sc1"),
		map[string]tftypes.Value{"bronze": level(str("l1"), "Bronze", 1), "silver": level(nullStr, "Silver", 2)},
		map[string]tftypes.Value{"existing": group(str("g1"), "Existing", 0), "new_group": group(nullStr, "New group", 1)},
		map[string]tftypes.Value{"old_check": check(str("c1"), "Old check", 0), "new_check": check(nullStr, "New check", 1)},
	)

	dyn := func(v tftypes.Value) *tfprotov6.DynamicValue {
		d, err := tfprotov6.NewDynamicValue(objType, v)
		if err != nil {
			t.Fatal(err)
		}
		return &d
	}

	server, err := providerserver.NewProtocol6WithError(provider.New("test")())()
	if err != nil {
		t.Fatal(err)
	}
	resp, err := server.PlanResourceChange(ctx, &tfprotov6.PlanResourceChangeRequest{
		TypeName:         "dx_scorecard",
		PriorState:       dyn(prior),
		Config:           dyn(config),
		ProposedNewState: dyn(proposed),
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range resp.Diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			t.Fatalf("plan error: %s: %s", d.Summary, d.Detail)
		}
	}
	planned, err := resp.PlannedState.Unmarshal(objType)
	if err != nil {
		t.Fatal(err)
	}

	idAt := func(attr, key string) tftypes.Value {
		v, _, err := tftypes.WalkAttributePath(planned,
			tftypes.NewAttributePath().WithAttributeName(attr).WithElementKeyString(key).WithAttributeName("id"))
		if err != nil {
			t.Fatalf("%s[%q].id: %s", attr, key, err)
		}
		val, ok := v.(tftypes.Value)
		if !ok {
			t.Fatalf("%s[%q].id is not a value", attr, key)
		}
		return val
	}

	for _, tc := range []struct{ attr, existing, existingID, added string }{
		{"levels", "bronze", "l1", "silver"},
		{"check_groups", "existing", "g1", "new_group"},
		{"checks", "old_check", "c1", "new_check"},
	} {
		if !idAt(tc.attr, tc.existing).Equal(str(tc.existingID)) {
			t.Errorf("%s[%q].id: want %q kept from state, got %s", tc.attr, tc.existing, tc.existingID, idAt(tc.attr, tc.existing))
		}
		if v := idAt(tc.attr, tc.added); v.IsKnown() {
			t.Errorf("%s[%q].id: want unknown, got %s", tc.attr, tc.added, v)
		}
	}
}
