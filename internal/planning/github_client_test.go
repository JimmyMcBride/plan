package planning

import "testing"

func TestDecodeGraphQLResponseReturnsErrors(t *testing.T) {
	raw := []byte(`{"errors":[{"message":"discussion access denied"},{"message":"secondary failure"}]}`)
	var target struct{}
	err := decodeGraphQLResponse(raw, &target)
	if err == nil {
		t.Fatal("expected graphql errors to be returned")
	}
	if got := err.Error(); got != "graphql request failed: discussion access denied; secondary failure" {
		t.Fatalf("unexpected graphql error: %v", err)
	}
}

func TestDecodeGraphQLResponseAllowsNilTarget(t *testing.T) {
	raw := []byte(`{"data":{"ok":true}}`)
	if err := decodeGraphQLResponse(raw, nil); err != nil {
		t.Fatalf("expected nil target to succeed: %v", err)
	}
}
