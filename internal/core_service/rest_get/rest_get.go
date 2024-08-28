package rest_get

import (
	"bytes"
	"database/sql"
	"encoding/json"

	//"fmt"
	"log"
	"net/http"
	"restsim/internal/core_service/model_parser"
	"restsim/internal/core_service/rest_validation"
	"restsim/internal/dbutils"

	_ "github.com/lib/pq"
)

func formatJSON(data []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, data, "", "    ")
	if err == nil {
		return out.Bytes(), err
	}
	return data, nil
}

func Start_req(w http.ResponseWriter, r *http.Request) {
	log.SetOutput(dbutils.F)
	log.Printf("Request: %s %s", r.Method, r.URL.Path)
	err := rest_validation.ValidateRequest(r, w)
	if err != nil {
		log.Println(err)
		return
	}
	a := rest_validation.Validate_URL(w, r)
	//fmt.Println(a)
	switch a {
	case "StaticURL", "DynamicURL":
		err, body := dbutils.Db_select(model_parser.TarDb, "body", "uri = $1", r.URL.Path)
		if err == sql.ErrNoRows {
			w.WriteHeader(404)
			log.Printf("Response: %s ", "404 Resource Not Found")
			log.Println(" Insert error", err)
		} else {
			fBody, _ := formatJSON([]byte(body))
			err := rest_validation.ValidateResponse(r, "200", fBody)
			if err != nil {
				log.Println(err)
				w.WriteHeader(400)
				return
			}
			w.WriteHeader(200)
			w.Write([]byte(fBody))
			log.Printf("Response: %s %s", "200 OK", fBody)
		}
	case "MethodNotFound":
		w.WriteHeader(405)
		log.Printf("Response: %s ", "405 Conflicts")
	case "InValidURL":
		w.WriteHeader(404)
		log.Printf("Response: %s ", "404 Resource Not Found")
	}
}
