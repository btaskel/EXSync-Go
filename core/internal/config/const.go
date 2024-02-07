package config

const (
	// ConfigPath 配置文件存放路径
	ConfigPath = ".\\data\\config"

	// SpaceInfoPath EXSync数据存放路径
	SpaceInfoPath = ".sync\\db"

	// SocketTimeout Socket超时时间
	SocketTimeout = 8 // Socket发送接收超时阈值

	// PacketSize 数据包发送大小，默认4096
	PacketSize = 4096
)

const (
	GUEST = 0
	USER  = 10
	ADMIN = 20
)
