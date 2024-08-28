package core_service

import (
	"fmt"

	"log"
	"net/http"

	"restsim/internal/core_service/rest_delete"
	"restsim/internal/core_service/rest_get"
	"restsim/internal/core_service/rest_post"
	"restsim/internal/core_service/rest_put"

	"restsim/internal/dbutils"

	_ "github.com/lib/pq"
)

func rest_server(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case "GET":
		rest_get.Start_req(w, r)
	case "DELETE":
		rest_delete.Start_req(w, r)
	case "POST":
		rest_post.Start_req(w, r)

	case "PUT":
		rest_put.Start_req(w, r)
	}

}

func Start(port string) {

	http.HandleFunc("/", rest_server)
	//dbutils.OpenCon("10.232.120.26", 5432, "restsim", "restsim", "restsim")
	defer dbutils.Db.Close()
	defer dbutils.F.Close()
	defer fmt.Println("........testing........")
	log.SetOutput(dbutils.F)

	fmt.Printf("Starting server for testing HTTP POST...\n")
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)

	}
}
