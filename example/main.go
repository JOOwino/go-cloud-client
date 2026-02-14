package main

import (
	"fmt"
	"log"
	"time"

	gocloudclient "go-cloud-client"
)

func main() {
	// Create a new client
	client, err := gocloudclient.NewClient()

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Fetch configuration
	config, err := client.GetConfig(gocloudclient.YamlDecoder{})
	if err != nil {
		log.Fatalf("Failed to fetch config: %v", err)
	}

	fmt.Printf("Profiles: %v\n", config.Profiles)
	fmt.Printf("Label: %s\n", config.Label)

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
