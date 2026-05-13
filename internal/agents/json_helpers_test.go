package agents

import "testing"

func TestUnmarshalCapabilities_Valid(t *testing.T) {
	t.Parallel()
	caps, err := unmarshalCapabilities(`["go","git"]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(caps) != 2 || caps[0] != "go" || caps[1] != "git" {
		t.Errorf("unexpected caps: %v", caps)
	}
}

func TestUnmarshalCapabilities_Empty(t *testing.T) {
	t.Parallel()
	caps, err := unmarshalCapabilities(`[]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(caps) != 0 {
		t.Errorf("expected empty slice, got: %v", caps)
	}
}

func TestUnmarshalCapabilities_InvalidJSON(t *testing.T) {
	t.Parallel()
	_, err := unmarshalCapabilities(`not-json`)
	if err == nil {
		t.Fatal("expected error for invalid json")
	}
}
