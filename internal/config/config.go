package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Profile holds the API URL and token for a named profile.
type Profile struct {
	APIURL   string `json:"api_url"`
	APIToken string `json:"api_token"`
}

// configFile represents the on-disk config format.
// Supports both the new profile format and the legacy flat format.
type configFile struct {
	// New format
	CurrentProfile string              `json:"current_profile,omitempty"`
	Profiles       map[string]*Profile `json:"profiles,omitempty"`

	// Legacy flat format (backwards compatibility)
	APIURL   string `json:"api_url,omitempty"`
	APIToken string `json:"api_token,omitempty"`
}

// Resolve builds a Profile by merging values from flags, environment variables,
// and the config file (~/.mvdata/config.json). Precedence: flags > env > file.
func Resolve(flagURL, flagToken, flagProfile string) (*Profile, error) {
	cfg := &Profile{}

	// Load the active profile from file (lowest precedence).
	if err := loadProfile(cfg, flagProfile); err != nil {
		return nil, err
	}

	// Environment variables override file.
	if v := os.Getenv("MVDATA_API_URL"); v != "" {
		cfg.APIURL = v
	}
	if v := os.Getenv("MVDATA_API_TOKEN"); v != "" {
		cfg.APIToken = v
	}

	// Flags override everything.
	if flagURL != "" {
		cfg.APIURL = flagURL
	}
	if flagToken != "" {
		cfg.APIToken = flagToken
	}

	if cfg.APIURL == "" {
		return nil, fmt.Errorf("API URL is required (set --api-url, MVDATA_API_URL, or api_url in ~/.mvdata/config.json)")
	}
	if cfg.APIToken == "" {
		return nil, fmt.Errorf("API token is required (set --api-token, MVDATA_API_TOKEN, or api_token in ~/.mvdata/config.json)")
	}

	return cfg, nil
}

// Save writes a profile to the config file. If the profile doesn't exist, it is
// created. If the file uses the legacy flat format, it is migrated to profiles.
func Save(profileName string, profile *Profile) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	cf := &configFile{
		Profiles: make(map[string]*Profile),
	}

	// Load existing config if present.
	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, cf)
	}

	// Migrate legacy flat format.
	if cf.Profiles == nil {
		cf.Profiles = make(map[string]*Profile)
	}
	if cf.APIURL != "" || cf.APIToken != "" {
		if _, ok := cf.Profiles["default"]; !ok {
			cf.Profiles["default"] = &Profile{
				APIURL:   cf.APIURL,
				APIToken: cf.APIToken,
			}
		}
		cf.APIURL = ""
		cf.APIToken = ""
	}

	cf.Profiles[profileName] = profile

	if cf.CurrentProfile == "" {
		cf.CurrentProfile = profileName
	}

	out, err := json.MarshalIndent(cf, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if err := os.WriteFile(path, append(out, '\n'), 0o600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// resolveProfileName determines which profile to use.
// Precedence: flag > env > file's current_profile > "default".
func resolveProfileName(flagProfile string) (string, error) {
	if flagProfile != "" {
		return flagProfile, nil
	}
	if v := os.Getenv("MVDATA_PROFILE"); v != "" {
		return v, nil
	}

	path, err := configPath()
	if err != nil {
		return "default", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "default", nil
	}

	var cf configFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return "default", nil
	}
	if cf.CurrentProfile != "" {
		return cf.CurrentProfile, nil
	}
	return "default", nil
}

func loadProfile(p *Profile, flagProfile string) error {
	path, err := configPath()
	if err != nil {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading config file: %w", err)
	}

	var cf configFile
	if err := json.Unmarshal(data, &cf); err != nil {
		return fmt.Errorf("parsing config file %s: %w", path, err)
	}

	// New profile format.
	if cf.Profiles != nil {
		profileName, err := resolveProfileName(flagProfile)
		if err != nil {
			return err
		}
		if profile, ok := cf.Profiles[profileName]; ok {
			p.APIURL = profile.APIURL
			p.APIToken = profile.APIToken
		}
		return nil
	}

	// Legacy flat format.
	p.APIURL = cf.APIURL
	p.APIToken = cf.APIToken
	return nil
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determining home directory: %w", err)
	}
	return filepath.Join(home, ".mvdata", "config.json"), nil
}
