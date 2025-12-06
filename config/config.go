package config

import (
	"os"
	"strconv"
	"strings"
)

func New() map[string]string {
	environ := os.Environ()
	envAsMap := make(map[string]string, len(environ))
	for _, entry := range environ {
		if entry != "" {
			key, value := split(entry)
			envAsMap[key] = value
		}
	}
	return envAsMap
}

// assumes entry is not the empty string
func split(entry string) (key, value string) {
	parts := strings.SplitN(entry, "=", 2)
	if len(parts) < 2 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

func GetString(config map[string]string, key string, defaultValue string) string {
	if config == nil {
		return defaultValue
	}

	if val, ok := config[key]; ok {
		return val
	}
	return defaultValue
}

func GetInt(config map[string]string, key string, defaultValue int) int {
	if config == nil {
		return defaultValue
	}

	s, ok := config[key]
	if !ok {
		return defaultValue
	}

	asInt, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}

	return asInt
}
