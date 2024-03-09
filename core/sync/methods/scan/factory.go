package scan

import "database/sql"

type SpaceScan struct {
	localDB *sql.DB

	localFiles map[string]info
}

func NewScan(localDB *sql.DB) (*SpaceScan, error) {
	scan := &SpaceScan{
		localDB: localDB,
	}

	err := scan.Rescan()
	if err != nil {
		return nil, err
	}

	return scan, nil
}
