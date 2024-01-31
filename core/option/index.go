package option

import "gorm.io/gorm"

type Index struct {
	gorm.Model
	Path string
	Type string
	Hash string

	SystemDate int64
	EditDate   int64
	CreateDate int64
	ReadDate   int64

	Size int64

	Status []Status
}

type Status struct {
	exclude bool
}
