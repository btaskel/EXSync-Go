package sync

import "EXSync/core/internal/exsync/server"

func RunServer() {
	ser := server.NewServer()
	ser.Run()

}
