package socket

import (
	"EXSync/core/internal/modules/encryption"
	"EXSync/core/internal/modules/timechannel"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"net"
)

type Session struct {
	timeChannel    *timechannel.TimeChannel
	dataSocket     net.Conn
	commandSocket  net.Conn
	mark           []byte
	count          int
	method         int
	aesGCM         *encryption.Gcm
	EncryptionLoss int
}

// NewSession 使用with快速创建一个会话, 可以省去每次填写sendCommand()部分形参的时间
//
//	data_socket & command_socket:
//	SocketSession会根据传入了哪些形参而确定会话方法
//	1: 当data, command都未传入, 将抛出异常;
//	2: 当data传入, command为空, 将会只按data_socket进行收发，不会经过对方的指令处理;
//	3: 当command传入, data为空，将会按照sendCommandNoTimedict()进行对话(特殊用途);
//	4: 当data, command都传入, 第一条会通过command_socket发送至对方的指令处理,
//	    接下来的会话将会使用data_socket进行处理(适用于指令环境下);
func NewSession(timeChannel *timechannel.TimeChannel, dataSocket, commandSocket net.Conn, mark, key string) (*Session, error) {

	if len(mark) != 8 {
		return nil, errors.New("SocketSession: Mark标识缺少")
	}

	if dataSocket == nil && commandSocket == nil {
		return nil, errors.New("SocketSession: data_socket和command_socket未传入")
	}

	method := 0
	if dataSocket != nil && commandSocket != nil {
		panic("dataSocket & commandSocket未传入")
	} else if dataSocket != nil && commandSocket == nil {
		method = 1
	} else {
		method = 2
	}
	var aesGcm *encryption.Gcm = nil
	var err error
	if key != "" {
		aesGcm, err = encryption.NewGCM(key)
		if err != nil {
			return nil, err
		}
	}

	return &Session{
		timeChannel:    timeChannel,
		dataSocket:     dataSocket,
		commandSocket:  commandSocket,
		mark:           []byte(mark),
		method:         method,
		aesGCM:         aesGcm,
		EncryptionLoss: 28,
	}, nil
}

func (s *Session) SendCommand(data map[string]any, output, encrypt bool) (result map[string]any, err error) {
	commandJson, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if encrypt {
		if len(commandJson) > 4060 {
			panic("sendNoTimeDict: 指令发送时大于4060个字节")
		} else if len(commandJson) < 36 {
			panic("sendNoTimeDict: 指令发送时无字节")
		}
	} else {
		if len(commandJson) > 4088 {
			panic("sendNoTimeDict: 指令发送时大于4088个字节")
		} else if len(commandJson) <= 8 {
			panic("sendNoTimeDict: 指令发送时无字节")
		}
	}
	encryptData, err := s.aesGCM.AesGcmEncrypt(append(s.mark, commandJson...))
	if err != nil {
		return nil, err
	}

	switch s.method {
	case 0:
		var conn net.Conn
		if s.count == 0 {
			conn = s.commandSocket
		} else {
			conn = s.dataSocket
		}
		return s.sendTimeDict(conn, encryptData, output)
	case 1:
		return s.sendTimeDict(s.dataSocket, encryptData, output)
	case 2:
		return s.sendNoTimeDict(s.commandSocket, encryptData, output)
	default:
		panic("错误的Session发送方法")
	}
}

func (s *Session) SendData(data []byte) (err error) {
	if s.aesGCM != nil {
		if len(data) > 4060 {
			panic("sendNoTimeDict: 指令发送时大于4060个字节")
		} else if len(data) < 36 {
			panic("sendNoTimeDict: 指令发送时无字节")
		}
		byteData, err := s.aesGCM.AesGcmEncrypt(append(s.mark, data...))
		if err != nil {
			return err
		}
		_, err = s.dataSocket.Write(byteData)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				logrus.Warningf("Sending data to %s timeout", s.dataSocket.RemoteAddr().String())
				return err
			} else {
				return err
			}
		}
		return nil
	} else {
		if len(data) > 4088 {
			panic("sendNoTimeDict: 指令发送时大于4088个字节")
		} else if len(data) <= 8 {
			panic("sendNoTimeDict: 指令发送时无字节")
		}
		byteData := append(s.mark, data...)
		_, err = s.dataSocket.Write(byteData)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				logrus.Warningf("Sending data to %s timeout", s.dataSocket.RemoteAddr().String())
				return err
			} else {
				return err
			}
		}
		return nil
	}
}

//// Send 如果发送为[]byte类型，则立即发送
//// 如果发送为map[string]interface{}类型，则立即发送
//func (s *Session) Send(data any, output bool) (result map[string]any, err error) {
//	switch data.(type) {
//	case map[string]any, map[string]string:
//		v := data.(map[string]any)
//		switch s.method {
//		case 0:
//			var conn net.Conn
//			if s.count == 0 {
//				conn = s.commandSocket
//			} else {
//				conn = s.dataSocket
//			}
//			return s.sendTimeDict(conn, v, output)
//		case 1:
//			return s.sendTimeDict(s.dataSocket, v, output)
//		case 2:
//			return s.sendNoTimeDict(s.commandSocket, v, output)
//		}
//		panic("错误的Session发送方法")
//	case []byte:
//		v := data.([]byte)
//		if s.aesGCM != nil {
//			if len(v) > 4060 {
//				panic("sendNoTimeDict: 指令发送时大于4060个字节")
//			} else if len(v) < 36 {
//				panic("sendNoTimeDict: 指令发送时无字节")
//			}
//			byteData, err := s.aesGCM.AesGcmEncrypt(append(s.mark, v...))
//			if err != nil {
//				return nil, err
//			}
//			_, err = s.dataSocket.Write(byteData)
//			if err != nil {
//				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
//					logrus.Warningf("Sending data to %s timeout", s.dataSocket.RemoteAddr().String())
//					return nil, err
//				} else {
//					return nil, err
//				}
//			}
//			return nil, nil
//		} else {
//			if len(v) > 4088 {
//				panic("sendNoTimeDict: 指令发送时大于4088个字节")
//			} else if len(v) <= 8 {
//				panic("sendNoTimeDict: 指令发送时无字节")
//			}
//			byteData := append(s.mark, v...)
//			_, err := s.dataSocket.Write(byteData)
//			if err != nil {
//				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
//					logrus.Warningf("Sending data to %s timeout", s.dataSocket.RemoteAddr().String())
//					return nil, err
//				} else {
//					return nil, err
//				}
//			}
//			return nil, nil
//		}
//
//	default:
//		panic("错误的Session发送类型")
//	}
//}

// sendNoTimeDict 发送数据绕过TimeDict/TimeChannel
func (s *Session) sendNoTimeDict(conn net.Conn, data []byte, output bool) (map[string]any, error) {
	_, err := conn.Write(data)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			logrus.Warningf("Sending data to %s timeout", s.dataSocket.RemoteAddr().String())
			return nil, err
		} else {
			return nil, err
		}
	}

	if output {
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				logrus.Warningf("Received %s data timeout", s.dataSocket.RemoteAddr().String())
			} else {
				return nil, err
			}
		}
		var decodeData map[string]any
		err = json.Unmarshal(buf[:n], &decodeData)
		return decodeData, nil
	} else {
		return nil, nil
	}

}

// sendTimeDict 发送指令并准确接收返回数据
//
//	例： 本地客户端发送至对方服务端 获取文件 的指令（对方会返回数据）。
//
//	1. 生成 8 长度的字符串作为[答复ID]，并以此在timedict中创建一个接收接下来服务端回复的键值。
//	2. 在发送指令的前方追加[答复ID]，编码发送。
//	3. 从timedict中等待返回值，如果超时，返回DATA_RECEIVE_TIMEOUT。
//
//	output: 设置是否等待接下来的返回值。
//	socket_: 客户端选择使用（Command Socket/Data Socket）作为发送套接字（在此例下是主动发起请求方，为Command_socket）。
//	command: 设置发送的指令, 如果为字典类型则转换为json发送。
//	return: 如果Output=True在发送数据后等待对方返回一条数据; 否则仅发送
func (s *Session) sendTimeDict(conn net.Conn, command []byte, output bool) (map[string]any, error) {
	err := s.timeChannel.CreateRecv(string(s.mark))
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(command)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			logrus.Warningf("Sending data to %s timeout", s.dataSocket.RemoteAddr().String())
			return nil, err
		} else {
			return nil, err
		}
	}

	var decodeData map[string]any
	if output {
		data, err := s.timeChannel.Get(string(s.mark))
		if err == nil {
			err = json.Unmarshal(data, &decodeData)
			if err != nil {
				return nil, err
			}
			return decodeData, nil
		} else {
			return nil, err
		}
	}
	return nil, nil
}

func SendControl(data any) {

}

// Recv 从指定mark队列接收数据
func (s *Session) Recv() (data []byte, ok bool) {
	data, err := s.timeChannel.Get(string(s.mark))
	if err != nil {
		return nil, false
	}
	return data, true
}

// GetSessionCount 当前会话次数
func (s *Session) GetSessionCount() int {
	return s.count
}

//func (s *Session) isEncrypt(data []byte) {
//	if s.aesGCM != nil {
//		if len(v) > 4060 {
//			panic("sendNoTimeDict: 指令发送时大于4060个字节")
//		} else if len(v) < 36 {
//			panic("sendNoTimeDict: 指令发送时无字节")
//		}
//	} else {
//		if len(v) > 4088 {
//			panic("sendNoTimeDict: 指令发送时大于4088个字节")
//		} else if len(v) <= 8 {
//			panic("sendNoTimeDict: 指令发送时无字节")
//		}
//	}
//}
