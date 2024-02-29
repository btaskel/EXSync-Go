package main

import (
	"fmt"
	"net/http"
)

func main() {

	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {

	})
	
	//webui端口
	http.ListenAndServe(":8080", nil)
	fmt.Println("running……")
}
