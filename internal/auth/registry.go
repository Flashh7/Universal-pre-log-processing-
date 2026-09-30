package auth

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	ID     string `yaml:"id"`
	APIKey string `yaml:"api_key"`
}

type Config struct {
	Applications []AppConfig `yaml:"applications"`
}

type Registry struct {
	apps map[string]string // map[ApplicationID]APIKey
}

func LoadRegistry(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	apps := make(map[string]string)
	for _, app := range cfg.Applications {
		apps[app.ID] = app.APIKey
	}

	return &Registry{apps: apps}, nil
}

func (r *Registry) Authenticate(appID, apiKey string) error {
	expectedKey, exists := r.apps[appID]
	if !exists {
		return errors.New("unknown application ID")
	}
	if expectedKey != apiKey {
		return errors.New("invalid API key")
	}
	return nil
}

func (r *Registry) Exists(appID string) bool {
	_, exists := r.apps[appID]
	return exists
}
