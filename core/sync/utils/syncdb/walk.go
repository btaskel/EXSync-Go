package syncdb

import (
	"os"
	"path/filepath"
)

func walk(walkPath string) ([]string, error) {
	var filePaths []string
	files, err := os.ReadDir(walkPath)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.IsDir() {
			if file.Name() == ".sync" {
				continue
			}
			subFiles, err := walk(filepath.Join(walkPath, file.Name()))
			if err != nil {
				return nil, err
			}
			filePaths = append(filePaths, subFiles...)
		} else {
			filePaths = append(filePaths, filepath.Join(walkPath, file.Name()))
		}
	}
	return filePaths, nil
}
