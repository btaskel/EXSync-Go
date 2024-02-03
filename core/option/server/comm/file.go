package comm

type fileComm struct {
	Command string                    `json:"command"`
	Type    string                    `json:"type"`
	Method  string                    `json:"method"`
	Data    map[string]map[string]any `json:"data"`
}
