package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var Config string

type monitorConfig struct {
	Type            string
	Endpoint        string
	Username        string
	Password        string
	DefaultAssignee string
	CreateTickets   bool
}

type generalConfig struct {
	LogLevel      string
	LogType       string
	AgentLifetime time.Duration
	AgentSeed     int64
	AgentEnv      string
}

type scannerConfig struct {
	Type          string
	Endpoint      string
	Username      string
	Password      string
	SeverityLevel int
}

type storeConfig struct {
	Type               string
	Endpoint           string
	IntervalAgentPush  time.Duration
	IntervalServerPull time.Duration
}

func init() {
	/*if Config != "" {
		viper.SetConfigFile(Config)
	}*/

	viper.SetConfigName("config")        // name of config file (without extension)
	viper.AddConfigPath("/etc/virity/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.virity") // call multiple times to add many search paths
	viper.AddConfigPath(".")             // optionally look for config in the working directory
	viper.AddConfigPath("$GOPATH/src/github.com/car2go/virity")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		fmt.Println("No config file found. Using default configuration")
	}

	viper.SetEnvPrefix("virity")
	viper.AutomaticEnv()

	replacer := strings.NewReplacer(".", "_", "-", "")
	viper.SetEnvKeyReplacer(replacer)

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:" + e.Name)
	})
}

func GetStoreConfig() storeConfig {
	return storeConfig{
		Endpoint:           viper.GetString("store.endpoint"),
		IntervalAgentPush:  viper.GetDuration("store.interval.agentPush"),
		IntervalServerPull: viper.GetDuration("store.interval.serverPoll"),
		Type:               viper.GetString("store.type"),
	}
}

func GetScanConfig() scannerConfig {
	return scannerConfig{
		Type:          viper.GetString("scanner.type"),
		Endpoint:      viper.GetString("scanner.endpoint"),
		Username:      viper.GetString("scanner.username"),
		Password:      viper.GetString("scanner.password"),
		SeverityLevel: viper.GetInt("scanner.severity-level"),
	}
}

func GetMonitorConfig() monitorConfig {
	return monitorConfig{
		Type:            viper.GetString("monitor.type"),
		Endpoint:        viper.GetString("monitor.endpoint"),
		Username:        viper.GetString("monitor.username"),
		Password:        viper.GetString("monitor.password"),
		DefaultAssignee: viper.GetString("monitor.default-assignee"),
		CreateTickets:   viper.GetBool("monitor.create-tickets"),
	}
}

func GetGeneralConfig() generalConfig {
	return generalConfig{
		LogLevel:      viper.GetString("general.loglevel"),
		LogType:       viper.GetString("general.logtype"),
		AgentLifetime: viper.GetDuration("general.agent-lifetime"),
		AgentSeed:     viper.GetInt64("general.agent-seed"),
		AgentEnv:      viper.GetString("general.agent-env"),
	}
}
