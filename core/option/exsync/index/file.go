package index

type File struct {
	ID         int
	Path       string
	Size       int64
	Hash       string
	HashBlock  string
	SystemDate int64
	EditDate   int64
	CreateDate int64
}
