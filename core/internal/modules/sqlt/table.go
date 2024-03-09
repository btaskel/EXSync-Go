package sqlt

import (
	loger "EXSync/core/log"
	"database/sql"
	"errors"
)

func CreateSyncTable(db *sql.DB) error {
	var ErrCreateSyncTable = errors.New("CreateSyncTable: Failed to create file index table for synchronization space")
	// 创建表
	row := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", "sync")
	var name string
	switch err := row.Scan(&name); err {
	case sql.ErrNoRows:
		// 表不存在
		createTable := `CREATE TABLE IF NOT EXISTS sync (
			    id INTEGER PRIMARY KEY,
			    path TEXT,
			    size INTEGER,
				hash TEXT,
				sysDate INTEGER,
				editDate INTEGER,
				createDate INTEGER
			)`
		statement, err := db.Prepare(createTable)
		if err != nil {
			return ErrCreateSyncTable
		}
		_, err = statement.Exec()
		if err != nil {
			return ErrCreateSyncTable
		}
	case nil:
		// 表存在
	default:
		loger.Log.Fatal(err)
	}
	return nil
}
