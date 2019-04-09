package env

import "os"

// GetEnvString returns an environment variable as a string if it exists. Otherwise
// it returns the default value.
func GetEnvString(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return def
}
