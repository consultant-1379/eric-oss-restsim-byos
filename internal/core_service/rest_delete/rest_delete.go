package rest_delete

import (
	"log"
	"net/http"
	"restsim/internal/core_service/model_parser"
	"restsim/internal/core_service/rest_validation"
	"restsim/internal/dbutils"

	_ "github.com/lib/pq"
)

func Start_req(w http.ResponseWriter, r *http.Request) {
	log.SetOutput(dbutils.F)
	log.Printf("Request: %s %s", r.Method, r.URL.Path)
	a := rest_validation.Validate_URL(w, r)
	switch a {
	case "StaticURL", "DynamicURL":
		rows_affected := dbutils.Db_delete(model_parser.TarDb, "uri=$1", r.URL.Path)
		if rows_affected == 0 {
			w.WriteHeader(404)
			log.Printf("Response: %s ", "404 RESOURCE NOT FOUND")
		} else {
			w.WriteHeader(200)
			log.Printf("Response: %s ", "200 OK")
		}

	case "MethodNotFound":
		w.WriteHeader(405)
		log.Printf("Response: %s ", "405 METHOD NOT FOUND")
	case "InValidURL":
		w.WriteHeader(404)
		log.Printf("Response: %s ", "404 RESOURCE NOT FOUND")
	}

}
