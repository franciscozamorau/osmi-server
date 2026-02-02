package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Environment maneja variables de entorno
type Environment struct{}

// NewEnvironment crea una nueva instancia
func NewEnvironment() *Environment {
	return &Environment{}
}

// Get obtiene variable de entorno con default
func (e *Environment) Get(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetRequired obtiene variable de entorno requerida
func (e *Environment) GetRequired(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", &EnvError{Key: key, Reason: "variable de entorno requerida no configurada"}
	}
	return value, nil
}

// GetInt obtiene variable como int
func (e *Environment) GetInt(key string, defaultValue int) int {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(strValue)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetBool obtiene variable como bool
func (e *Environment) GetBool(key string, defaultValue bool) bool {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(strValue)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetDuration obtiene variable como duration
func (e *Environment) GetDuration(key string, defaultValue time.Duration) time.Duration {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(strValue)
	if err != nil {
		return defaultValue
	}
	return value
}

// GetStringSlice obtiene variable como slice de strings
func (e *Environment) GetStringSlice(key string, separator string, defaultValue []string) []string {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue
	}

	return strings.Split(strValue, separator)
}

// GetFloat obtiene variable como float64
func (e *Environment) GetFloat(key string, defaultValue float64) float64 {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue
	}

	value, err := strconv.ParseFloat(strValue, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

// IsSet verifica si variable está configurada
func (e *Environment) IsSet(key string) bool {
	return os.Getenv(key) != ""
}

// EnvError error de variable de entorno
type EnvError struct {
	Key    string
	Reason string
}

func (e *EnvError) Error() string {
	return fmt.Sprintf("environment variable error [%s]: %s", e.Key, e.Reason)
}

// Validators
func (e *Environment) ValidatePort(key string) (int, error) {
	port := e.GetInt(key, 0)
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("invalid port %d for key %s", port, key)
	}
	return port, nil
}

func (e *Environment) ValidatePositiveInt(key string, defaultValue int) (int, error) {
	value := e.GetInt(key, defaultValue)
	if value <= 0 {
		return 0, fmt.Errorf("value must be positive for key %s", key)
	}
	return value, nil
}

// LoadFromFile carga variables desde archivo
func (e *Environment) LoadFromFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read env file: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}

	return nil
}
