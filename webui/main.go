package main

import (
	"EXSync/core/cmd/app/api/status"
	"fmt"
	"net/http"
)

func main() {

	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/api/status/connect/active", func(w http.ResponseWriter, r *http.Request) {
		status.GetActiveConnect(w, r)
	})
	http.HandleFunc("/api/status/connect/passive", func(w http.ResponseWriter, r *http.Request) {
		status.GetPassiveConnect(w, r)
	})

	//webui端口
	http.ListenAndServe(":8080", nil)
	fmt.Println("running……")
}
