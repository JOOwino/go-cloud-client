package gocloudclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// AppClient represents a Spring Cloud Config Server client
type AppClient struct {
	BaseURL         string `yaml:"base_url"`
	Timeout         int    `yaml:"timeout"`
	httpClient      *http.Client
	ApplicationName string `yaml:"application_name"`
	Profile         string `yaml:"profile"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
}

type AppClientConfig struct {
	BaseURL         string `yaml:"base_url"`
	Timeout         int    `yaml:"timeout"`
	httpClient      *http.Client
	ApplicationName string `yaml:"application_name"`
	Profile         string `yaml:"profile"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
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
	Timeout    int
	HTTPClient *http.Client
}

// NewClient creates a new Spring Cloud Config Server client
func NewClient() (*AppClient, error) {
	dir, _ := os.Getwd()
	data, err := os.ReadFile(dir + "/app.yaml")
	if err != nil {
		return nil, err
	}
	config := &AppClient{}
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	fmt.Printf("Config: %v\n", config)

	if config.BaseURL == "" {
		return nil, fmt.Errorf("baseURL is required")
	}

	// Remove trailing slash if present
	config.BaseURL = strings.TrimSuffix(config.BaseURL, "/")

	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	config.httpClient = &http.Client{
		Timeout: timeout,
	}
	return config, nil
}

// GetConfig fetches configuration from Spring Cloud Config Server
// Parameters:
//   - application: The application name (e.g., "myapp")
//   - profile: The profile (e.g., "dev", "prod"). Can be comma-separated for multiple profiles
//   - label: Optional label/branch (e.g., "master", "develop"). Defaults to "master" if empty
func (c *AppClient) GetConfig() (*ConfigResponse, error) {
	if c.ApplicationName == "" {
		return nil, fmt.Errorf("application name is required")
	}

	if c.Profile == "" {
		c.ApplicationName = "default"
	}

	// Build the URL: {baseURL}/{application}/{profile}/{label}
	url := fmt.Sprintf("%s/%s/%s", c.BaseURL, c.ApplicationName, c.Profile)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set basic auth if credentials are provided
	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}

	fmt.Println("Debug-2")

	// Set Accept header to prefer JSON
	req.Header.Set("Accept", "application/json")

	fmt.Printf("Timeout: %v\n", c.httpClient.Timeout)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("config server returned status %d: %s", resp.StatusCode, string(body))
	}
	fmt.Println("Debug-02")

	var configResp ConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&configResp); err != nil {
		fmt.Println("Debug-03")
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	fmt.Println("Debug-04")

	return &configResp, nil
}

func (c *AppClient) Sample() error {
	return nil
}

// GetConfigYAML fetches configuration and returns it in YAML format
func (c *AppClient) GetConfigYAML(application, profile, label string) (string, error) {
	if application == "" {
		return "", fmt.Errorf("application name is required")
	}

	if profile == "" {
		profile = "default"
	}

	if label == "" {
		label = "master"
	}

	url := fmt.Sprintf("%s/%s-%s.yml", c.BaseURL, application, profile)
	if label != "" && label != "master" {
		url = fmt.Sprintf("%s/%s/%s-%s.yml", c.BaseURL, label, application, profile)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
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
