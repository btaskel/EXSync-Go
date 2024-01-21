package config

import (
	"EXSync/core/internal/modules/hashext"
	"EXSync/core/option"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const configPath = ".\\data\\config"

var Config = initConfig()

// initConfig 启动时进行初始化读取
func initConfig() *option.ConfigStruct {
	config, err := LoadConfig()
	if err != nil {
		return nil
	}
	return config
}

// LoadConfig 加载Config文件
func LoadConfig() (result *option.ConfigStruct, err error) {
	err = CreateConfig()
	if err != nil {
		return nil, err
	}

	file, err := os.ReadFile(filepath.Join(configPath, "config.json"))
	if err != nil {
		return
	}
	var config option.ConfigStruct
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	// log-logLevel
	if config.Log.LogLevel == "" {
		config.Log.LogLevel = "info"
	}

	switch strings.ToLower(config.Log.LogLevel) {
	case "debug":
		config.Log.LogLevel = "debug"
	case "info":
		config.Log.LogLevel = "info"
	case "warning":
		config.Log.LogLevel = "warning"
	case "error":
		config.Log.LogLevel = "error"
	case "none":
		config.Log.LogLevel = "none"
	default:
		config.Log.LogLevel = "info"
	}

	// server-addr-id
	if config.Server.Addr.ID != "" {
		if len(config.Server.Addr.ID) < 4 {
			logrus.Warning("If the device ID length is less than 4, there may be a security risk.")
		}
	} else {
		logrus.Info("The device ID is already random, which will hide your device.")
	}

	// server-addr-ip
	ip := net.ParseIP(config.Server.Addr.IP)
	if ip == nil {
		logrus.Error("The host IP address is not ipv4 or ipv6!")
		logrus.Warning("The host IP address is not filled in and has been defaulted to 0.0.0.0!")
		config.Server.Addr.IP = "0.0.0.0"
	} else {
		if ip.To4() != nil {
			logrus.Debug("A valid IPv4 address")
		} else if ip.To16() != nil {
			logrus.Debug("A valid IPv6 address")
		}
	}

	// server-addr-port
	if config.Server.Addr.Port < 1024 && config.Server.Addr.Port > 65535 {
		logrus.Warning("Port number setting error! Has been defaulted to 5001!")
		config.Server.Addr.Port = 5001
	}

	// server-addr-password
	if config.Server.Addr.Password == "" {
		logrus.Warning("The host has not set a password, and RSA&AES will be used for encryption during transmission")
	} else if len(config.Server.Addr.Password) < 4 {
		logrus.Error("Password length is less than 4! Should be between 4 and 48 characters!")
		os.Exit(1)
	} else if len(config.Server.Addr.Password) > 48 {
		logrus.Error("The password length is greater than 48! Should be between 4 and 48 characters!")
		os.Exit(1)
	}

	// server-setting-IOBalance
	//if config.Server.Setting.IOBalance {
	//
	//}

	// server-scan-enabled

	// server-scan-type
	switch strings.ToLower(config.Server.Scan.Type) {
	case "lan":
		config.Server.Scan.Type = "lan"
	case "white":
		config.Server.Scan.Type = "white"
	case "black":
		config.Server.Scan.Type = "black"
	default:
		logrus.Warning("Scan: No device discovery mode specified, default to LAN mode.")
		config.Server.Scan.Type = "lan"
	}

	// server-scan-max
	if config.Server.Scan.Max < 1 {
		logrus.Error("The maximum number of devices cannot be less than 1!")
		os.Exit(1)
	}

	// server-scan-device
	if config.Server.Scan.Type != "lan" && len(config.Server.Scan.Devices) == 0 {
		logrus.Errorf("Scan mode is %s, but device list is empty", config.Server.Scan.Type)
	}

	// server-proxy-enabled
	if config.Server.Proxy.Enabled {
		// server-proxy-hostname

		if net.ParseIP(config.Server.Proxy.Hostname) == nil {
			logrus.Errorf("Invalid proxy server IP: %s", config.Server.Proxy.Hostname)
		}

		// server-proxy-port
		if config.Server.Proxy.Port < 1024 && config.Server.Proxy.Port > 65536 {
			logrus.Error("Proxy: Port number setting error!")
		}
		// server-proxy-username
		if config.Server.Proxy.Username != "" {
			// server-proxy-password
			if config.Server.Proxy.Password == "" {
				logrus.Error("Missing proxy server username and password!")
			}
		}

		// userdata
		count := 0
		userdataList := make([]string, 8)
		for _, userdata := range config.Userdata {
			if userdata.Spacename == "" {
				logrus.Errorf("The %v th sync space is named empty! This space will not start!", count)
			} else {
				for _, s := range userdataList {
					if s == userdata.Spacename {
						logrus.Errorf("Duplicate naming of synchronization space %v!", s)
						os.Exit(1)
					}
				}
			}
			spacenameLength := len(userdata.Spacename)
			if spacenameLength > 20 && spacenameLength < 2 {
				logrus.Warningf("The length of the synchronization space %s name should be between 2 and 20 characters!", userdata.Spacename)
			}
			userdataList = append(userdataList, userdata.Spacename)
			if _, err = os.Stat(userdata.Path); err != nil {
				if os.IsNotExist(err) {
					logrus.Errorf("The sync space path named %s is invalid, it will not work!", userdata.Spacename)
					os.Exit(1)
				}
			}
			if userdata.Interval < 1 {
				config.Userdata[count].Interval = 30
				logrus.Warningf("The time interval setting for %s is incorrect and has been reset to 30 seconds!", userdata.Spacename)
			}
			count += 1
		}

	}

	return &config, nil
}

// CreateConfig 创建配置文件
func CreateConfig() (err error) {
	config := option.ConfigStruct{
		Log: struct {
			LogLevel string `json:"loglevel"`
		}{
			LogLevel: "info",
		},
		Server: struct {
			Addr struct {
				ID       string `json:"id"`
				IP       string `json:"ip"`
				Port     int    `json:"port"`
				Password string `json:"password"`
			} `json:"addr"`
			Setting struct {
				Encode     string `json:"encode"`
				IOBalance  bool   `json:"iobalance"`
				Encryption string `json:"encryption"`
			} `json:"setting"`
			Scan struct {
				Enabled bool     `json:"enabled"`
				Type    string   `json:"type"`
				Max     int      `json:"max"`
				Devices []string `json:"devices"`
			} `json:"scan"`
			Plugin struct {
				Enabled   bool     `json:"enabled"`
				Blacklist []string `json:"blacklist"`
			} `json:"plugin"`
			Proxy struct {
				Enabled  bool   `json:"enabled"`
				Hostname string `json:"hostname"`
				Port     int    `json:"port"`
				Username string `json:"username"`
				Password string `json:"password"`
			} `json:"proxy"`
		}{
			Addr: struct {
				ID       string `json:"id"`
				IP       string `json:"ip"`
				Port     int    `json:"port"`
				Password string `json:"password"`
			}{
				ID:       hashext.GetRandomStr(8),
				IP:       "127.0.0.1",
				Port:     5002,
				Password: hashext.GetRandomStr(10),
			},
			Setting: struct {
				Encode     string `json:"encode"`
				IOBalance  bool   `json:"iobalance"`
				Encryption string `json:"encryption"`
			}{
				Encode:     "utf-8",
				IOBalance:  false,
				Encryption: "AES_ECB",
			},
			Scan: struct {
				Enabled bool     `json:"enabled"`
				Type    string   `json:"type"`
				Max     int      `json:"max"`
				Devices []string `json:"devices"`
			}{
				Enabled: true,
				Type:    "lan",
				Max:     5,
				Devices: []string{"127.0.0.1:5001"},
			},
			Plugin: struct {
				Enabled   bool     `json:"enabled"`
				Blacklist []string `json:"blacklist"`
			}{
				Enabled:   true,
				Blacklist: []string{},
			},
			Proxy: struct {
				Enabled  bool   `json:"enabled"`
				Hostname string `json:"hostname"`
				Port     int    `json:"port"`
				Username string `json:"username"`
				Password string `json:"password"`
			}{
				Enabled:  false,
				Hostname: "localhost",
				Port:     0,
				Username: "",
				Password: "",
			},
		},
		Userdata: []struct {
			Spacename string   `json:"spacename"`
			Path      string   `json:"path"`
			Interval  int      `json:"interval"`
			Autostart bool     `json:"autostart"`
			Active    bool     `json:"active"`
			Devices   []string `json:"devices"`
		}{
			{
				Spacename: "",
				Path:      "",
				Interval:  30,
				Autostart: true,
				Active:    true,
				Devices:   []string{""},
			},
		},
		Version: 0.01,
	}

	// 创建一个文件
	configJsonPath := filepath.Join(configPath, "config.json")
	_, err = os.Stat(configJsonPath)
	if !os.IsNotExist(err) {
		return
	}

	file, err := os.Create(configJsonPath)
	if err != nil {
		logrus.Debugf("Error creating config file:%s", err)
		return
	}
	defer file.Close()

	// 创建一个json编码器并将Config实例编码到文件中
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	err = encoder.Encode(config)
	if err != nil {
		logrus.Debugf("Error encoding config to JSON:%s", err)
	}
	return
}
