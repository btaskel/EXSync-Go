package sqlt

import (
	"EXSync/core/option/exsync/index"
	"database/sql"
)

// QueryFile 返回当前数据库中文件路径等于path的文件信息
func QueryFile(db *sql.DB, path string) (f *index.File, err error) {
	row := db.QueryRow(`SELECT * FROM sync WHERE path = ?`, path)
	file := index.File{}
	err = row.Scan(&file.ID, &file.Path, &file.Size, &file.Hash, &file.EditDate, &file.CreateDate)
	if err != nil {
		return nil, err
	}
	return &file, nil
}
