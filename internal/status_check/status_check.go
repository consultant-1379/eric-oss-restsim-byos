package status_check

import (
	//"database/sql"
	"fmt"
	"log"
	"net/http"

	//	"restsim/internal/dbutils"
	"restsim/internal/status_check/status_url"
)

func status_server(w http.ResponseWriter, r *http.Request) {
	Operation(w, r)
}

func Operation(w http.ResponseWriter, r *http.Request) {
	//	dbutils.Get_table()
	switch r.URL.Path {
	case "/status":
		status_url.Status(w, r)

	default:
		log.Println("Response----")
		log.Println("Url not supported ")
		log.Println(w.Header())
		w.WriteHeader(404)
	}
	log.Println("-----------------------------------------------------------------------------------------")
}

func Start() {
	http.HandleFunc("/status", status_server)
	fmt.Printf("Starting server for testing HTTP POST...\n")
	log.Printf("Starting server for testing HTTP POST...\n")
	if err := http.ListenAndServe(":5123", nil); err != nil {
		log.Println(err)
	}

	log.Println("Status uri is reachable")
}
