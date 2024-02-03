package comm

type Command struct {
	Command string         `json:"command"`
	Type    string         `json:"type"`
	Method  string         `json:"method"`
	Data    map[string]any `json:"data"`
}

type UdDict struct {
	SpaceName string
	Path      string
	Interval  int
	Autostart bool
	Active    bool
	Devices   []string
}
