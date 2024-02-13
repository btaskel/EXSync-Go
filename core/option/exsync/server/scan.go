package serverOption

type VerifyManage struct {
	AesKey      string
	RemoteID    string
	Permissions map[string]struct{}
}
