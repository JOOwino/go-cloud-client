package gocloudclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Client represents a Spring Cloud Config Server client
type Client struct {
	baseURL    string
	httpClient *http.Client
	username   string
	password   string
}

// ConfigResponse represents the response from Spring Cloud Config Server
type ConfigResponse struct {
	Name            string           `json:"name"`
	Profiles        []string         `json:"profiles"`
	Label           string           `json:"label,omitempty"`
	Version         string           `json:"version,omitempty"`
	State           string           `json:"state,omitempty"`
	PropertySources []PropertySource `json:"propertySources"`
}

type ConfigEnvVariables struct {
}

// PropertySource represents a property source in the config response
type PropertySource struct {
	Name   string                 `json:"name"`
	Source map[string]interface{} `json:"source"`
}

// ClientConfig holds configuration for creating a new client
type ClientConfig struct {
	BaseURL    string
	Username   string
	Password   string
	Timeout    time.Duration
	HTTPClient *http.Client
}

const (
	APP_PREFIX = "app"
	YAML       = "yaml"
	TOML       = "toml"
	JSON       = "json"
	ENV        = "enc=v"
)

func init() {
	//Get File Contents

	//Check if there is a .env file and load data. Should be in a specific format
	//The data should have a defined prefix(going with app.{conf-format}

	switch {

	}

}

func fetchConfigFile(filename string) ([]byte, error) {
	_, err := os.Stat(filename)
	if err != nil {
		return nil, errors.New("error fetching file")
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// NewClient creates a new Spring Cloud Config Server client
func NewClient(config ClientConfig) (*Client, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("baseURL is required")
	}

	// Remove trailing slash if present
	baseURL := strings.TrimSuffix(config.BaseURL, "/")

	httpClient := config.HTTPClient
	if httpClient == nil {
		timeout := config.Timeout
		if timeout == 0 {
			timeout = 30 * time.Second
		}
		httpClient = &http.Client{
			Timeout: timeout,
		}
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		username:   config.Username,
		password:   config.Password,
	}, nil
}

// GetConfig fetches configuration from Spring Cloud Config Server
// Parameters:
//   - application: The application name (e.g., "myapp")
//   - profile: The profile (e.g., "dev", "prod"). Can be comma-separated for multiple profiles
//   - label: Optional label/branch (e.g., "master", "develop"). Defaults to "master" if empty
func (c *Client) GetConfig(application, profile, label string) (*ConfigResponse, error) {
	if application == "" {
		return nil, fmt.Errorf("application name is required")
	}

	if profile == "" {
		profile = "default"
	}

	if label == "" {
		label = "master"
	}

	// Build the URL: {baseURL}/{application}/{profile}/{label}
	url := fmt.Sprintf("%s/%s/%s/%s", c.baseURL, application, profile, label)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set basic auth if credentials are provided
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	// Set Accept header to prefer JSON
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("config server returned status %d: %s", resp.StatusCode, string(body))
	}

	var configResp ConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&configResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &configResp, nil
}

// GetConfigYAML fetches configuration and returns it in YAML format
func (c *Client) GetConfigYAML(application, profile, label string) (string, error) {
	if application == "" {
		return "", fmt.Errorf("application name is required")
	}

	if profile == "" {
		profile = "default"
	}

	if label == "" {
		label = "master"
	}

	url := fmt.Sprintf("%s/%s-%s.yml", c.baseURL, application, profile)
	if label != "" && label != "master" {
		url = fmt.Sprintf("%s/%s/%s-%s.yml", c.baseURL, label, application, profile)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	req.Header.Set("Accept", "application/x-yaml, text/yaml, text/x-yaml")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("config server returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// GetPropertySources returns all property sources flattened into a single map
// Later property sources (in the array) override earlier ones
func (configResp *ConfigResponse) GetPropertySources() map[string]interface{} {
	result := make(map[string]interface{})

	// Iterate in reverse order so later sources override earlier ones
	for i := len(configResp.PropertySources) - 1; i >= 0; i-- {
		ps := configResp.PropertySources[i]
		for k, v := range ps.Source {
			result[k] = v
		}
	}

	return result
}

// GetValue retrieves a configuration value by key
func (configResp *ConfigResponse) GetValue(key string) (interface{}, bool) {
	properties := configResp.GetPropertySources()
	value, exists := properties[key]
	return value, exists
}

// GetString retrieves a configuration value as a string
func (configResp *ConfigResponse) GetString(key string) (string, bool) {
	value, exists := configResp.GetValue(key)
	if !exists {
		return "", false
	}

	str, ok := value.(string)
	if ok {
		return str, true
	}

	// Try to convert other types to string
	return fmt.Sprintf("%v", value), true
}

// GetInt retrieves a configuration value as an int
func (configResp *ConfigResponse) GetInt(key string) (int, bool) {
	value, exists := configResp.GetValue(key)
	if !exists {
		return 0, false
	}

	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// GetBool retrieves a configuration value as a bool
func (configResp *ConfigResponse) GetBool(key string) (bool, bool) {
	value, exists := configResp.GetValue(key)
	if !exists {
		return false, false
	}

	switch v := value.(type) {
	case bool:
		return v, true
	case string:
		return strings.ToLower(v) == "true" || v == "1", true
	default:
		return false, false
	}
}

// ToYAML converts the configuration response to YAML format
func (configResp *ConfigResponse) ToYAML() (string, error) {
	properties := configResp.GetPropertySources()
	data, err := yaml.Marshal(properties)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %w", err)
	}
	return string(data), nil
}

// ToJSON converts the configuration response to JSON format
func (configResp *ConfigResponse) ToJSON() (string, error) {
	properties := configResp.GetPropertySources()
	data, err := json.MarshalIndent(properties, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(data), nil
}
