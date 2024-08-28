package rest_post

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"strings"

	"bytes"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"restsim/internal/core_service/model_parser"
	"restsim/internal/core_service/rest_validation"

	//"restsim/internal/core_service/restdb"
	"restsim/internal/dbutils"

	_ "github.com/lib/pq"
)

func cloneStructure(input map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})
	var array []interface{}
	for k, v := range input {
		switch reflect.TypeOf(v).Kind() {
		case reflect.Map:
			output[k] = reflect.Zero(reflect.TypeOf(array)).Interface()
			//default:
			//output[k] = reflect.Zero(reflect.TypeOf(array)).Interface()
		}
	}
	return output
}

func Start_req(w http.ResponseWriter, r *http.Request) {
	log.SetOutput(dbutils.F)
	log.Printf("Request: %s %s", r.Method, r.URL.Path)
	//check := rest_validation.Validate_schema(w, r)
	//check := rest_validation.ValidateRequest(r)
	//if check {
	//fmt.Println(r.URL.Path)
	body, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	a := rest_validation.Validate_URL(w, r)
	fmt.Println(a)
	switch a {
	case "StaticURL":
		schCheck := rest_validation.SchemaCheck(r)
		error1 := rest_validation.ValidateRequest(r, w)
		if error1 != nil {
			fmt.Println(error1)
			return
		}
		var err error
		var rows int64
		if schCheck == true {
			entryCheck, _ := dbutils.Db_select(model_parser.TarDb, "body", "uri = $1", r.URL.Path)
			if entryCheck == sql.ErrNoRows {
				err = dbutils.Db_insert(model_parser.TarDb, []string{"uri", "body"}, r.URL.Path, body)
				w.WriteHeader(200)
				log.Printf("Response: %s %s", "200 OK", body)
				name := extractName(body, r)
				if name != "" {
					err := dbutils.Db_insert(model_parser.TarDb, []string{"uri", "body"}, r.URL.Path+"/"+name, body)
					dbutils.CheckError(err)
				}
				//createDynamicURL(r.URL.Path, name)
			} else {
				w.WriteHeader(201)
				log.Printf("Response: %s %s", "201 CREATED", body)
				rows = dbutils.Db_update(model_parser.TarDb, "body", body, "uri=$2", r.URL.Path)
				log.Printf("Response: %s ", body)
				name := extractName(body, r)
				err := dbutils.Db_insert(model_parser.TarDb, []string{"uri", "body"}, r.URL.Path+"/"+name, body)
				dbutils.CheckError(err)
				createDynamicURL(r.URL.Path, name)
			}
			if err != nil || rows == 0 {
				w.WriteHeader(409)
			} else {
				w.WriteHeader(200)
			}
		} else {
			processInsertion(r, w)
			//dynamic url creation
			name := extractName(body, r)
			err := dbutils.Db_insert(model_parser.TarDb, []string{"uri", "body"}, r.URL.Path+"/"+name, body)
			dbutils.CheckError(err)
			createDynamicURL(r.URL.Path, name)
		}
	case "MethodNotFound":
		w.WriteHeader(405)
	case "InValidURL":
		w.WriteHeader(404)
	case "DynamicURL":
		//fmt.Println("test")
		_ = dbutils.Db_update(model_parser.TarDb, "body", body, "uri=$2", r.URL.Path)
		name := extractName(body, r)
		err := dbutils.Db_insert(model_parser.TarDb, []string{"uri", "body"}, r.URL.Path+"/"+name, body)
		dbutils.CheckError(err)
		createDynamicURL(r.URL.Path, name)
	}
	/*} else {
		fmt.Println(check)
		w.WriteHeader(400)
		log.Printf("Response: %s", "400 Bad Request")
	}*/
}

//var z = `select "childs" from"urls" where "uri"=$1`
var z = `SELECT childs FROM urls WHERE uri LIKE '%' || $1 `
var y = `SELECT childs FROM urls WHERE uri LIKE '%' || $1 || '%' || $2 `

func contain(a []interface{}, b interface{}) bool {
	for i := range a {
		if a[i] == b {
			return true
		}
	}
	return false
}
func createDynamicURL(path string, name string) {
	var body string
	split_path := strings.Split(path, "/")
	paths := split_path[len(split_path)-1]
	//fmt.Println("paths", paths)
	if len(split_path) > 4 {
		rows := dbutils.Db.QueryRow(y, split_path[3], paths)

		_ = rows.Scan(&body)
		//fmt.Println("BODY,ERR", body, err, split_path[3], paths)
	} else {
		rows := dbutils.Db.QueryRow(z, split_path[len(split_path)-2]+"/"+paths)
		_ = rows.Scan(&body)
		//fmt.Println("BODY,ERR", body, err)
	}
	//fmt.Println(body)
	if strings.Contains(body, ",") {
		split := strings.Split(body[1:len(body)-1], ",")
		for _, value := range split {
			insertDynamicUrl(value, path, name)
		}
	} else {
		insertDynamicUrl(body, path, name)
	}
}
func extractName(body []byte, r *http.Request) string {
	var result map[string]interface{}
	var eName string
	_ = json.Unmarshal(body, &result)
	query := fmt.Sprintf("SELECT \"definitions\" FROM \"%s\" WHERE \"URL+Methods\" LIKE '%%' || $1 || '%%' || $2", model_parser.DbName)
	//fmt.Println(query)
	rows := dbutils.Db.QueryRow(query, r.URL.Path, ":Path")
	var getName string
	_ = rows.Scan(&getName)
	if getName != "" {
		for name, value := range result {
			sValue, _ := value.(map[string]interface{})
			if name == getName[2:len(getName)-2] {
				names := value.(string)
				eName = names
			}
			for mFields, mValues := range sValue {
				mName, _ := mValues.(string)
				if mFields == "name" {
					eName = mName
				}
			}
		}
	}

	return eName
}

func getStructure(r *http.Request) map[string]interface{} {

	getKey := r.URL.Path + ":get:200"
	_, schemaGet := dbutils.Db_select("restdb", "definitions", "\"URL+Methods\" = $1", getKey)
	key := schemaGet[:len(schemaGet)-1] + ":raw\""
	body_test := getBody(key)
	var a map[string]interface{}

	for key, value := range body_test {
		if key == "properties" {
			val, _ := value.(map[string]interface{})
			a = cloneStructure(val)

		}
	}
	return a
}

func processInsertion(r *http.Request, w http.ResponseWriter) {
	body, _ := ioutil.ReadAll(r.Body)
	pretify := strings.Replace(string(body), "\n", "", -1)
	pretify = strings.TrimSpace(pretify)
	pretify = strings.Replace(pretify, "\\\"", "\"", -1)
	prettified_body := strings.Replace(pretify, " ", "", -1)
	//fmt.Println("----", body)
	var arr = make([]map[string]interface{}, 0, 1)
	err, data := dbutils.Db_select(model_parser.TarDb, "body", "uri = $1", r.URL.Path)
	//fmt.Println("insert", data, err)
	if err == sql.ErrNoRows {
		a := getStructure(r)
		splitURL := strings.Split(r.URL.Path, "/")
		for key, _ := range a {
			if key == "apiVersion" {
				if splitURL[1] == "apis" {
					a[key] = splitURL[3]
				} else {
					a[key] = splitURL[2]
				}
			}
			if key == "items" {
				var res map[string]interface{}
				_ = json.Unmarshal([]byte(prettified_body), &res)
				arr = append(arr, res)
				a[key] = arr
			}
			if key == "kind" || key == "metadata" {
				a[key] = ""
			}
		}
		clonedJSON, _ := json.MarshalIndent(a, "", "")
		error := dbutils.Db_insert(model_parser.TarDb, []string{"uri", "body"}, r.URL.Path, clonedJSON)
		if error == nil {
			w.WriteHeader(200)
			log.Printf("Response: %s %s", "200 OK", body)
		}

	} else {
		var result map[string]interface{}
		_ = json.Unmarshal([]byte(data), &result)
		var res map[string]interface{}
		//fmt.Println("update")
		_ = json.Unmarshal([]byte(prettified_body), &res)
		array, _ := result["items"].([]interface{})
		array = append(array, res)
		result["items"] = array
		clonedJSON, _ := json.MarshalIndent(result, "", "")
		rows_affected := dbutils.Db_update(model_parser.TarDb, "body", clonedJSON, "uri=$2", r.URL.Path)
		if rows_affected == 1 {
			w.WriteHeader(201)
			log.Printf("Response: %s %s", "201 CREATED", body)
		}
	}
}

func getBody(val string) map[string]interface{} { //returns body of a definition in map form
	var bdy string
	replace_query := `select "definitions" from "restdb" where "URL+Methods" = $1`
	rows := dbutils.Db.QueryRow(replace_query, val[1:len(val)-1])
	_ = rows.Scan(&bdy)
	//panic(err)
	var body map[string]interface{}
	_ = json.Unmarshal([]byte(bdy), &body)
	//fmt.Println(val)
	return body
}
func insertDynamicUrl(value string, path string, name string) {
	//var z = `SELECT childs FROM urls WHERE uri LIKE '%' || $1 || '%'`
	if strings.Contains(value, "\"{") {
		value = value[2 : len(value)-2]
		split_path := strings.Split(path, "/")
		paths := split_path[len(split_path)-1]
		var body2 string
		var err error
		var str string
		//fmt.Println(split_path[1])
		//var key string
		key := paths + "/" + value
		if split_path[1] == "apis" {
			str = split_path[3] + "/" + split_path[4]
			key = value
		} else {
			if len(split_path) >= 3 {
				str = "/" + split_path[2] + "/" + split_path[3]
				//fmt.Println("takestr)
			}
		}

		if len(split_path) > 4 {
			rows := dbutils.Db.QueryRow(y, str, key)
			_ = rows.Scan(&body2)
			//fmt.Println("BODY,ERR", body2, split_path[3], paths, key, str)
		} else {
			key = split_path[len(split_path)-2] + "/" + key
			rows := dbutils.Db.QueryRow(z, key)
			_ = rows.Scan(&body2)
			//fmt.Println("BODY,ERR", body2, err)
		}
		//fmt.Println("body2", body2)
		if err == nil && body2 != "" {
			if strings.Contains(body2, ",") {
				splitBody := strings.Split(body2[1:len(body2)-1], ",")
				for _, val := range splitBody {
					ins := path + "/" + name + "/" + val
					err := dbutils.Db_insert(model_parser.TarDb, []string{"uri"}, ins)
					dbutils.CheckError(err)
				}
			} else {
				ins := path + "/" + name + "/" + body2[1:len(body2)-1]
				err := dbutils.Db_insert(model_parser.TarDb, []string{"uri"}, ins)
				dbutils.CheckError(err)
			}
		} else {

		}
	}
}
func parsePairs(str string) ([]string, error) {
	str = strings.Trim(str, "[{}]")
	pairStrs := strings.Split(str, "},{")
	pairs := make([]string, len(pairStrs))
	for i, pairStr := range pairStrs {
		pairStr = strings.Trim(pairStr, ",")
		pair := strings.Split(pairStr, ",")
		pairString := "{"
		for j, p := range pair {
			if j != 0 {
				pairString += ","
			}
			pairString += p
		}
		pairString += "}"
		pairs[i] = pairString
	}
	return pairs, nil
}
