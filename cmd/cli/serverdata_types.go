package cli

import (
	logz "github.com/kubex-ecosystem/logz"
)

type IWebSrvServerData interface {
	GetWebSrvServerData() (WebSrvProcessServer, error)
}
type WebSrvServerData struct {
	Ports      []string       `json:"ports"`
	PubAddress string         `json:"pubAddress"`
	AuthToken  string         `json:"authToken"`
	Uptime     string         `json:"uptime"`
	Processes  []WebSrvServer `json:"processes"`
}

func NewWebSrvServerData(ports []string, pubAddress, authToken, uptime string, processes []WebSrvServer) IWebSrvServerData {
	return &WebSrvServerData{
		Ports:      ports,
		PubAddress: pubAddress,
		AuthToken:  authToken,
		Uptime:     uptime,
		Processes:  processes,
	}
}

func (w *WebSrvServerData) GetWebSrvServerData() (WebSrvProcessServer, error) {
	if w.Ports == nil {
		logz.Log("error", "No ports found")
		return WebSrvProcessServer{}, nil
	}
	if w.PubAddress == "" {
		logz.Log("error", "No public address found")
		return WebSrvProcessServer{}, nil
	}
	if w.AuthToken == "" {
		logz.Log("error", "No auth token found")
		return WebSrvProcessServer{}, nil
	}
	if w.Uptime == "" {
		logz.Log("error", "No uptime found")
		return WebSrvProcessServer{}, nil
	}
	if w.Processes == nil {
		logz.Log("error", "No processes found")
		return WebSrvProcessServer{}, nil
	}
	return WebSrvProcessServer{}, nil
}
