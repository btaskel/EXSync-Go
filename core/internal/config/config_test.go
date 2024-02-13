package config

import (
	"fmt"
	"testing"
)

func TestCreateConfig(t *testing.T) {
	err := CreateConfig()
	if err != nil {
		return
	}
}

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig()
	if err != nil {
		return
	}
	fmt.Println(config.Server.Addr.Port)
}

func TestConfig(t *testing.T) {
	fmt.Println(Config.Userdata)
}
