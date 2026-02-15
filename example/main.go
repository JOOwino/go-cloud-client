package main

import (
	"fmt"
	"log"
	"time"

	gocloudclient "go-cloud-client"
)

type ExampleConf struct {
	ParentConf Conf `yaml:"conf"`
}

type Conf struct {
	Variable1 string `yaml:"var1"`
	Variable2 string `yaml:"var2"`
	Variable3 string `yaml:"var3"`
	Variable4 SubKey `yaml:"var4"`
}

type SubKey struct {
	NestedVar1 string `yaml:"sub_var_4"`
	NestedVar2 string `yaml:"sub_var_5"`
}

func main() {
	// Create a new client
	client, err := gocloudclient.NewClient()

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	var conf ExampleConf

	// Fetch configuration
	err = client.GetConfig(gocloudclient.YamlDecoder{}, &conf)
	if err != nil {
		log.Fatalf("Failed to fetch config: %v", err)
	}

	fmt.Printf("Conf Value: %v \n", conf.ParentConf)
	// Example with cached client
	fmt.Println("\n=== Using Cached Client ===")
	cachedClient := gocloudclient.NewCachedClient(client, 5*time.Minute)

	// First call - fetches from server
	err = cachedClient.GetConfig("myapp", "dev", "master",
		gocloudclient.YamlDecoder{}, &conf)
	if err != nil {
		log.Fatalf("Failed to fetch config: %v", err)
	}
	fmt.Printf("Config fetched (1st call): \n")

	// Second call - returns from cache
	err = cachedClient.GetConfig("myapp", "dev", "master",
		gocloudclient.YamlDecoder{}, &conf)
	if err != nil {
		log.Fatalf("Failed to fetch config: %v", err)
	}
	fmt.Printf("Config fetched (2nd call, from cache): \n")

	// Clear cache
	cachedClient.ClearCache()
	fmt.Println("Cache cleared")
}
