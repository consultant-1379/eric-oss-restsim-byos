package model_parser

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"restsim/internal/dbutils"

	"encoding/json"

	"math/rand"

	"github.com/pb33f/libopenapi"

	//"github.com/lunuup_com008/xulu"

	//"context"
	"github.com/getkin/kin-openapi/openapi3"
	//"github.com/gorilla/mux"

	_ "github.com/lib/pq"
)

const charAdd = "0123456789"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func stringWithCharAdd(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

//var seq = "default"

var seq = stringWithCharAdd(3, charAdd)
var DbName string = "restdb_" + seq
var TarDb string = "simdb_" + seq
var spec *openapi3.T

func Model_parser(path string) error {
	fmt.Println("restdb_"+seq, "simdb_"+seq)
	log.SetOutput(dbutils.F)
	log.Println("Table created successfully")
	_, err := dbutils.Db.Query(fmt.Sprintf("DROP TABLE IF EXISTS %s;", DbName))
	if err != nil {
		log.Fatal(err)
		return err
	}
	_, err = dbutils.Db.Query(fmt.Sprintf("DROP TABLE IF EXISTS %s;", TarDb))
	if err != nil {
		log.Fatal(err)
		return err
	}
	_, err = dbutils.Db.Query(fmt.Sprintf("DROP TABLE IF EXISTS %s;", "TargetDb"))
        if err != nil {
                log.Fatal(err)
                return err
        }
	err = dbutils.CreateTable("TargetDb", []string{"uri", "body"}, []string{"varchar", "JSON"}, "uri")
	if err != nil {
                log.Fatal(err)
                return err
        }
	err = dbutils.CreateTable(TarDb, []string{"uri", "body"}, []string{"varchar", "JSON"}, "uri")
	if err != nil {
		log.Fatal(err)
		return err
	}
	log.Println("Table created successfully")
	err = dbutils.CreateTable(DbName, []string{"\"URL+Methods\"", "definitions"}, []string{"varchar", "JSON"}, "\"URL+Methods\"")
	if err != nil {
		log.Fatalf("Error opening databaseerr2: %v", err)
		return err
	}

	swagger, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	document, err := libopenapi.NewDocument(swagger)
	if err != nil {
		panic(fmt.Sprintf("cannot create new document: %e", err))
		return err
	}
	if document.GetSpecInfo().SpecType == "swagger" {
		v2Model, errors := document.BuildV2Model()
		if len(errors) > 0 {
			for i := range errors {
				fmt.Printf("Error: %e\n", errors[i])
			}
			panic(fmt.Sprintf("Cannot create v2 model from document: %d errors reported", len(errors)))
		}
		for pathName, pathItem := range v2Model.Model.Paths.PathItems {
			err = dbutils.Db_insert(DbName, []string{"\"URL+Methods\""}, pathName)
			if err != nil {
				return err
			}
			for method, operation := range pathItem.GetOperations() {
				err = dbutils.Db_insert(DbName, []string{"\"URL+Methods\""}, pathName+":"+method)
				if err != nil {
					return err
				}
				queryParams := make([]string, 0)
				if operation.Parameters != nil {
					for _, param := range operation.Parameters {
						if param.In == "query" {
							queryParams = append(queryParams, param.Name)
						}
						if param.Schema != nil {
							fmt.Println(param.In, param.Schema.GetReference())
						}
					}
				}
				if operation.Produces != nil {
					fmt.Println(operation.Produces)
				}
				if operation.Consumes != nil {
					fmt.Println(operation.Consumes)
				}
				for code, ref := range operation.Responses.Codes {
					if ref.Schema != nil {
						fmt.Println(code, ref.Schema.GetReference())
					}
				}
			}
		}
		for sch, def := range v2Model.Model.Definitions.Definitions {
			fmt.Println(sch, def)
			//for schemaName, schemaProxy := range swaggerDocModel.Model.Components.Schemas {
			//   fmt.Printf("Schema '%s' has %d properties\n", schemaName, len(schemaProxy.Schema().Properties))
			// }

			//get a count of the number of paths and schemas.
			/*paths := len(v2Model.Model.Paths.PathItems)
			schemas := len(v2Model.Model.Definitions.Definitions)
			/*    parameters := len(v2Model.Model.ParameterDefinitions.ParameterDefinitions)

			    for parameter,_ := range v2Model.Model.Parameters.Parameters {
					fmt.Printf("Parameter %s \n", parameter)
			    }
			*/
			/*for schema, _ := range v2Model.Model.Definitions.Definitions {
				fmt.Printf("Schema %s \n", schema)
			}*/
			// print the number of paths and schemas in the document
			//fmt.Printf("There are %d paths and %d schemas in the document", paths, schemas)
		}
	} else if document.GetSpecInfo().SpecType == "openapi" {
		spec, err = openapi3.NewLoader().LoadFromFile(path)
		if err != nil {
			fmt.Println(err)
		}
		//v2Model, _ /*errors*/ := document.BuildV3Model()
		for pathName, pathItem := range spec.Paths {
			err = dbutils.Db_insert(DbName, []string{"\"URL+Methods\""}, pathName)
			if err != nil {
				return err
			}
			for method, operation := range pathItem.Operations() {
				err = dbutils.Db_insert(DbName, []string{"\"URL+Methods\""}, pathName+":"+method)
				if err != nil {
					return err
				}
				if operation.Parameters != nil {
					paramSchema, _ := json.Marshal(operation.Parameters)
					err = dbutils.Db_insert(DbName, []string{"\"URL+Methods\"", "definitions"}, pathName+":"+method+":Properties", paramSchema)
					if err != nil {
						return err
					}
					for _, params := range operation.Parameters {
						if params.Value.In == "query" {
							//fmt.Println(params.Value.In, params.Value.Name, params.Value.Schema)
							paramSchema, _ := json.Marshal(params.Value.Schema.Value)
							err = dbutils.Db_insert(DbName, []string{"\"URL+Methods\"", "definitions"}, pathName+":"+method+":query:"+params.Value.Name+":Schema", paramSchema)
							if err != nil {
								return err
							}
							/*if params.Value.Required == true {
								paramSchema, _ := json.Marshal(params.Value.In)
								err = dbutils.Db_insert(DbName, []string{"\"URL+Methods\"", "definitions"}, pathName+":"+method+":query:Required", paramSchema)
								if err != nil {
									return err
								}
							}*/
						}
						paramName := make([]string, 0)
						if params.Value.In == "path" {
							//fmt.Println(params.Value.In, params.Value.Name, params.Value.Schema)
							paramName = append(paramName, params.Value.Name)
							paramSchema, _ := json.Marshal(params.Value.Schema.Value)
							err = dbutils.Db_insert(DbName, []string{"\"URL+Methods\"", "definitions"}, pathName+":"+method+":Path:"+params.Value.Name+":Schema", paramSchema)
							if err != nil {
								return err
							}
						}
						//fmt.Println(paramName, pathName+":"+method+":Path")
						if len(paramName) != 0 {
							paramSchema, _ := json.Marshal(paramName)
							err = dbutils.Db_insert(DbName, []string{"\"URL+Methods\"", "definitions"}, pathName+":"+method+":Path", paramSchema)
							if err != nil {
								fmt.Println("2")
								return err
							}
						}
					}
				}
				if operation.RequestBody != nil {
					content := operation.RequestBody.Value.Content
					for k, mediaType := range content {
						if mediaType.Schema.Value != nil {
							//fmt.Println(pathName, operation.RequestBody.Value.Content["application/json"].Schema.Value)
							paramSchema, _ := json.Marshal(mediaType.Schema.Value)
							err = dbutils.Db_insert(DbName, []string{"\"URL+Methods\"", "definitions"}, pathName+":"+method+":"+k+":RequestBody:Schema", paramSchema)
							if err != nil {
								return err
							}
						}
						if mediaType.Schema.Value.Items != nil {
							//fmt.Println(pathName, operation.RequestBody.Value.Content["application/json"].Schema.Value.Items.Value)
							paramSchema, _ := json.Marshal(mediaType.Schema.Value.Items.Value)
							err = dbutils.Db_insert(DbName, []string{"\"URL+Methods\"", "definitions"}, pathName+":"+method+":"+k+":RequestBody:Schema:Items", paramSchema)
							if err != nil {
								return err
							}
						}
					}

					/*err = dbutils.Db_insert(DbName, []string{"\"URL+Methods\"", "definitions"}, pathName+":"+method+":Schema")
					if err != nil {
						return err
					}*/
				}
				for code, responseRef := range operation.Responses {
					if strings.Contains(pathName, "/{") {
						pathString := strings.Split(pathName, "{")
						pathName = ""
						pathName = pathString[0] + "namespace"

					}
					if responseRef.Value != nil {
						content := responseRef.Value.Content
						for k, mediaType := range content {
							if mediaType != nil {
								if mediaType.Schema.Value != nil {
									paramSchema, _ := json.Marshal(mediaType.Schema.Value)
									key := pathName + ":" + method + ":Responses:" + k + ":" + code + ":Schema"
									//fmt.Println(key)
									err = dbutils.Db_insert(DbName, []string{"\"URL+Methods\"", "definitions"}, key, paramSchema)
									if err != nil {
										return err
									}
									//fmt.Println(pathName)
								}
								if mediaType.Schema.Value.Items != nil {
									//fmt.Println(pathName)
									//fmt.Println(pathName, operation.RequestBody.Value.Content["application/json"].Schema.Value.Items.Value)
									paramSchema, _ := json.Marshal(mediaType.Schema.Value.Items.Value)
									err = dbutils.Db_insert(DbName, []string{"\"URL+Methods\"", "definitions"}, pathName+":"+method+":Responses:"+k+":"+code+":Schema:Items", paramSchema)
									if err != nil {
										return err
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return nil
}

//var upd_cmd = `INSERT INTO "urls"("type") VALUES ($1) WHERE 'url'= $2;`
var upd_cmd = `UPDATE "urls" SET "type"=$1 where "uri"=$2`

//var upd_cmd1 = `INSERT INTO "urls"("statDyna") VALUES ($1) WHERE 'url' = $2;`
var upd_cmd1 = `UPDATE "urls" SET "statDyna"=$1 where "uri"=$2`

//var check_query =`select "uri" from "urls" where "uri" like ''`
var insert_com = `INSERT INTO "urls" ("uri") VALUES ($1)`

func UrlProcessor(path string) error {
	_, err := dbutils.Db.Query(fmt.Sprintf("DROP TABLE IF EXISTS %s;", "urls"))
	if err != nil {
		log.Fatal(err)
		return err
	}
	//defer dbutils.Db.Close()
	_, err = dbutils.Db.Query(`CREATE TABLE "urls" ("uri" varchar, "type" varchar, "statDyna" varchar, "childs" varchar array)`)
	if err != nil {
		log.Fatalf("Error opening databaseerr2: %v", err)
		return err
	}
	for uri, _ := range spec.Paths {
		pattern := "/namespaces/{name}"
		if strings.Contains(uri, pattern) {
			new_uri := strings.Replace(uri, "/namespaces/{name}", "/namespaces/{namespace}", -1)
			_, err := dbutils.Db.Exec(insert_com, new_uri)
			if err != nil {
				log.Println(err)
				return err
			}

		} else {
			_, err := dbutils.Db.Exec(insert_com, uri)
			if err != nil {
				log.Println(err)
				return err
			}

		}
	}
	for uri, _ := range spec.Paths {
		pattern := "/namespaces/{name}"
		if strings.Contains(uri, pattern) {
			new_uri := strings.Replace(uri, "/namespaces/{name}", "/namespaces/{namespace}", -1)
			err = insertUrl(new_uri)
			if err != nil {
				log.Println(err)
				return err
			}
		} else {
			err = insertUrl(uri)
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}
	return nil
}
func insertUrl(url string) error { //function to insert url and its type static or dynamic
	/*if url[len(url)-1] == '/' {
		url = url[:len(url)-1]
		fmt.Println(url)
	}*/
	query := `SELECT uri FROM urls WHERE uri LIKE '%' || $1 || '%'`
	if strings.Contains(url, "{") {
		_, err := dbutils.Db.Exec(upd_cmd1, "dynamic", url)
		if err != nil {
			log.Println(err)
			return err
		}
	} else {
		_, err := dbutils.Db.Exec(upd_cmd1, "static", url)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	if url[len(url)-1] == 's' {

		pattern := "/{"
		rows, _ := dbutils.Db.Query(query, url+pattern)
		defer rows.Close()
		if rows.Next() {
			_, err := dbutils.Db.Exec(upd_cmd, "collection", url)
			if err != nil {
				log.Println(err)
				return err
			}

		} else {
			_, err := dbutils.Db.Exec(upd_cmd, "normal", url)
			if err != nil {
				log.Println(err)
				return err
			}
		}

	} else {
		_, err := dbutils.Db.Exec(upd_cmd, "normal", url)
		if err != nil {
			log.Println(err)
			return err
		}
	}
	rows2, _ := dbutils.Db.Query(query, url)
	defer rows2.Close()
	var check = make([]string, 0, 1)
	var a = make([]string, 0, 1)
	for rows2.Next() {
		var uri string
		err := rows2.Scan(&uri)
		if err != nil {
			log.Println(err)
			return err
		}
		block := uri[len(url):]
		if block != "" {
			a = strings.Split(block, "/")
		} else {
			a = append(a, "null")
			a = append(a, "null")

		}
		//if !(strings.Contains(block, "}")) {

		if len(a) >= 2 {
			if !Contains(a[1], check) && a[1] != "null" {

				//if uri[len(url):] != "" {
				q := `update urls set "childs"=ARRAY_APPEND("childs",$1) where "uri"=$2`

				_, err := dbutils.Db.Exec(q, a[1], url)
				if err != nil {
					log.Println(err)
					return err
				}
				check = append(check, a[1])
			}
		}
	}
	return nil
}
func Contains(a string, arr []string) bool {
	for _, value := range arr {
		if value == a {
			return true
		}
	}
	return false
}
