// Package interfaces provides the interfaces for the GoBE application
package interfaces

import (
	"net/http"

	"github.com/kubex-ecosystem/logz"
)

type IGoBE interface {
	StartGoBE()
	HandleValidate(w http.ResponseWriter, r *http.Request)
	HandleContact(w http.ResponseWriter, r *http.Request)
	RateLimit(w http.ResponseWriter, r *http.Request) bool
	Initialize() error
	GetLogFilePath() string
	GetConfigFilePath() string
	GetLogger() *logz.LoggerZ
	Mu() IMutexes
	GetReference() IReference
	Environment() IEnvironment
}
