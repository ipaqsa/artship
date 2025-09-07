package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Auth struct {
	Path       string                  `json:"path" yaml:"Path"`
	Registries map[string]RegistryAuth `json:"registries" yaml:"Registries"`
}

// RegistryAuth represents authentication for a specific registry
type RegistryAuth struct {
	Registry      string `json:"registry" yaml:"Registry"`
	Username      string `json:"username,omitempty" yaml:"Username,omitempty"`
	Authenticated bool   `json:"authenticated" yaml:"Authenticated"`
}

// String displays authentication information in human-readable format
func (a *Auth) String() string {
	marshalled, _ := yaml.Marshal(a)

	return string(marshalled)
}

func (c *Client) GetAuth() (*Auth, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get user home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".docker", "config.json")
	c.logger.Debug("Reading Docker config from: %s", configPath)

	// Check if config file exists
	if _, err = os.Stat(configPath); os.IsNotExist(err) {
		c.logger.Debug("Docker config file not found")
		return &Auth{
			Path:       configPath,
			Registries: make(map[string]RegistryAuth),
		}, nil
	}

	// Read the config file
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read Docker config file: %w", err)
	}

	var config struct {
		Auths       map[string]interface{} `json:"auths,omitempty"`
		CredHelpers map[string]string      `json:"credHelpers,omitempty"`
	}

	if err = json.Unmarshal(raw, &config); err != nil {
		return nil, fmt.Errorf("parse Docker auth config: %w", err)
	}

	c.logger.Debug("Found %d auth entries and %d credential helpers", len(config.Auths), len(config.CredHelpers))

	// Extract authentication information
	authInfo := &Auth{
		Path:       configPath,
		Registries: make(map[string]RegistryAuth),
	}

	// Process auths
	for registry, authData := range config.Auths {
		regAuth := RegistryAuth{
			Registry:      registry,
			Authenticated: true,
		}

		// Try to extract username if available
		if authMap, ok := authData.(map[string]interface{}); ok {
			if username, exists := authMap["username"].(string); exists {
				regAuth.Username = username
			}
			// Note: We don't extract passwords for security reasons
		}

		authInfo.Registries[registry] = regAuth
	}

	// Process credential helpers
	for registry, helper := range config.CredHelpers {
		if _, exists := authInfo.Registries[registry]; !exists {
			authInfo.Registries[registry] = RegistryAuth{
				Registry: registry,
				Username: fmt.Sprintf("(managed by %s)", helper),
			}
		}
	}

	return authInfo, nil
}
