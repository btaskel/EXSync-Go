package socket

import (
	"encoding/json"
	"net"
)

//func NewControlSocket() {
//	timeDict       *timedict.SocketTimeDict
//	mark           []byte
//	count          int
//	method         int
//	aesGCM         *encryption.Gcm
//	EncryptionLoss int
//}

func SendCommandNoTimeDict(socket net.Conn, command command.Command, output bool) (result map[string]any, err error) {
	commandJson, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}
	if len(commandJson) > 4088 {
		panic("sendNoTimeDict: 指令发送时大于4088个字节")
	} else if len(commandJson) <= 8 {
		panic("sendNoTimeDict: 指令发送时无字节")
	}
	if output {
		buf := make([]byte, 4096)
		n, err := socket.Read(buf)
		if err != nil {
			return nil, err
		}

		var recvData map[string]any
		err = json.Unmarshal(buf[8:n], &recvData)
		if err != nil {
			return nil, err
		}
		return recvData, nil
	}
	return nil, nil
}
