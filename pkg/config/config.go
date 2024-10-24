package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/KillianMeersman/chaperone/pkg/log"
)

// Get an environment variable as a string.
func GetString(name, defaultValue string, isSecret bool) string {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	return value
}

// Get an environment variable as a string. Panic if this variable is not defined
func MustGetString(name string, isSecret bool) string {
	value := os.Getenv(name)
	if value == "" {
		panic(fmt.Sprintf("environment variable '%s' is required but not present", name))
	}
	return value
}

// Get an environment variable as an int64.
func GetInt64(name string, defaultValue int64, isSecret bool) int64 {
	str := GetString(name, "", isSecret)
	if str == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(str, 10, 64)

	if err != nil {
		log.Fatal("invalid configuration value", "name", name)
	}

	return value
}

// Get an environment variable as an int64. Panics if this variable is not defined.
func MustGetInt64(name string, isSecret bool) int64 {
	str := MustGetString(name, isSecret)
	value, err := strconv.ParseInt(str, 10, 64)

	if err != nil {
		log.Fatal("invalid configuration value", "name", name)
	}

	return value
}

// Get an environment variable as an float64.
func GetFloat64(name string, defaultValue float64, isSecret bool) float64 {
	str := GetString(name, "", isSecret)
	if str == "" {
		return defaultValue
	}

	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.Fatal("invalid configuration value", "name", name)
	}

	return value
}

// Get an environment variable as an float64. Panics if this variable is not defined.
func MustGetFloat64(name string, isSecret bool) float64 {
	str := MustGetString(name, isSecret)

	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.Fatal("invalid configuration value", "name", name)
	}

	return value
}

// Get an environment variable as a bool.
func GetBool(name string, defaultValue bool, isSecret bool) bool {
	str := GetString(name, "", isSecret)
	if str == "" {
		return defaultValue
	}

	lowerStr := strings.ToLower(str)
	if lowerStr == "true" || lowerStr == "yes" || lowerStr == "1" {
		return true
	} else if lowerStr == "false" || lowerStr == "no" || lowerStr == "0" {
		return false
	}

	log.Fatal("invalid configuration value", "name", name)
	return false
}

// Get an environment variable as a bool. Panics if this variable is not defined.
func MustGetBool(name string, isSecret bool) bool {
	str := MustGetString(name, isSecret)

	lowerStr := strings.ToLower(str)
	if lowerStr == "true" || lowerStr == "yes" || lowerStr == "1" {
		return true
	} else if lowerStr == "false" || lowerStr == "no" || lowerStr == "0" {
		return false
	}

	log.Fatal("invalid configuration value", "name", name)
	return false
}

// Get an environment variable in the form of `a=1,b=2,c=3` as an map of strings.
func GetStringMap(name string, defaultValue map[string]string, isSecret bool) map[string]string {
	str := GetString(name, "", isSecret)

	parts := strings.Split(str, ",")
	strMap := make(map[string]string)

	for _, part := range parts {
		subParts := strings.SplitN(part, "=", 2)
		if len(subParts) < 2 {
			log.Fatal("invalid configuration value", "name", name)
		}
		strMap[subParts[0]] = subParts[1]
	}

	return strMap
}

// Get an environment variable in the form of `a=1,b=2,c=3` as an map of strings. Panics if this variable is not defined.
func MustGetStringMap(name string, isSecret bool) map[string]string {
	str := MustGetString(name, isSecret)

	parts := strings.Split(str, ",")
	strMap := make(map[string]string)

	for _, part := range parts {
		subParts := strings.SplitN(part, "=", 2)
		if len(subParts) < 2 {
			log.Fatal("invalid configuration value", "name", name)
		}
		strMap[subParts[0]] = subParts[1]
	}

	return strMap
}
