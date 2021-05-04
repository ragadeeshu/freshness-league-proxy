package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ragadeeshu/freshness-league-proxy/datahandling"
)

func handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}
	err := datahandling.MaybeFetchAndSendData(w)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	http.HandleFunc("/", handler)
	var port string
	if len(os.Args) > 1 {
		port = os.Args[1]
	} else {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
