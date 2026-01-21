package gocloudclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  ClientConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: ClientConfig{
				BaseURL: "http://localhost:8888",
			},
			wantErr: false,
		},
		{
			name: "empty baseURL",
			config: ClientConfig{
				BaseURL: "",
			},
			wantErr: true,
		},
		{
			name: "baseURL with trailing slash",
			config: ClientConfig{
				BaseURL: "http://localhost:8888/",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestGetConfig(t *testing.T) {
	// Mock Spring Cloud Config Server response
	mockResponse := ConfigResponse{
		Name:     "myapp",
		Profiles: []string{"dev"},
		Label:    "master",
		PropertySources: []PropertySource{
			{
				Name: "application-dev.yml",
				Source: map[string]interface{}{
					"spring.datasource.host": "localhost",
					"spring.datasource.port": 5432,
					"app.debug":              true,
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/myapp/dev/master" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client, err := NewClient(ClientConfig{
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	config, err := client.GetConfig("myapp", "dev", "master")
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	if config.Name != "myapp" {
		t.Errorf("Expected name 'myapp', got '%s'", config.Name)
	}

	// Test value retrieval
	host, exists := config.GetString("spring.datasource.host")
	if !exists || host != "localhost" {
		t.Errorf("GetString() = %v, %v, want 'localhost', true", host, exists)
	}

	port, exists := config.GetInt("spring.datasource.port")
	if !exists || port != 5432 {
		t.Errorf("GetInt() = %v, %v, want 5432, true", port, exists)
	}

	debug, exists := config.GetBool("app.debug")
	if !exists || !debug {
		t.Errorf("GetBool() = %v, %v, want true, true", debug, exists)
	}
}

func TestGetConfigWithDefaults(t *testing.T) {
	mockResponse := ConfigResponse{
		Name:     "myapp",
		Profiles: []string{"default"},
		Label:    "master",
		PropertySources: []PropertySource{
			{
				Name:   "application-default.yml",
				Source: map[string]interface{}{},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client, _ := NewClient(ClientConfig{BaseURL: server.URL})

	// Test default profile
	_, err := client.GetConfig("myapp", "", "")
	if err != nil {
		t.Errorf("GetConfig() with empty profile/label should work: %v", err)
	}
}

func TestCachedClient(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		mockResponse := ConfigResponse{
			Name:     "myapp",
			Profiles: []string{"dev"},
			Label:    "master",
			PropertySources: []PropertySource{
				{
					Name:   "application-dev.yml",
					Source: map[string]interface{}{"test": "value"},
				},
			},
		}
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	client, _ := NewClient(ClientConfig{BaseURL: server.URL})
	cachedClient := NewCachedClient(client, 5*time.Minute)

	// First call - should hit server
	_, err := cachedClient.GetConfig("myapp", "dev", "master")
	if err != nil {
		t.Fatalf("First GetConfig() failed: %v", err)
	}

	// Second call - should use cache
	_, err = cachedClient.GetConfig("myapp", "dev", "master")
	if err != nil {
		t.Fatalf("Second GetConfig() failed: %v", err)
	}

	// Should only have called server once
	if callCount != 1 {
		t.Errorf("Expected 1 server call, got %d", callCount)
	}

	// Clear cache and call again - should hit server
	cachedClient.ClearCache()
	_, err = cachedClient.GetConfig("myapp", "dev", "master")
	if err != nil {
		t.Fatalf("Third GetConfig() failed: %v", err)
	}

	if callCount != 2 {
		t.Errorf("Expected 2 server calls after cache clear, got %d", callCount)
	}
}

func TestGetPropertySources(t *testing.T) {
	config := &ConfigResponse{
		PropertySources: []PropertySource{
			{
				Name: "first",
				Source: map[string]interface{}{
					"key1": "value1",
					"key2": "old",
				},
			},
			{
				Name: "second",
				Source: map[string]interface{}{
					"key2": "new", // Should override "old"
					"key3": "value3",
				},
			},
		},
	}

	props := config.GetPropertySources()

	if props["key1"] != "value1" {
		t.Errorf("key1 = %v, want 'value1'", props["key1"])
	}

	if props["key2"] != "new" {
		t.Errorf("key2 = %v, want 'new' (should override 'old')", props["key2"])
	}

	if props["key3"] != "value3" {
		t.Errorf("key3 = %v, want 'value3'", props["key3"])
	}
}
