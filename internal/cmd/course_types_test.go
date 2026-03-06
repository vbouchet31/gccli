package cmd

import (
	"strings"
	"testing"
)

func TestResolveCourseActivityType_Valid(t *testing.T) {
	tests := []struct {
		key  string
		want int
	}{
		{"running", 1},
		{"cycling", 2},
		{"hiking", 3},
		{"gravel_cycling", 143},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, err := resolveCourseActivityType(tt.key)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestResolveCourseActivityType_Unknown(t *testing.T) {
	_, err := resolveCourseActivityType("swimming")
	if err == nil {
		t.Fatal("expected error for unknown type")
	}
	if !strings.Contains(err.Error(), "swimming") {
		t.Errorf("error should mention the invalid type: %v", err)
	}
	if !strings.Contains(err.Error(), "cycling") {
		t.Errorf("error should list valid types: %v", err)
	}
}
