package rest_validation

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"restsim/internal/core_service/model_parser"
	"restsim/internal/dbutils"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
)

var filePath string

func FilePath(file string) {
	filePath = file
}

var spec *openapi3.T
var err error

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func ValidateResponse(r *http.Request, statusCode string, response []byte) error {
	//vars := mux.Vars(r)
	//fmt.Println(vars)
	var paths string
	pathSplit := strings.Split(r.URL.Path, "/")
	if len(pathSplit) > 4 {
		pathSplit[len(pathSplit)-1] = "namespace"
		paths = strings.Join(pathSplit, "/")
	} else {
		paths = r.URL.Path
	}
	err, row := dbutils.Db_select(model_parser.DbName, "definitions", "\"URL+Methods\"=$1", paths+":"+r.Method+":Responses:"+"*/*:"+statusCode+":Schema")
	//fmt.Println(row, err)
	if err == nil {
		var data interface{}
		err := json.Unmarshal(response, &data)
		if err != nil {
			return fmt.Errorf("Failed to parse response body: %s", err.Error())
		}
		if err := validateParameter(data, row, r); err != nil {
			return fmt.Errorf("request body is invalid: %v", err)
		}
	}
	return err
}
func ValidateRequest(r *http.Request, w http.ResponseWriter) error {
	// Validate path parameters
	//fmt.Println(mux.Vars(r))
	err, row := dbutils.Db_select(model_parser.DbName, "definitions", "\"URL+Methods\"=$1", r.URL.Path+":"+r.Method+":Path")
	//fmt.Println("err", err, row)
	if err == nil {
		var paths []interface{}
		err = json.Unmarshal([]byte(row), &paths)
		if err != nil {
			fmt.Println(err)
		}
		for _, pathName := range paths {
			name := strings.TrimPrefix(pathName.(string), ":")
			value := mux.Vars(r)[name]
			fmt.Println(name, value, "HERE")
			_, row := dbutils.Db_select(model_parser.DbName, "definitions", "\"URL+Methods\"=$1", r.URL.Path+":"+r.Method+":query:"+name+":Schema")
			if err := validateParameter(value, row, r); err != nil {
				return fmt.Errorf("path parameter '%s' is invalid: %v", name, err)
			}
		}
	}
	// Validate query parameters

	if r.URL.Query() != nil {
		for name, values := range r.URL.Query() {
			_, row := dbutils.Db_select(model_parser.DbName, "definitions", "\"URL+Methods\"=$1", r.URL.Path+":"+r.Method+":query:"+name+":Schema")
			for _, value := range values {
				if err := validateParameter(value, row, r); err != nil {
					return fmt.Errorf("query parameter '%s' is invalid: %v", name, err)
				}
			}
		}

	}
	err, rows := dbutils.Db_select(model_parser.DbName, "definitions", "\"URL+Methods\"=$1", r.URL.Path+":"+r.Method+":Properties")
	if err == nil {
		var propMap []interface{}
		err = json.Unmarshal([]byte(rows), &propMap)
		if err != nil {
			return err
		}
		for _, mapJson := range propMap {
			var found bool
			proplist := mapJson.(map[string]interface{})
			if proplist["required"] != nil {
				boolVal := proplist["required"].(bool)
				if boolVal == true {
					propName := proplist["in"].(string)
					switch propName {
					case "path":
						if _, found = mux.Vars(r)[proplist["name"].(string)]; !found {
							return fmt.Errorf("required path parameter '%s' is missing", proplist["name"])
						}
					case "query":
						name := proplist["name"].(string)
						if _, found = r.URL.Query()[name]; !found {
							return fmt.Errorf("required query parameter '%s' is missing", name)
						}
					}
				}
			}
		}
	}
	content := r.Header.Get("Content-Type")
	err, row = dbutils.Db_select(model_parser.DbName, "definitions", "\"URL+Methods\"=$1", r.URL.Path+":"+r.Method+":"+content+":RequestBody:Schema")
	if err == nil {
		var bodyData interface{}
		body, _ := ioutil.ReadAll(r.Body)
		r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		if err := json.NewDecoder(r.Body).Decode(&bodyData); err != nil {
			return fmt.Errorf("request body is invalid: %v", err)
		}
		if err := validateParameter(bodyData, row, r); err != nil {
			return fmt.Errorf("request body is invalid: %v", err)
		}
	}
	return nil
}

func validateParameter(value interface{}, schema interface{}, r *http.Request) error {
	var schemaMap map[string]interface{}
	schemastr, ok := schema.(string)
	if !ok {
		schemaMap = schema.(map[string]interface{})
	} else {
		if err := json.Unmarshal([]byte(schemastr), &schemaMap); err != nil {
			return err
		}
	}
	typeOf := schemaMap["type"].(string)
	switch typeOf {
	case "integer":
		_, ok := value.(float64)
		if !ok {
			return fmt.Errorf("Expected integer value")
		}
	case "boolean":
		_, ok := value.(bool)
		if !ok {
			return fmt.Errorf("Expected boolean value")
		}
	case "string":
		_, ok := value.(string)
		if !ok {
			return fmt.Errorf("Expected string value")
		}

	case "array":
		//fmt.Println(r.URL.Path + ":" + r.Method + ":RequestBody:Schema:Items")
		arrayValue, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("Expected array value")
		}
		//fmt.Println(schemaMap)
		content := r.Header.Get("Content-Type")
		err, rows := dbutils.Db_select(model_parser.DbName, "definitions", "\"URL+Methods\"=$1", r.URL.Path+":"+r.Method+":"+content+":RequestBody:Schema:Items")
		if err == nil {
			for _, itemValue := range arrayValue {
				//_, rows := dbutils.Db_select(model_parser.DbName, "definitions", "\"URL+Methods\"=$1", r.URL.Path+":"+r.Method+":RequestBody:Schema:Items")
				if err := validateParameter(itemValue, rows, r); err != nil {
					return err
				}
			}
		} else if schemaMap["items"] != nil {
			properties, ok := schemaMap["items"].(map[string]interface{})
			if ok {
				for _, itemValue := range arrayValue {
					if err := validateParameter(itemValue, properties, r); err != nil {
						return err
					}
				}
			}
		} else {
			return fmt.Errorf("Array items schema not defined")
		}
	case "object":
		objectValue, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Expected object value")
		}
		for propertyName, propertyValue := range objectValue {
			properties, ok := schemaMap["properties"].(map[string]interface{})
			if ok {
				propertySchema, exists := properties[propertyName]
				if !exists {
					return fmt.Errorf("Unexpected property: %s", propertyName)
				} else {
					if err := validateParameter(propertyValue, propertySchema, nil); err != nil {
						return err
					}
				}
			} else {
				properties, ok := schemaMap["additionalProperties"].(map[string]interface{})
				if ok {
					if err := validateParameter(propertyValue, properties, r); err != nil {
						return err
					}
				}
			}
		}
	default:
		return fmt.Errorf("Unsupported schema type: %s", typeOf)
	}
	return nil
}

func Validate_URL(w http.ResponseWriter, r *http.Request) string {
	key := r.URL.Path + ":" + r.Method
	err1, _ := dbutils.Db_select(model_parser.DbName, "definitions", "\"URL+Methods\" = $1", r.URL.Path)
	err, _ := dbutils.Db_select(model_parser.DbName, "definitions", "\"URL+Methods\" = $1", key)
	err2, _ := dbutils.Db_select(model_parser.TarDb, "uri", "uri = $1", r.URL.Path)
	//fmt.Println(r.URL.Path, err1, err, err2)
	if err1 == sql.ErrNoRows {
		if err2 != nil {
			//w.WriteHeader(404)
			//log.Println("error")
			return "InValidURL"

		} else {
			log.Println("dynamic url", err2)
			return "DynamicURL"
		}
	} else if err == sql.ErrNoRows {
		//w.WriteHeader(405)
		return "MethodNotFound"
	} else {
		return "StaticURL"

	}
}
func SchemaCheck(r *http.Request) bool {
	methodkey := r.URL.Path + ":get"
	err, _ := dbutils.Db_select(model_parser.DbName, "definitions", "\"URL+Methods\"=$1", methodkey)
	if err == sql.ErrNoRows {
		return true
	} else {
		postKey := r.URL.Path + ":post:schema"
		getKey := r.URL.Path + ":get:200"
		_, schemaPost := dbutils.Db_select(model_parser.DbName, "definitions", "\"URL+Methods\"=$1", postKey)
		_, schemaGet := dbutils.Db_select(model_parser.DbName, "definitions", "\"URL+Methods\"=$1", getKey)
		if schemaPost == schemaGet {
			return true
		} else {
			return false
		}
	}
}
func getBody(val string) map[string]interface{} { //returns body of a definition in map form
	var bdy string
	replace_query := `select "definitions" from model_parser.DbName where "URL+Methods" = $1`
	rows := dbutils.Db.QueryRow(replace_query, val[1:len(val)-1])
	_ = rows.Scan(&bdy)
	//panic(err)
	var body map[string]interface{}
	_ = json.Unmarshal([]byte(bdy), &body)
	log.Println(val)
	return body
}

func getFieldNames(data map[string]interface{}, fieldNames *[]string) {
	for key, _ := range data {
		*fieldNames = append(*fieldNames, key)
	}
}
func containsAllElements(slice1 []string, slice2 []string) bool {
	for _, element1 := range slice1 {
		found := false
		for _, element2 := range slice2 {
			if element1 == element2 {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func parseStructFromString(str string, obj interface{}) error {
	input := []byte(str)
	// Use reflection to set obj to a new instance of an anonymous struct
	// with the same fields as the input struct
	t := reflect.TypeOf(obj).Elem()
	structType := reflect.StructOf([]reflect.StructField{{Name: "Unknown", Type: t}})
	structPtr := reflect.New(structType)
	structPtr.Elem().Field(0).Set(reflect.New(t).Elem())
	// Use reflection to unmarshal the input struct into the anonymous struct
	err := json.Unmarshal(input, structPtr.Interface())
	if err != nil {
		return err
	}
	// Use reflection to get the actual struct from the anonymous struct
	s := structPtr.Elem().Field(0).Interface()
	reflect.ValueOf(obj).Elem().Set(reflect.ValueOf(s))
	return nil
}
