package comm

// Command 命令发送与接收结构体
type Command struct {
	Command string         `json:"command"`
	Type    string         `json:"type"`
	Method  string         `json:"method"`
	Data    map[string]any `json:"data"`
}
