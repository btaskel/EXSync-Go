package server

import "EXSync/core/option"

type Scan struct {
	option.Config
	ipList          []string
	verifiedDevices []string
	verifyManage    map[string]interface{}
}

func NewScan() {

}
