package main

import (
	"github.com/car2go/virity/internal/config"
	"github.com/car2go/virity/internal/pluginregistry"
)

func createMonitor() ([]pluginregistry.Monitor, error) {
	configMonitor := config.GetMonitorConfig()
	config := make([]pluginregistry.Config, len(configMonitor))
	for index, val := range configMonitor {
		config[index] = pluginregistry.Config{
			PluginID: val.Type,
			Endpoint: val.Endpoint,
			//User:            configMonitor.Username,
			//Password:        configMonitor.Password,
			DefaultAssignee: val.DefaultAssignee,
			CreateTickets:   val.CreateTickets,
		}
	}
	monitor, err := pluginregistry.NewMonitor(config)
	if err != nil {
		return nil, err
	}

	return monitor, nil
}

func createScanner() (pluginregistry.Scan, error) {
	configScan := config.GetScanConfig()
	scan, err := pluginregistry.NewScanner(configScan.Type, pluginregistry.Config{
		Endpoint: configScan.Endpoint,
		User:     configScan.Username,
		Password: configScan.Password,
	})
	if err != nil {
		return nil, err
	}

	return scan, nil
}

func createStore() (pluginregistry.Store, error) {
	configStore := config.GetStoreConfig()
	store, err := pluginregistry.NewStore(configStore.Type, pluginregistry.Config{
		Endpoint: configStore.Endpoint,
	})
	if err != nil {
		return nil, err
	}

	return store, nil
}
