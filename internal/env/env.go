package env

import (
	"os"
	"strconv"
)

const (
	ServerBaseDns   = "SERVER_BASE_DNS"
	WebProxyBaseDns = "WEB_BASE_DNS"
	ServerSshDns    = "SERVER_SSH_DNS"
	ServerSshPort   = "SERVER_SSH_PORT"
	PrivateSshKey   = "PRIVATE_SSH_KEY"
)

func GetOrDefault(envKey, defaultVal string) string {
	val := os.Getenv(envKey)
	if val == "" {
		return defaultVal
	}
	return os.Getenv(envKey)
}

func Get(envKey string) string {
	return os.Getenv(envKey)
}

func GetEnvWithDefaultAsInt(envKey string, defaultVal int) int {
	val := os.Getenv(envKey)
	if val == "" {
		return defaultVal
	}
	intValue, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intValue
}
