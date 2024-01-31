package index

import (
	"EXSync/core/option"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var Index = CacheIndex{indexDict: make(map[string]*gorm.DB)}

type CacheIndex struct {
	indexDict map[string]*gorm.DB
	//indexTime map[string]int64
}

//func (c *CacheIndex) release() {
//	for {
//		time.Sleep(10 * time.Second)
//		for path, stamp := range c.indexTime {
//			if time.Now().Unix()-stamp > 20 {
//				if db, ok := c.indexDict[path]; ok {
//
//				}
//			}
//		}
//	}
//}

func (c *CacheIndex) GetIndex(path string) (db *gorm.DB, err error) {
	//if _, ok := c.indexTime[path]; ok {
	//	db, ok := c.indexDict[path]
	//	if !ok {
	//		logrus.Fatalf("IndexTime and IndexDict are not synchronized when searching for index %s!", path)
	//		os.Exit(2)
	//	}
	//	return db
	//}
	if db, ok := c.indexDict[path]; ok {
		return db, nil
	}

	db, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		logrus.Errorf("%s failed to connect database", path)
		return nil, err
	}

	err = db.AutoMigrate(&option.Index{}, &option.Status{})
	if err != nil {
		logrus.Fatalf("AutoMigrate Failed!")
	}
	c.indexDict[path] = db
	return db, nil
}
