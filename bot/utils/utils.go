package utils

import (
	"fmt"
	"os"
)

func GetenvOrPanic(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("missing environment variable %s", key))
	}

	return value
}

func GetenvOrError(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("missing environment variable %s", key)
	}

	return value, nil
}
