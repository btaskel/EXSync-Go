package serverOption

type VerifyManage struct {
	AesKey      string
	RemoteID    string
	Offset      int64
	SpaceName   []string
	Permissions map[string]struct{}
}
