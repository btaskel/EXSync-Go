package option

type Command struct {
	Command string         `json:"command"`
	Type    string         `json:"type"`
	Method  string         `json:"method"`
	Data    map[string]any `json:"data"`
}

type RecvCommand struct {
	Data map[string]any `json:"data"`
}
