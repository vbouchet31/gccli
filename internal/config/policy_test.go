package config_test

import (
	"testing"

	"github.com/bpauli/gccli/internal/config"
)

func TestPolicyAllowlist(t *testing.T) {
	p := &config.Policy{
		Mode:  config.PolicyModeAllow,
		Allow: []string{"health", "activities list"},
	}

	cases := []struct {
		path    string
		wantErr bool
	}{
		{"health summary", false},
		{"health sleep", false},
		{"activities list", false},
		{"activity delete", true},
		{"workouts delete", true},
		{"auth token", true},
	}

	for _, tc := range cases {
		err := p.Check(tc.path)
		if (err != nil) != tc.wantErr {
			t.Errorf("Check(%q): got err=%v, wantErr=%v", tc.path, err, tc.wantErr)
		}
	}
}

func TestPolicyDenylist(t *testing.T) {
	p := &config.Policy{
		Mode: config.PolicyModeDeny,
		Deny: []string{"activity delete", "workouts delete", "auth token"},
	}

	cases := []struct {
		path    string
		wantErr bool
	}{
		{"health summary", false},
		{"activities list", false},
		{"activity delete", true},
		{"workouts delete", true},
		{"auth token", true},
	}

	for _, tc := range cases {
		err := p.Check(tc.path)
		if (err != nil) != tc.wantErr {
			t.Errorf("Check(%q): got err=%v, wantErr=%v", tc.path, err, tc.wantErr)
		}
	}
}

func TestPolicyNil(t *testing.T) {
	var p *config.Policy
	if err := p.Check("activity delete"); err != nil {
		t.Errorf("nil policy should be a no-op, got: %v", err)
	}
}

func TestPolicyPrefixMatching(t *testing.T) {
	p := &config.Policy{
		Mode:  config.PolicyModeAllow,
		Allow: []string{"health"},
	}
	// "health" prefix should allow all health subcommands
	for _, path := range []string{"health summary", "health sleep", "health hrv", "health body-battery"} {
		if err := p.Check(path); err != nil {
			t.Errorf("expected %q to be allowed by prefix, got: %v", path, err)
		}
	}
	// Non-health commands should be denied
	if err := p.Check("activity delete"); err == nil {
		t.Error("expected 'activity delete' to be denied")
	}
}

func TestPolicyExactMatch(t *testing.T) {
	p := &config.Policy{
		Mode:  config.PolicyModeAllow,
		Allow: []string{"activities list"},
	}
	// Exact match should be allowed
	if err := p.Check("activities list"); err != nil {
		t.Errorf("expected exact match to be allowed, got: %v", err)
	}
	// Prefix of exact match should be denied
	if err := p.Check("activities"); err == nil {
		t.Error("expected 'activities' alone to be denied when only 'activities list' is in allowlist")
	}
}
