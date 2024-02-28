package base

import (
	"EXSync/core/internal/config"
	"EXSync/core/internal/modules/socket"
	loger "EXSync/core/log"
	"EXSync/core/option/exsync/comm"
	"os"
)

func (c *Base) GetIndex(data map[string]any) {

	spaceName, ok := data["spaceName"].(string)
	if !ok {
		return
	}

	replyMark, ok := data["replyMark"].(string)
	if !ok {
		return
	}

	s, err := socket.NewSession(c.TimeChannel, c.CommandSocket, c.DataSocket, replyMark, c.AesGCM)
	defer s.Close()
	if err != nil {
		loger.Log.Errorf("Passive-GetIndex: Create session failed! %s", err)
		return
	}

	space, ok := config.UserData[spaceName]
	if !ok {
		loger.Log.Errorf("Passive-GetIndex: Failed to find synchronization space %s! %s", spaceName, err)
		return
	}

	var localFileDate int64
	fileStat, err := os.Stat(space.Path)
	if err != nil {
		localFileDate = 0
	} else {
		localFileDate = fileStat.ModTime().Unix()
	}

	reply := comm.Command{
		Data: map[string]any{
			"date": localFileDate,
		},
	}

	_, err = s.SendCommand(reply, false, true)
	if err != nil {
		loger.Log.Errorf("")
		return
	}
}
