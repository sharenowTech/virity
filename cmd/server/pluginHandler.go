package main

import (
	"github.com/sharenowTech/virity/internal/config"
	"github.com/sharenowTech/virity/internal/pluginregistry"
)

func createMonitor() (pluginregistry.Monitor, error) {
	configMonitor := config.GetMonitorConfig()
	monitor, err := pluginregistry.NewMonitor(configMonitor.Type, pluginregistry.Config{
		Endpoint:        configMonitor.Endpoint,
		User:            configMonitor.Username,
		Password:        configMonitor.Password,
		DefaultAssignee: configMonitor.DefaultAssignee,
		CreateTickets:   configMonitor.CreateTickets,
	})
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
