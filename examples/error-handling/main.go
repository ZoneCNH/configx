package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/ZoneCNH/configx/pkg/configx"
)

// appConfig demonstrates struct tags: required fields, defaults, and secret fields.
type appConfig struct {
	Name    string               `config:"APP_NAME" required:"true"`
	Port    int                  `config:"PORT" default:"8080"`
	Timeout time.Duration        `config:"TIMEOUT" default:"30s"`
	DBPass  configx.SecretString `config:"DB_PASSWORD"`
	Debug   bool                 `config:"DEBUG" default:"false"`
}

// Validate implements the configx.Validator interface for post-decode checks.
func (c appConfig) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got %d", c.Port)
	}
	return nil
}

func main() {
	ctx := context.Background()

	// --- 1. Missing config file ---
	fmt.Println("=== Missing Config File ===")
	missingFileErr(ctx)

	// --- 2. Invalid format ---
	fmt.Println("\n=== Invalid Format ===")
	invalidFormatErr(ctx)

	// --- 3. Multiple sources with merge priority (env overrides file) ---
	fmt.Println("\n=== Merge Priority ===")
	mergePriority(ctx)

	// --- 4. SecretString redaction ---
	fmt.Println("\n=== Secret Redaction ===")
	secretRedaction(ctx)

	// --- 5. Validation error ---
	fmt.Println("\n=== Validation Error ===")
	validationError(ctx)
}

// missingFileErr shows how configx reports a missing file error.
func missingFileErr(ctx context.Context) {
	_, err := configx.LoadJSONFile(ctx, "/nonexistent/config.json")
	if err == nil {
		fmt.Println("ERROR: expected an error for missing file")
		return
	}

	var cfgErr *configx.Error
	if errors.As(err, &cfgErr) {
		fmt.Printf("kind=%s retryable=%v\n", cfgErr.Kind, cfgErr.Retryable)
	} else {
		fmt.Println("non-configx error:", err)
	}
}

// invalidFormatErr shows the error when a file has invalid JSON.
func invalidFormatErr(ctx context.Context) {
	// Write a malformed JSON file to a temp location.
	tmpFile, err := os.CreateTemp("", "configx-bad-*.json")
	if err != nil {
		fmt.Printf("ERROR creating temp file: %v\n", err)
		return
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(`{"key": "value", broken`); err != nil {
		fmt.Printf("ERROR writing temp file: %v\n", err)
		return
	}
	tmpFile.Close()

	_, err = configx.LoadJSONFile(ctx, tmpFile.Name())
	if err == nil {
		fmt.Println("ERROR: expected an error for invalid JSON")
		return
	}

	var cfgErr *configx.Error
	if errors.As(err, &cfgErr) {
		fmt.Printf("kind=%s\n", cfgErr.Kind)
	} else {
		fmt.Println("non-configx error:", err)
	}
}

// mergePriority demonstrates that later sources override earlier ones (LastWins).
func mergePriority(ctx context.Context) {
	envKey := "APP_PORT"
	original := os.Getenv(envKey)
	defer os.Setenv(envKey, original)

	os.Setenv(envKey, "9090")

	loader := configx.NewLoader().
		AddSource(configx.NewMapSource("defaults", map[string]string{
			"APP_NAME": "from-file",
			"PORT":     "3000",
		})).
		AddSource(configx.NewAllEnvSource("APP_"))

	result, err := loader.Load(ctx)
	if err != nil {
		fmt.Printf("ERROR loading: %v\n", err)
		return
	}

	// NewAllEnvSource strips the prefix, so APP_PORT becomes PORT.
	port, ok := result.Get("PORT")
	if !ok {
		fmt.Println("ERROR: expected PORT in result")
		return
	}

	name, _ := result.Get("APP_NAME")
	fmt.Printf("APP_NAME=%s PORT=%s (env overrode map)\n", name, port)
}

// secretRedaction shows SecretString values are redacted in sanitized output.
func secretRedaction(ctx context.Context) {
	loader := configx.NewLoader().
		AddSource(configx.NewSecretMapSource("secrets", map[string]string{
			"APP_NAME":    "myapp",
			"DB_PASSWORD": "super-secret-123",
		}, []string{"DB_PASSWORD"}))

	result, err := loader.Load(ctx)
	if err != nil {
		fmt.Printf("ERROR loading: %v\n", err)
		return
	}

	sanitized := result.Sanitize()
	fmt.Printf("DB_PASSWORD raw:       %q\n", result.Values["DB_PASSWORD"].Value)
	fmt.Printf("DB_PASSWORD sanitized: %q\n", sanitized.Values["DB_PASSWORD"].Value)
}

// validationError demonstrates required-field validation via Decode.
func validationError(ctx context.Context) {
	// Load a result that is missing the required APP_NAME key.
	loader := configx.NewLoader().
		AddSource(configx.NewMapSource("partial", map[string]string{
			"PORT": "8080",
		}))

	result, err := loader.Load(ctx)
	if err != nil {
		fmt.Printf("ERROR loading: %v\n", err)
		return
	}

	var cfg appConfig
	err = configx.Decode(result, &cfg)
	if err == nil {
		fmt.Println("ERROR: expected validation error for missing APP_NAME")
		return
	}

	if configx.IsKind(err, configx.ErrorKindValidation) {
		fmt.Println("caught validation error as expected")
	} else {
		fmt.Printf("unexpected error kind: %v\n", err)
	}
}
