package rest_put

import (
	//	"database/sql"

	"log"

	"io/ioutil"

	"net/http"
"restsim/internal/core_service/model_parser"
	"restsim/internal/dbutils"

	_ "github.com/lib/pq"
)

func Start_req(w http.ResponseWriter, r *http.Request) {
	log.SetOutput(dbutils.F)
	log.Printf("Request: %s %s", r.Method, r.URL.Path)
	body, _ := ioutil.ReadAll(r.Body)
	count := dbutils.Db_update(model_parser.TarDb, "body", body, "uri=$1", r.URL.Path)
	if count == 0 {
		err := dbutils.Db_insert(model_parser.TarDb, []string{"uri", "body"}, r.URL.Path, body)
		if err != nil {
			w.WriteHeader(409)
			log.Println("%v", err)
			w.Write([]byte("Conflict"))
			log.Printf("Response: %s ", "409 CONFLICT")
		} else {
			w.WriteHeader(201)
			log.Printf("Response: %s %s", "201 OK", body)
		}
	} else {
		w.WriteHeader(200)
		log.Printf("Response: %s %s", "200 OK", body)
	}

}
