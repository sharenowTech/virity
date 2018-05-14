package pluginregistry

import (
	"fmt"
)

// Store the "new" functions of the
var scanner = make(map[string]func(config Config) Scan, 2)
var monitor = make(map[string]func(config Config) Monitor, 2)
var store = make(map[string]func(config Config) Store, 2)

// RegisterScanner saves the "new" function of a scanner plugin
func RegisterScanner(key string, s func(config Config) Scan) (string, error) {
	if _, ok := scanner[key]; ok {
		//log.Warn("Plugin " + key + " already assigned.")
		return key, fmt.Errorf("plugin has already been assigned")
	}
	scanner[key] = s
	return key, nil
}

// RegisterMonitor saves the "new" function of a monitor plugin
func RegisterMonitor(key string, m func(config Config) Monitor) (string, error) {
	if _, ok := monitor[key]; ok {
		//log.Warn("Plugin " + key + " already assigned.")
		return key, fmt.Errorf("plugin has already been assigned")
	}
	monitor[key] = m
	return key, nil
}

// RegisterStore saves the "new" function of a store plugin
func RegisterStore(key string, s func(config Config) Store) (string, error) {
	if _, ok := store[key]; ok {
		//log.Warn("Plugin " + key + " already assigned.")
		return key, fmt.Errorf("plugin has already been assigned")
	}
	store[key] = s
	return key, nil
}

// NewScanner creats a new scanner instance with provided plugin key
func NewScanner(key string, config Config) (Scan, error) {
	if _, ok := scanner[key]; !ok {
		//log.Warn("Plugin " + key + " already assigned.")
		return nil, fmt.Errorf("scanner %s not found", key)
	}
	newScanner := scanner[key]
	scanner := newScanner(config)
	return scanner, nil
}

// NewMonitor creates a new monitor instance with provided plugin key
func NewMonitor(key string, config Config) (Monitor, error) {
	if _, ok := monitor[key]; !ok {
		//log.Warn("Plugin " + key + " already assigned.")
		return nil, fmt.Errorf("monitor %s not found", key)
	}
	newMonitor := monitor[key]
	monitor := newMonitor(config)
	return monitor, nil
}

// NewStore creates a new store instance with provided plugin key
func NewStore(key string, config Config) (Store, error) {
	if _, ok := store[key]; !ok {
		//log.Warn("Plugin " + key + " already assigned.")
		return nil, fmt.Errorf("store %s not found", key)
	}
	newStore := store[key]
	store := newStore(config)
	return store, nil
}
