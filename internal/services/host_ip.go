package services

import (
	"os"
	"strings"
)

const hostIPEnv = "CANALIZE_HOST_IPV4"

// resolveHostIP retorna o IP configurado via env ou usa 0.0.0.0 como padrão.
func resolveHostIP() string {
	if v := strings.TrimSpace(os.Getenv(hostIPEnv)); v != "" {
		return v
	}
	return "0.0.0.0"
}
