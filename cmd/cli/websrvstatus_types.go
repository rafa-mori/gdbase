package cli

import (
	"fmt"

	logz "github.com/kubex-ecosystem/logz"
)

type IWebSrvServerStatus interface {
	FollowWebSrvServerStatus(interval int, follow bool, logs bool) error
	NewWebSrvServerProcessData() ([]IWebSrvProcessServer, error)
}
type WebSrvServerStatus struct {
	Interval int
	Follow   bool
	Logs     bool
}

func (wss *WebSrvServerProcessData) FollowWebSrvServerStatus(interval int, follow bool, logs bool) error {
	if interval < 1 {
		logz.Log("error", "Interval must be greater than 0")
		return nil
	}

	process := NewWebSrvServerProcessData()
	procs, err := process.GetWevSrvServerProcessData()
	if err != nil {
		return err
	}
	var processes []IWebSrvServerData
	for _, proc := range procs {
		srvData, srvDataErr := proc.GetServerData()
		if srvDataErr != nil {
			return srvDataErr
		}
		prc := NewWebSrvServerData(proc.Ports, proc.GetPubAddress(), srvData.GetAuthToken(), srvData.GetUptime(), srvData.GetProcesses())
		processes = append(processes, prc)
	}

	if follow {
		for _, prc := range processes {
			if logs {
				prcData, prcDataErr := prc.GetWebSrvServerData()
				if prcDataErr != nil {
					return prcDataErr
				}
				logz.Log("info", fmt.Sprintf("Web Server Process Data - Ports: %v, Pub Address: %v, Auth Token: %v, Uptime: %v, Processes: %v", prcData.Ports, prcData.PubAddress, prcData.AuthToken, prcData.Uptime, prcData.Processes))
			}
		}
	}

	return nil
}
func (wss *WebSrvServerProcessData) NewWebSrvServerProcessData() ([]IWebSrvProcessServer, error) {
	return []IWebSrvProcessServer{}, nil
}
