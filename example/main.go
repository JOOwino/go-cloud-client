package main

import (
	"fmt"
	"log"
	"time"

	gocloudclient "go-cloud-client"
)

func main() {
	// Create a new client
	client, err := gocloudclient.NewClient(gocloudclient.ClientConfig{
		BaseURL:  "http://localhost:8888",
		Username: "", // Optional: if your config server requires auth
		Password: "", // Optional
		Timeout:  30 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Fetch configuration
	config, err := client.GetConfig("myapp", "dev", "master")
	if err != nil {
		log.Fatalf("Failed to fetch config: %v", err)
	}

	fmt.Printf("Application: %s\n", config.Name)
	fmt.Printf("Profiles: %v\n", config.Profiles)
	fmt.Printf("Label: %s\n", config.Label)

	// Get specific values
	dbHost, _ := config.GetString("spring.datasource.host")
	dbPort, _ := config.GetInt("spring.datasource.port")
	debugMode, _ := config.GetBool("app.debug")

	fmt.Printf("\nDatabase Host: %s\n", dbHost)
	fmt.Printf("Database Port: %d\n", dbPort)
	fmt.Printf("Debug Mode: %v\n", debugMode)

	// Get all properties as a map
	allProperties := config.GetPropertySources()
	fmt.Printf("\nTotal properties: %d\n", len(allProperties))

	// Convert to YAML
	yamlConfig, err := config.ToYAML()
	if err == nil {
		fmt.Printf("\nYAML Configuration:\n%s\n", yamlConfig)
	}

	// Example with cached client
	fmt.Println("\n=== Using Cached Client ===")
	cachedClient := gocloudclient.NewCachedClient(client, 5*time.Minute)

	// First call - fetches from server
	config1, err := cachedClient.GetConfig("myapp", "dev", "master")
	if err != nil {
		log.Fatalf("Failed to fetch config: %v", err)
	}
	fmt.Printf("Config fetched (1st call): %s\n", config1.Name)

	// Second call - returns from cache
	config2, err := cachedClient.GetConfig("myapp", "dev", "master")
	if err != nil {
		log.Fatalf("Failed to fetch config: %v", err)
	}
	fmt.Printf("Config fetched (2nd call, from cache): %s\n", config2.Name)

	// Clear cache
	cachedClient.ClearCache()
	fmt.Println("Cache cleared")
}
