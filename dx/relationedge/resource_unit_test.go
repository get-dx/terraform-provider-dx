package relationedge

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestParseEdgeIDRoundTrip(t *testing.T) {
	model := RelationEdgeModel{
		RelationIdentifier:     types.StringValue("depends-on"),
		SourceEntityIdentifier: types.StringValue("team/authn-service"),
		TargetEntityIdentifier: types.StringValue("authn-service-events"),
	}

	relation, source, target, err := parseEdgeID(edgeID(model))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if relation != "depends-on" || source != "team/authn-service" || target != "authn-service-events" {
		t.Fatalf("unexpected parts: %q, %q, %q", relation, source, target)
	}
}

func TestParseEdgeIDRejectsMalformedIDs(t *testing.T) {
	for _, id := range []string{
		"",
		"relation",
		"relation+source",
		"relation+source+target+extra",
		"+source+target",
		"relation++target",
		"relation+source+",
		"relation/source/target",
	} {
		if _, _, _, err := parseEdgeID(id); err == nil {
			t.Errorf("expected error for %q", id)
		}
	}
}
