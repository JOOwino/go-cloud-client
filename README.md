# Go Cloud Client

A Go client library for fetching configurations from Spring Cloud Config Server.

## Features

- ✅ Fetch configuration from Spring Cloud Config Server
- ✅ Support for multiple profiles and labels (branches)
- ✅ Basic authentication support
- ✅ Type-safe value retrieval (String, Int, Bool)
- ✅ Configuration caching with TTL
- ✅ YAML and JSON format support
- ✅ Custom HTTP client support

## Installation

```bash
go get go-cloud-client
```

Or clone the repository:

```bash
git clone <repository-url>
cd go-cloud-client
go mod download
```

## Usage

### Basic Example

```go
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
        Username: "user",  // Optional
        Password: "pass",  // Optional
        Timeout:  30 * time.Second,
    })
    if err != nil {
        log.Fatal(err)
    }

    // Fetch configuration
    config, err := client.GetConfig("myapp", "dev", "master")
    if err != nil {
        log.Fatal(err)
    }

    // Get configuration values
    dbHost, _ := config.GetString("spring.datasource.host")
    dbPort, _ := config.GetInt("spring.datasource.port")
    debugMode, _ := config.GetBool("app.debug")

    fmt.Printf("DB Host: %s\n", dbHost)
    fmt.Printf("DB Port: %d\n", dbPort)
    fmt.Printf("Debug: %v\n", debugMode)
}
```

### With Caching

```go
// Create a cached client with 5-minute TTL
cachedClient := gocloudclient.NewCachedClient(client, 5*time.Minute)

// First call - fetches from server
config1, _ := cachedClient.GetConfig("myapp", "dev", "master")

// Second call - returns from cache (if within TTL)
config2, _ := cachedClient.GetConfig("myapp", "dev", "master")

// Clear cache manually
cachedClient.ClearCache()

// Invalidate specific entry
cachedClient.InvalidateCache("myapp", "dev", "master")
```

### Multiple Profiles

```go
// Fetch config with multiple profiles (comma-separated)
config, err := client.GetConfig("myapp", "dev,common", "master")
```

### Custom Label/Branch

```go
// Fetch from a specific branch
config, err := client.GetConfig("myapp", "prod", "release-v1.0")
```

### Get Raw YAML

```go
// Fetch configuration as raw YAML
yamlConfig, err := client.GetConfigYAML("myapp", "dev", "master")
fmt.Println(yamlConfig)
```

### Convert to Different Formats

```go
// Convert config response to YAML
yamlStr, err := config.ToYAML()

// Convert config response to JSON
jsonStr, err := config.ToJSON()
```

### Access All Properties

```go
// Get all properties as a map
allProperties := config.GetPropertySources()

// Iterate through properties
for key, value := range allProperties {
    fmt.Printf("%s = %v\n", key, value)
}
```

## API Reference

### ClientConfig

Configuration for creating a new client:

```go
type ClientConfig struct {
    BaseURL    string        // Required: Base URL of Spring Cloud Config Server
    Username   string        // Optional: Username for basic auth
    Password   string        // Optional: Password for basic auth
    Timeout    time.Duration // Optional: HTTP client timeout (default: 30s)
    HTTPClient *http.Client  // Optional: Custom HTTP client
}
```

### Client Methods

#### `GetConfig(application, profile, label string) (*ConfigResponse, error)`

Fetches configuration from Spring Cloud Config Server.

- `application`: Application name (e.g., "myapp")
- `profile`: Profile(s) - can be comma-separated (e.g., "dev", "dev,common")
- `label`: Label/branch (e.g., "master", "develop") - defaults to "master" if empty

Returns a `ConfigResponse` object.

#### `GetConfigYAML(application, profile, label string) (string, error)`

Fetches configuration in raw YAML format.

### ConfigResponse Methods

#### `GetValue(key string) (interface{}, bool)`

Returns the raw value for a given key.

#### `GetString(key string) (string, bool)`

Returns the value as a string.

#### `GetInt(key string) (int, bool)`

Returns the value as an int.

#### `GetBool(key string) (bool, bool)`

Returns the value as a bool.

#### `GetPropertySources() map[string]interface{}`

Returns all properties flattened into a single map (later sources override earlier ones).

#### `ToYAML() (string, error)`

Converts the configuration to YAML format.

#### `ToJSON() (string, error)`

Converts the configuration to JSON format.

### CachedClient

A wrapper around `Client` that adds caching capabilities:

```go
cachedClient := gocloudclient.NewCachedClient(client, 5*time.Minute)
```

#### Methods

- `GetConfig(application, profile, label string) (*ConfigResponse, error)` - Get config with caching
- `ClearCache()` - Clear all cached entries
- `InvalidateCache(application, profile, label string)` - Invalidate specific cache entry

## Spring Cloud Config Server Endpoints

The client uses the following Spring Cloud Config Server endpoints:

- **JSON format**: `GET {baseURL}/{application}/{profile}/{label}`
- **YAML format**: `GET {baseURL}/{application}-{profile}.yml` or `{baseURL}/{label}/{application}-{profile}.yml`

## Error Handling

The client returns errors for:
- Invalid configuration (missing baseURL, etc.)
- Network errors
- HTTP errors (non-200 status codes)
- JSON/YAML parsing errors

Always check errors:

```go
config, err := client.GetConfig("myapp", "dev", "master")
if err != nil {
    log.Printf("Error fetching config: %v", err)
    return
}
```

## Examples

See the `example/` directory for more complete examples.

## Testing

Run tests:

```bash
go test ./...
```

## License

MIT

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

