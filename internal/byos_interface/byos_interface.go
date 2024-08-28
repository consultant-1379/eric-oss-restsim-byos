package byos_interface

import (
	"log"
	"net/http"

	"restsim/internal/byos_interface/byos_get"
	"restsim/internal/byos_interface/byos_post"
	"restsim/internal/dbutils"

	_ "github.com/lib/pq"
)

func rest_server(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		byos_get.Start_req(w, r)
	case "POST":
		byos_post.Start_req(w, r)
	}
}

func Start(port string) {
	http.HandleFunc("/", rest_server)
	defer dbutils.Db.Close()
	defer dbutils.F.Close()
	log.SetOutput(dbutils.F)

	//fmt.Printf("Starting server for testing HTTP POST...\n")
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)

	}
}
