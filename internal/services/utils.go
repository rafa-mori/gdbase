// Package services contains utility functions for managing database services.
package services

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/go-connections/nat"

	u "github.com/kubex-ecosystem/gdbase/utils"
	logz "github.com/kubex-ecosystem/logz"
)

func SlitMessage(recPayload []string) (id, msg []string) {
	if recPayload[1] == "" {
		id = recPayload[:2]
		msg = recPayload[2:]
	} else {
		id = recPayload[:1]
		msg = recPayload[1:]
	}
	return
}

func RndomName() string {
	return "broker-" + RandStringBytes(5)
}
func RandStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func FindAvailablePort(basePort int, maxAttempts int) (string, error) {
	for z := 0; z < maxAttempts; z++ {
		if basePort+z < 1024 || basePort+z > 49151 {
			continue
		}
		port := fmt.Sprintf("%d", basePort+z)
		isOpen, err := u.CheckPortOpen(port)
		if err != nil {
			return "", fmt.Errorf("error checking port %s: %w", port, err)
		}
		if !isOpen {
			logz.Log("warn", fmt.Sprintf("⚠️ Port %s is occupied, trying the next one...\n", port))
			continue
		}
		logz.Log("info", fmt.Sprintf("✅ Available port found: %s\n", port))
		return port, nil
	}
	return "", fmt.Errorf("no available port in range %d-%d", basePort, basePort+maxAttempts-1)
}

func WriteInitDBSQL(initVolumePath, initDBSQL, initDBSQLData string) (string, error) {
	if err := os.MkdirAll(initVolumePath, 0755); err != nil {
		logz.Log("error", fmt.Sprintf("Error creating directory: %v", err))
		return "", err
	}
	filePath := filepath.Join(initVolumePath, initDBSQL)
	if _, err := os.Stat(filePath); err == nil {
		logz.Log("debug", fmt.Sprintf("File %s already exists, skipping creation", filePath))
		return filePath, nil
	}
	if err := os.WriteFile(filePath, []byte(initDBSQLData), 0644); err != nil {
		logz.Log("error", fmt.Sprintf("Error writing file: %v", err))
		return "", err
	}
	logz.Log("info", fmt.Sprintf("✅ File %s created successfully!\n", filePath))
	return filePath, nil
}

func ExtractPort(port nat.PortMap) any {
	// Verifica se a porta é válida
	if port == nil {
		return nil
	}
	// Extrai a porta e o protocolo do primeiro elemento do map
	for k := range port {
		portStr := strings.Split(string(k), "/")
		if len(portStr) != 2 {
			return nil
		}
		portNum := portStr[0]
		protocol := portStr[1]
		return map[string]string{
			"port":     portNum,
			"protocol": protocol,
		}
	}
	return nil
}
