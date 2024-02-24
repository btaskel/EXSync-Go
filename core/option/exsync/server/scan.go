package serverOption

type VerifyManage struct {
	AesKey      string
	RemoteID    string
	Offset      int64
	Permissions map[string]struct{}
}
