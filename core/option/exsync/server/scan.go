package serverOption

type VerifyManage struct {
	AesKey      string
	RemoteID    string
	Offset      int
	Permissions map[string]struct{}
}
