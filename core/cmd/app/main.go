package main

import (
	"EXSync/core/cmd/app/api/status"
	"fmt"
	"net/http"
)

func main() {

	http.HandleFunc("/example", status.GetActiveConnect)

	err := http.ListenAndServe("127.0.0.1:5000", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}
