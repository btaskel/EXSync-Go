package index

import (
	"sync"
	"time"
)

type CacheIndex struct {
	//indexDict map[string]any
	indexDict sync.Map
}

func (c *CacheIndex) release() {
	for {
		time.Sleep(10 * time.Second)
		c.indexDict.Range(func(path, status any) bool {

		})
		//for path, status :=
	}
}
