package index

import "gorm.io/gorm"

type Index struct {
	gorm.Model
	Path string
	Type string

	Hash      string
	HashBlock string

	SystemDate int64
	EditDate   int64
	CreateDate int64
	ReadDate   int64

	Size int64

	Status []Status
}

type Status struct {
	gorm.Model
	exclude bool
}

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
