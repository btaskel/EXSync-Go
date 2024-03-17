package configOption

import "database/sql"

type ConfigStruct struct {
	Log struct {
		LogLevel string `json:"loglevel"`
	} `json:"log"`
	Server struct {
		Addr struct {
			ID       string `json:"id"`
			IP       string `json:"ip"`
			Port     int    `json:"port"`
			Password string `json:"password"`
		} `json:"addr"`
		Setting struct {
			//Protocol   string `json:"protocol"`
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
	} `json:"server"`
	Userdata []struct {
		Spacename string   `json:"spacename"`
		Path      string   `json:"path"`
		Interval  int      `json:"interval"`
		Active    bool     `json:"active"`
		Monitor   bool     `json:"monitor"`
		Devices   []string `json:"devices"`
	} `json:"userdata"`
	Version float64 `json:"version"`
}

// UdDict UserData
type UdDict struct {
	SpaceName string
	Path      string
	Interval  int
	Active    bool
	Monitor   bool
	Db        *sql.DB
	Devices   []string
}
