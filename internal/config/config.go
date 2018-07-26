package config

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

var Config string

type monitorConfig struct {
	Type     string `json:"type"`
	Endpoint string `json:"endpoint"`
	//Username        string `json:"username"`
	//Password        string `json:"password"`
	DefaultAssignee string `json:"default-assignee"`
	CreateTickets   bool   `json:"create-tickets"`
}

type generalConfig struct {
	LogLevel      string        `json:"loglevel"`
	LogType       string        `json:"logtype"`
	AgentLifetime time.Duration `json:"agent-lifetime"`
	AgentSeed     int64         `json:"agent-seed"`
	AgentEnv      string        `json:"agent-env"`
}

type scannerConfig struct {
	Type          string `json:"type"`
	Endpoint      string `json:"endpoint"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	SeverityLevel int    `json:"severity-level"`
}

type storeConfig struct {
	Type               string        `json:"type"`
	Endpoint           string        `json:"endpoint"`
	IntervalAgentPush  time.Duration `json:"agentPush"`
	IntervalServerPull time.Duration `json:"serverPoll"`
}

type Duration struct {
	time.Duration
}

func (self *Duration) UnmarshalJSON(b []byte) (err error) {
	s := string(b)

	// Get rid of the quotes "" around the value.
	// A second option would be to include them
	s = s[1 : len(s)-1]

	t, err := time.ParseDuration(s)

	self.Duration = t
	return
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
		IntervalAgentPush:  viper.GetDuration("store.agentPush"),
		IntervalServerPull: viper.GetDuration("store.serverPoll"),
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

func GetMonitorConfig() []monitorConfig {
	data := viper.Get("monitor")
	list := data.([]interface{})

	configList := make([]monitorConfig, len(list))

	for index, _ := range list {
		marshall(list[index], &configList[index])
	}

	fmt.Println(configList)
	return configList

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

func marshall(v interface{}, obj interface{}) {

	elemMap := cast.ToStringMap(v)

	bytes, err := json.Marshal(&elemMap)
	if err != nil {
		panic(fmt.Sprintf("Could not parse config: %v", err))
	}

	err = json.Unmarshal(bytes, &obj)
	if err != nil {
		panic(fmt.Sprintf("Could not parse config: %v", err))
	}
}
