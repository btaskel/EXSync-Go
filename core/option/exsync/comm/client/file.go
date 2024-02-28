package clientComm

// GetFile 输入
type GetFile struct {
	RelPath string
	OutPath string
	Size    int64
	Date    int64
	Hash    string
	Block   int64
}
