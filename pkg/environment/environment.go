package environment

import "os"

// GetEnvOrDefault returns the value of the environment variable with the given key if it exists, otherwise it returns the given default value.
func GetEnvOrDefault(key string, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}
