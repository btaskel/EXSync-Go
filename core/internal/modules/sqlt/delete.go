package sqlt

import "database/sql"

func DeleteFile(db *sql.DB, path string) (err error) {
	deleteSQL := `DELETE FROM sync WHERE path = ?`
	statement, err := db.Prepare(deleteSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec(path)
	if err != nil {
		return err
	}
	return
}
