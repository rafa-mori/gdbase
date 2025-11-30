// Package broker provides functionality to manage brokers in the system.
package broker

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/goccy/go-json"
)

type BrokerManager struct{}

func NewBrokerManager() *BrokerManager { return &BrokerManager{} }

func (bm *BrokerManager) GetBrokers() []BrokerInfoLock {
	registryFile := "/tmp/gkbxsrv_registry.json"
	var brokers []BrokerInfoLock
	data, _ := os.ReadFile(registryFile)

	if unmarshalErr := json.Unmarshal(data, &brokers); unmarshalErr != nil {
		brokers = []BrokerInfoLock{}
	}
	return brokers
}
func (bm *BrokerManager) loadBrokerInfo(configDir string) ([]BrokerInfoLock, error) {
	var brokers []BrokerInfoLock
	err := filepath.Walk(configDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			var broker BrokerInfo
			unmarshalErr := json.Unmarshal(data, &broker)
			if unmarshalErr != nil {
				return unmarshalErr
			}

			brokers = append(brokers, BrokerInfoLock{
				Name:  broker.Name,
				Port:  broker.Port,
				PID:   broker.PID,
				Time:  broker.Time,
				path:  path,
				flock: sync.Mutex{},
			})
		}
		return nil
	})
	return brokers, err
}
