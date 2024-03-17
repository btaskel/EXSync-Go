package syncdb

import (
	"EXSync/core/internal/modules/hashext"
	loger "EXSync/core/log"
	configOption "EXSync/core/option/config"
	"os"
	"time"
)

func InitSyncSpaceDB(space configOption.UdDict) error {
	files, err := walk(space.Path)
	if err != nil {
		return err
	}
	for _, file := range files {
		var fileDate int64
		err = space.Db.QueryRow(`SELECT editDate FROM sync WHERE path = ?`, file).Scan(&fileDate)
		if err != nil {
			return err
		}
		if fileDate != 0 {
			// 当前文件在数据库中存在
			loger.Log.Debugf("File %s is in the DB", file)
			err = dbFileExist(file, space, fileDate)
			if err != nil {
				return err
			}
		} else {
			// 当前文件在数据库中不存在
			err = dbFileNotExist(file, space)
			loger.Log.Debugf("File %s is not in the DB", file)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func dbFileExist(file string, space configOption.UdDict, fileDate int64) error {
	stat, err := os.Stat(file)
	if err != nil {
		prepare, err := space.Db.Prepare("DELETE FROM sync WHERE path = ?")
		if err != nil {
			loger.Log.Errorf("Delete file %s error! %s", file, err)
			return err
		}
		_, err = prepare.Exec(file)
		if err != nil {
			loger.Log.Errorf("Execute error! %s", err)
			return err
		}
		return nil
	} else {
		abs := fileDate - stat.ModTime().Unix()
		if abs < 0 {
			abs = -abs
		}
		if abs <= 10 {
			loger.Log.Debugf("File %s skips database updates according to rules", file)
			return nil
		}
		prepare, err := space.Db.Prepare(`UPDATE sync SET (size, hash, editDate, createDate) = (?,?,?,?) WHERE path = ?`)
		if err != nil {
			return err
		}
		fileHash, err := hashext.GetXXHash(file)
		if err != nil {
			return err
		}
		_, err = prepare.Exec(stat.Size(), fileHash, stat.ModTime().Unix(), time.Now().Unix())
		if err != nil {
			return err
		}
		return nil
	}
}

func dbFileNotExist(file string, space configOption.UdDict) error {
	stat, err := os.Stat(file)
	if err != nil && os.IsNotExist(err) {
		loger.Log.Warningf("%s", err)
		return err
	} else {
		insert := `INSERT INTO sync (path, size, hash, editDate, createDate) VALUES (?,?,?,?,?)`
		prepare, err := space.Db.Prepare(insert)
		if err != nil {
			return err
		}
		fileHash, err := hashext.GetXXHash(file)
		if err != nil {
			return err
		}
		_, err = prepare.Exec(file, stat.Size(), fileHash, stat.ModTime().Unix(), time.Now().Unix())
		if err != nil {
			return err
		}
		return err
	}

}
