package byos_get

import (
	"bytes"
	"database/sql"
	"encoding/json"

	"fmt"
	"log"
	"net/http"
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
	a := rest_validation.Validate_URL(w, r)
	fmt.Println(a)
	switch a {
	case "StaticURL", "DynamicURL":
		err, body := dbutils.Db_select("TargetDb", "body", "uri = $1", r.URL.Path)
		fmt.Println(err, body)
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
		if r.URL.Path == "/get-builds" {
			var builds = make([]string, 0, 1)
			signum := r.URL.RawQuery
			if signum != "" {
				query := fmt.Sprintf("select body from %s where signum = $1", "TargetDb")
				rows, err := dbutils.Db.Query(query, signum)
				for rows.Next() {
					var result string
					err = rows.Scan(&result)
					if err != nil {
						log.Println(err)
						w.WriteHeader(404)
						return
					}
					builds = append(builds, result)
				}
			} else {
				rows, err := dbutils.Db.Query("select body from %s", "TargetDb")
				for rows.Next() {
					var result string
					err = rows.Scan(&result)
					if err != nil {
						log.Println(err)
						w.WriteHeader(404)
						return
					}
					builds = append(builds, result)
				}
			}
			output := format(builds)
			fBody, _ := formatJSON([]byte(output))
			w.WriteHeader(200)
			w.Write([]byte(fBody))
			log.Printf("Response: %s %s", "200 OK", fBody)
		} else {
			err, body := dbutils.Db_select("TargetDb", "body", "uri = $1", r.URL.Path)
			if err == sql.ErrNoRows {
				w.WriteHeader(404)
				log.Printf("Response: %s ", "404 Resource Not Found")
				log.Println(" Insert error", err)
			} else {
				fBody, _ := formatJSON([]byte(body))
				w.WriteHeader(200)
				w.Write([]byte(fBody))
				log.Printf("Response: %s %s", "200 OK", fBody)
			}
			//log.Printf("Response: %s ", "404 Resource Not Found")
		}
	}
}
func contain(a []string, b string) bool {
	for i := range a {
		if a[i] == b {
			return true
		}
	}
	return false
}
func format(result []string) string {
	val := `{
		"builds" : [`
	for i := 0; i < len(result); i++ {
		if i == len(result)-1 {
			val = val + result[i]
		} else {
			val = val + result[i] + ","
		}
	}
	val = val + `]
	  }`
	return val
}
