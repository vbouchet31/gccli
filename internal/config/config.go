package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Environment variable names for configuration overrides.
const (
	EnvAccount        = "GCCLI_ACCOUNT"
	EnvDomain         = "GCCLI_DOMAIN"
	EnvColor          = "GCCLI_COLOR"
	EnvJSON           = "GCCLI_JSON"
	EnvPlain          = "GCCLI_PLAIN"
	EnvKeyringBackend = "GCCLI_KEYRING_BACKEND"
	EnvPolicy         = "GCCLI_POLICY"
)

// PolicyMode defines whether the policy operates as an allowlist or denylist.
type PolicyMode string

const (
	PolicyModeAllow PolicyMode = "allowlist"
	PolicyModeDeny  PolicyMode = "denylist"
)

// Policy defines the command execution policy.
// Commands are matched by prefix against the full subcommand path (e.g. "activity delete").
type Policy struct {
	Mode  PolicyMode `json:"mode"`
	Allow []string   `json:"allow,omitempty"`
	Deny  []string   `json:"deny,omitempty"`
}

// Check returns an error if the given command path is not permitted by the policy.
func (p *Policy) Check(commandPath string) error {
	if p == nil {
		return nil
	}
	switch p.Mode {
	case PolicyModeAllow:
		for _, entry := range p.Allow {
			if commandPath == entry || strings.HasPrefix(commandPath, entry+" ") {
				return nil
			}
		}
		return fmt.Errorf("command %q is not in the allowlist", commandPath)
	case PolicyModeDeny:
		for _, entry := range p.Deny {
			if commandPath == entry || strings.HasPrefix(commandPath, entry+" ") {
				return fmt.Errorf("command %q is denied by policy", commandPath)
			}
		}
		return nil
	default:
		return fmt.Errorf("unknown policy mode %q", p.Mode)
	}
}

// File represents the on-disk configuration file.
type File struct {
	DefaultAccount  string              `json:"default_account,omitempty"`
	KeyringBackend  string              `json:"keyring_backend,omitempty"`
	DomainName      string              `json:"domain,omitempty"`
	DefaultFormat   string              `json:"default_format,omitempty"`
	ActivitySummary map[string][]string `json:"activity_summary,omitempty"`
	Policy          *Policy             `json:"policy,omitempty"`
}

// Read loads the configuration from the default config file path.
// If the file does not exist, it returns a zero-value File (no error).
// If GCCLI_POLICY env var is set, the policy is loaded from that file and overrides
// any policy in the main config file.
func Read() (*File, error) {
	p, err := ConfigFilePath()
	if err != nil {
		return nil, err
	}
	f, err := ReadFrom(p)
	if err != nil {
		return nil, err
	}

	// Allow overriding policy from a separate file via env var.
	if policyPath := os.Getenv(EnvPolicy); policyPath != "" {
		data, err := os.ReadFile(policyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read policy file %q: %w", policyPath, err)
		}
		var policyWrapper struct {
			Policy *Policy `json:"policy"`
		}
		if err := json.Unmarshal(data, &policyWrapper); err != nil {
			return nil, fmt.Errorf("failed to parse policy file %q: %w", policyPath, err)
		}
		f.Policy = policyWrapper.Policy
	}

	return f, nil
}

// ReadFrom loads configuration from the given file path.
// If the file does not exist, it returns a zero-value File (no error).
func ReadFrom(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &File{}, nil
		}
		return nil, err
	}
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

// Write saves the configuration to the default config file path.
func Write(f *File) error {
	p, err := ConfigFilePath()
	if err != nil {
		return err
	}
	return WriteTo(f, p)
}

// WriteTo saves the configuration to the given file path,
// creating parent directories as needed.
func WriteTo(f *File, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}

// Domain returns the effective domain, checking the environment variable first,
// then the config file, then falling back to the default.
func (f *File) Domain() string {
	if v := os.Getenv(EnvDomain); v != "" {
		return v
	}
	if f.DomainName != "" {
		return f.DomainName
	}
	return "garmin.com"
}

// KeyringBackendValue returns the effective keyring backend, checking the
// environment variable first, then the config file.
func (f *File) KeyringBackendValue() string {
	if v := os.Getenv(EnvKeyringBackend); v != "" {
		return v
	}
	return f.KeyringBackend
}

// Account returns the effective account email, checking the environment variable
// first, then the config file's default account.
func (f *File) Account() string {
	if v := os.Getenv(EnvAccount); v != "" {
		return v
	}
	return f.DefaultAccount
}

// IsJSON returns true if the GCCLI_JSON environment variable is set to a truthy value.
func IsJSON() bool {
	return isTruthy(os.Getenv(EnvJSON))
}

// IsPlain returns true if the GCCLI_PLAIN environment variable is set to a truthy value.
func IsPlain() bool {
	return isTruthy(os.Getenv(EnvPlain))
}

// ColorMode returns the color mode from the GCCLI_COLOR environment variable.
// Returns empty string if not set.
func ColorMode() string {
	return os.Getenv(EnvColor)
}

func isTruthy(v string) bool {
	switch v {
	case "1", "true", "yes":
		return true
	}
	return false
}
