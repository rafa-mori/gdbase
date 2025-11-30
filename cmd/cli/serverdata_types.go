package cli

import (
	gl "github.com/kubex-ecosystem/logz"
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
		gl.Log("error", "No ports found")
		return WebSrvProcessServer{}, nil
	}
	if w.PubAddress == "" {
		gl.Log("error", "No public address found")
		return WebSrvProcessServer{}, nil
	}
	if w.AuthToken == "" {
		gl.Log("error", "No auth token found")
		return WebSrvProcessServer{}, nil
	}
	if w.Uptime == "" {
		gl.Log("error", "No uptime found")
		return WebSrvProcessServer{}, nil
	}
	if w.Processes == nil {
		gl.Log("error", "No processes found")
		return WebSrvProcessServer{}, nil
	}
	return WebSrvProcessServer{}, nil
}
