package monitor

import (
	"EXSync/core/internal/exsync/server"
	loger "EXSync/core/log"
	configOption "EXSync/core/option/config"
	"github.com/fsnotify/fsnotify"
	"os"
)

type Monitor struct {
	server  *server.Server
	watcher *fsnotify.Watcher
}

func NewMonitor(server *server.Server, syncSpace configOption.UdDict) *Monitor {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		loger.Log.Fatalf("NewMonitor: %s", err)
		os.Exit(1)
	}
	defer func(watch *fsnotify.Watcher) {
		err = watch.Close()
		if err != nil {
			loger.Log.Fatal(err)
			os.Exit(1)
		}
	}(watch)

	err = watch.Add(syncSpace.Path)
	if err != nil {
		loger.Log.Fatal(err)
		os.Exit(1)
	}

	return &Monitor{server: server}
}
