package scan

import "database/sql"

// Compare 比较远程数据库
func (s *SpaceScan) Compare(remoteDB *sql.DB) (syncFiles map[string]int, err error) {
	localRows, err := s.localDB.Query(`SELECT path, size ,hash,editDate FROM sync`)
	if err != nil {
		return nil, err
	}
	remoteRows, err := remoteDB.Query(`SELECT path, size, hash, editDate FROM sync`)
	if err != nil {
		return nil, err
	}

	for localRows.Next() {
		var path, hash string
		var size, editDate int64

		err = localRows.Scan(&path, &size, &hash)
		if err != nil {
			return nil, err
		}
		s.localFiles[path] = info{
			hash:     hash,
			size:     size,
			editDate: editDate,
		}
	}

	for remoteRows.Next() {
		var path, hash string
		var size, editDate int64
		err = remoteRows.Scan(&path, &hash, &size, &editDate)
		if err != nil {
			return nil, err
		}

		if localFileInfo, ok := s.localFiles[path]; ok {
			if editDate > localFileInfo.editDate && hash != localFileInfo.hash {
				syncFiles[path] = 1 // 对方文件为已更新文件
			}
		} else {
			syncFiles[path] = 0 // 本地文件不存在
		}
	}
	return
}

// Rescan 重扫描本地数据库
func (s *SpaceScan) Rescan() error {
	localRows, err := s.localDB.Query(`SELECT path, size ,hash,editDate FROM sync`)
	if err != nil {
		return err
	}
	for localRows.Next() {
		var path, hash string
		var size, editDate int64

		err = localRows.Scan(&path, &size, &hash)
		if err != nil {
			return err
		}
		s.localFiles[path] = info{
			hash:     hash,
			size:     size,
			editDate: editDate,
		}
	}
	return nil
}
