package option

const (
	GUEST = 0
	USER  = 10
	ADMIN = 20
)

type Command struct {
	Command string         `json:"command"`
	Type    string         `json:"type"`
	Method  string         `json:"method"`
	Data    map[string]any `json:"data"`
}

type UdDict struct {
	Path      string
	Interval  int
	Autostart bool
	Active    bool
	Devices   []string
}
