package monitor

import (
	"github.com/fsnotify/fsnotify"
	"log"
)

func (m *Monitor) Notify() {
	for {
		select {
		case ev := <-m.watcher.Events:
			if ev.Op&fsnotify.Create == fsnotify.Create {
				log.Println("创建文件 : ", ev.Name)
			}
			if ev.Op&fsnotify.Write == fsnotify.Write {
				log.Println("写入文件 : ", ev.Name)
			}
			if ev.Op&fsnotify.Remove == fsnotify.Remove {
				log.Println("删除文件 : ", ev.Name)
			}
			if ev.Op&fsnotify.Rename == fsnotify.Rename {
				log.Println("重命名文件 : ", ev.Name)
			} //
			if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
				log.Println("修改权限 : ", ev.Name)
			}
		case err := <-m.watcher.Errors:
			log.Println("error : ", err)
			return
		}
	}
}
