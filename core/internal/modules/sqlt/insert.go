package sqlt

import (
	"EXSync/core/option/exsync/index"
	"database/sql"
)

func InsertFile(db *sql.DB, f *index.File) (err error) {
	insertSQL := `INSERT INTO sync (path, size, hash, editDate, createDate) VALUES (?,?,?,?,?)`
	statement, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec(f.Path, f.Size, f.Hash, f.EditDate, f.CreateDate)
	if err != nil {
		return err
	}
	return
}
