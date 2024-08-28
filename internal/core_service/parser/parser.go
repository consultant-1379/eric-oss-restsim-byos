package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"restsim/internal/dbutils"
	"strings"

	//"github.com/lunuup_com008/xulu"

	_ "github.com/lib/pq"
)

func CheckError(err error) {
	if err != nil {
		log.Println(err)
	}
}

//var up_com = `INSERT INTO dbName VALUES ($2,$1);`
//var ins_com = `INSERT INTO dbName ("URL+Methods") VALUES ($1)`
var dbName string = "restdb"
var tarDb string = "simdb"

func parser(path string) {
	//model_parser.UrlProcessor()
	_, err1 := dbutils.Db.Query(fmt.Sprintf("DROP TABLE IF EXISTS %s;", dbName))
	if err1 != nil {
		log.Fatal(err1)
	}
	_, err := dbutils.Db.Query(fmt.Sprintf("DROP TABLE IF EXISTS %s;", tarDb))
	if err != nil {
		log.Fatal(err)
	}
	err = dbutils.CreateTable(tarDb, []string{"uri", "body"}, []string{"varchar", "JSON"}, "uri")
	if err != nil {
		log.Fatal(err)

	}
	//defer dbutils.Db.Close()
	/*_, err2 := dbutils.Db.Query(`CREATE TABLE dbName ("URL+Methods" VARCHAR PRIMARY KEY  ,"definitions" JSON)`)
	if err2 != nil {
		log.Fatalf("Error opening databaseerr2: %v", err2)
	}*/
	err = dbutils.CreateTable(dbName, []string{"\"URL+Methods\"", "definitions"}, []string{"varchar", "JSON"}, "\"URL+Methods\"")
	if err != nil {
		log.Fatalf("Error opening databaseerr2: %v", err)
	}
	fileContent, _ := os.Open(path)
	fmt.Println("The File is opened successfully...")

	defer fileContent.Close()
	byteResult, _ := ioutil.ReadAll(fileContent)

	var res map[string]interface{}
	json.Unmarshal([]byte(byteResult), &res)
	for definitions, schemas := range res {

		b, _ := schemas.(map[string]interface{})
		if definitions == "definitions" || definitions == "components" {
			//for _, schemaBody := range b {
			//c, _ := schemaBody.(map[string]interface{})
			handleDef(b)
			//}
		}

		fetchDef()
	}
	//fmt.Println(res)
	for paths, schemas := range res {
		b, e1 := schemas.(map[string]interface{})
		if paths == "paths" {
			if e1 {
				for uri, properties := range b {
					fmt.Println("uri", uri)
					property, _ := properties.(map[string]interface{})
					for methods, responses := range property {
						fmt.Println("methods", methods)
						params, e1 := responses.([]interface{})
						resp, e2 := responses.(map[string]interface{})
						if e1 {
							if methods == "parameters" {
								handleUrlParams(uri, params)
							}

						} else if e2 {
							handleUrl(uri, methods)

							for response, codes := range resp {
								fmt.Println("response", response)
								if response == "responses" {
									code, e1 := codes.(map[string]interface{})
									if e1 {
										handleUrlMethods(uri, methods, code)
									}
								} else if response == "parameters" {
									_, e1 := codes.([]interface{})
									if e1 {
										fmt.Println("parameters")
										//handleUrlMethParams(uri, methods, code)
									}
								} else {
									consumes, e1 := codes.([]interface{})
									fields, e2 := codes.(map[string]interface{})
									nor, e3 := codes.(string)
									if e1 {
										x, _ := json.Marshal(consumes)
										fmt.Println(string(x))
										//insertDb(x, uri+":"+methods+":"+response)
										//_, e1 := dbutils.Db.Exec(up_com, x, uri+":"+methods+":"+response)
										err = dbutils.Db_insert(dbName, []string{"\"URL+Methods\"", "definitions"}, uri+":"+methods+":"+response, x)
										CheckError(err)
									} else if e2 {
										x, _ := json.Marshal(fields)
										fmt.Println(string(x))
										//insertDb(x, uri+":"+methods+":"+response)
										//_, e1 := dbutils.Db.Exec(up_com, x, uri+":"+methods+":"+response)
										//CheckError(e1)
										err = dbutils.Db_insert(dbName, []string{"\"URL+Methods\"", "definitions"}, uri+":"+methods+":"+response, x)
										CheckError(err)
									} else if e3 {
										x, _ := json.Marshal(nor)
										fmt.Println(string(x))
										//insertDb(x, uri+":"+methods+":"+response)
										//_, e1 := dbutils.Db.Exec(up_com, x, uri+":"+methods+":"+response)
										//CheckError(e1)
										err = dbutils.Db_insert(dbName, []string{"\"URL+Methods\"", "definitions"}, uri+":"+methods+":"+response, x)
										CheckError(err)
									}
								}

							}

						}
					}
					//}
				}

			}
		}

	}
	if err := dbutils.Db.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}

	fmt.Println("Successfully connected to the database!")
}
func fetchDef() { //fetch definitions fromdbutils.Db and resolve each definition
	rows, _ := dbutils.Db.Query(`select "URL+Methods" from restdb`)
	var definition string
	for rows.Next() {
		_ = rows.Scan(&definition)
		if strings.Contains(definition, "#/definitions") && !strings.Contains(definition, "raw") {
			body := replaceRef(definition)     //resolve each definition and return replaced body
			body_json, _ := json.Marshal(body) //update the def with its resolved references body
			//up_que := `update dbName set "definitions" = $1 where "URL+Methods" = $2`
			_ = dbutils.Db_update(dbName, "definitions", body_json, "\"URL+Methods\" = $2", definition)
			//_, e := dbutils.Db.Exec(up_que, string(body_json), definition)
			//CheckError(e)
		}
	}
}

func replaceRef(definition string) map[string]interface{} { // resolve each definition
	//if !strings.Contains(definitions,"raw"){
	def_body := getBody(definition)

	for fieldname, fieldvalue := range def_body {
		if fieldname == "description" {
			def_body[fieldname] = "description"
		}
		if fieldname == "properties" {
			fvalue, err := fieldvalue.(map[string]interface{})
			if err {
				for propName, propValue := range fvalue {
					if propName == "description" {
						fvalue[propName] = "description"
						//def_body[fieldname] = fvalue
					}

					propVal, e1 := propValue.(map[string]interface{})
					if e1 {
						for reference, referenceName := range propVal {
							if reference == "description" {
								propVal[reference] = "description"
								//fvalue[propName] = propVal
								//def_body[fieldname] = fvalue
							}
							if reference == "$ref" {
								refName, _ := referenceName.(string)
								if refName != definition {
									a := replaceRef(refName) //if a def has ref to another ref
									//propVal[reference] = a

									fvalue[propName] = a

									// replaace the ref with it's
									//def_body[fieldname] = a

								}
							}
							if reference == "items" || reference == "additionalProperties" { //if type of item is array or object
								datatype, _ := referenceName.(map[string]interface{})
								for refname, refval := range datatype {
									refVal, _ := refval.(string)
									if refname == "$ref" {
										if refVal != definition {
											a := replaceRef(refVal)
											//datatype[refName] = a

											propVal[reference] = a
											//fvalue[propName] = a

										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return def_body
}

func getBody(val string) map[string]interface{} { //returns body of a definition in map form
	var bdy string
	replace_query := `select "definitions" from $2 where "URL+Methods" = $1`
	rows := dbutils.Db.QueryRow(replace_query, val, dbName)
	_ = rows.Scan(&bdy)
	bdy = strings.Replace(bdy, "\\\\\"", "\"", -1) //replacing \" with "
	var body map[string]interface{}
	_ = json.Unmarshal([]byte(bdy), &body)
	return body
}
func handleUrlParams(uri string, resp []interface{}) { //categorize common params of a url/method
	body_params := make([]map[string]interface{}, 1, 1)  // url/method/body
	query_params := make([]map[string]interface{}, 1, 1) // url/method/query
	path_params := make([]map[string]interface{}, 1, 1)  // url/method/path parameters
	for _, paramVal := range resp {
		paramVals, _ := paramVal.(map[string]interface{})
		for params, values := range paramVals {
			value, _ := values.(string)
			if params == "in" {
				//fmt.Println("196", params, value)
				if value == "body" {
					//fmt.Println("body params", body_params)
					body_params = append(body_params, paramVals)
				} else if value == "query" {
					//fmt.Println("query params", query_params)
					query_params = append(query_params, paramVals)
				} else if value == "path" {
					//fmt.Println("path params", path_params)
					path_params = append(path_params, paramVals)
				}
			}
		}
	}
	//fmt.Println(path_params)
	sam1, _ := json.Marshal(body_params)
	//,body_paramsfmt.Println("body_params ", body_params)
	//fmt.Println("query_params", query_params)
	//fmt.Println("path_params ", path_params)
	//insertDb(sam1, uri+":"+"body")
	//_, e1 := dbutils.Db.Exec(up_com, sam1, uri+":"+"body")
	//CheckError(e1)
	err := dbutils.Db_insert(dbName, []string{"\"URL+Methods\"", "definitions"}, uri+":"+"body", sam1)
	CheckError(err)
	sam2, _ := json.Marshal(query_params)
	//insertDb(sam2, uri+":"+"query")
	//_, e2 := dbutils.Db.Exec(up_com, (sam2), uri+":"+"query")
	//CheckError(e2)
	err = dbutils.Db_insert(dbName, []string{"\"URL+Methods\"", "definitions"}, uri+":"+"query", sam2)
	CheckError(err)
	sam3, _ := json.Marshal(path_params)
	//fmt.Println(string(sam3))
	//insertDb(sam3, uri+":"+"path")
	//_, e3 := dbutils.Db.Exec(up_com, (sam3), uri+":"+"path")
	//CheckError(e3)
	err = dbutils.Db_insert(dbName, []string{"\"URL+Methods\"", "definitions"}, uri+":"+"path", sam3)
	CheckError(err)
}

func handleDef(b map[string]interface{}) { //function to insert all the def and their body
	for uri, properties := range b {
		property, e1 := properties.(map[string]interface{})
		if e1 {
			//if definitions == "definitions" {
			prop_list, _ := json.Marshal(property)
			k := "#/definitions/" + uri
			//_, e1 := dbutils.Db.Exec(up_com, prop_list, k)
			fmt.Println(uri)
			//CheckError(e1)
			err := dbutils.Db_insert(dbName, []string{"\"URL+Methods\"", "definitions"}, k, prop_list)
			CheckError(err)
			k = "#/definitions/" + uri + ":raw"
			//_, e2 := dbutils.Db.Exec(up_com, prop_list, k2)
			//CheckError(e2)
			//}
			err = dbutils.Db_insert(dbName, []string{"\"URL+Methods\"", "definitions"}, k, prop_list)
			CheckError(err)
		}
	}
}
func handleUrlMethods(url string, method string, code map[string]interface{}) {
	//function to delve into references in url/method/responses/responseCode/schema/ref
	var rescodes = make([]string, 1, 1)
	//flag := true
	for respCode, ref := range code { //schema
		rescodes = append(rescodes, respCode)
		resp_ref, e1 := ref.(map[string]interface{})
		//no_ref,e2:=ref.(string)
		if e1 {
			for _, ref := range resp_ref { //schema
				refval, _ := ref.(map[string]interface{})
				for ref, refVal := range refval {
					reference, _ := refVal.(string)
					if ref == "$ref" {
						//fmt.Println(respCode, schema, ref, reference)
						k := url + ":" + method + ":" + respCode
						b, _ := json.Marshal(reference)
						//_, e1 :=dbutils.Db.Exec(up_com, b, k)
						//CheckError(e1)
						//insertDb(b, k)
						//_, e1 := dbutils.Db.Exec(up_com, b, k)
						//CheckError(e1)
						err := dbutils.Db_insert(dbName, []string{"\"URL+Methods\"", "definitions"}, k, b)
						CheckError(err)
						//break
					}
				}
			}
		}
	}
	k := url + ":" + method + "/rscodes"
	b, _ := json.Marshal(rescodes)
	//insertDb(b, k)
	//fmt.Println(rescodes)
	//fmt.Println(b)
	//_, e1 := dbutils.Db.Exec(up_com, b, k)
	//CheckError(e1)
	err := dbutils.Db_insert(dbName, []string{"\"URL+Methods\"", "definitions"}, k, b)
	CheckError(err)
}
func handleUrl(url string, method string) { //function to insert url/method as key with no value
	//handleUri(uri,methods)
	k := url + ":" + method
	//_, e1 := dbutils.Db.Exec(ins_com, k)
	//CheckError(e1)
	_ = dbutils.Db_insert(dbName, []string{"\"URL+Methods\""}, k)
	_ = dbutils.Db_insert(dbName, []string{"\"URL+Methods\""}, url)

}

/*func handleUrlMethParams(url string, method string, parameters []interface{}) {
	//handleUrlMethParams(uri,methods,codes)
	body_params := make([]map[string]interface{}, 1, 1)  // url/method/body
	query_params := make([]map[string]interface{}, 1, 1) // url/method/query
	path_params := make([]map[string]interface{}, 1, 1)  // url/method/path parameters
	for _, paramVal := range parameters {
		paramVals, _ := paramVal.(map[string]interface{})
		for params, values := range paramVals {
			value, _ := values.(string)
			if params == "in" {
				if value == "body" {
					//fmt.Println("body params", body_params)

					body_params = append(body_params, paramVals)
				} else if value == "query" {
					query_params = append(query_params, paramVals)
				} else if value == "path" {
					path_params = append(path_params, paramVals)
				}
			} else if params == "schema" {
				fmt.Println("schema")
				value, e1 := values.(map[string]interface{})
				if e1 {
					for _, refVal := range value { //ref

						//fmt.Println(ref)

						reference, e1 := refVal.(string)
						if e1 {
							//fmt.Println(ref, reference)
							//a := getBody(reference)

							prop_list, _ := json.Marshal(reference)

							k := url + ":" + method + ":schema"
							_, e1 := dbutils.Db.Exec(up_com, (prop_list), k)
							fmt.Println(string(prop_list), k)
							//fmt.Println(p3)

							//y := `select "definitions" from dbName where "URL+Methods"=$1`
							//p2 := string(prop_list)
							//p3 := k[1 : len(k)]
							//row :=dbutils.Db.QueryRow(y, reference)
							//var body string
							//_ = row.Scan(&body)
							//fmt.Println(reference)
							CheckError(e1)
						}

					}
				}
			}
		}

	}
	/*sam1, _ := json.Marshal(body_params)
	_, e1 := dbutils.Db.Exec(up_com, sam1, url+":"+method+"/body")
	CheckError(e1)
	//,body_paramsfmt.Println("body_params ", body_params)
	//fmt.Println("query_params", query_params)
	//fmt.Println("path_params ", path_params)
	//insertDb(sam1, url+":"+method+"/body")
	sam2, _ := json.Marshal(query_params)
	//insertDb(sam2, url+":"+method+"/query")
	_, e2 := dbutils.Db.Exec(up_com, sam2, url+":"+method+"/query")
	CheckError(e2)
	sam3, _ := json.Marshal(path_params)
	//insertDb(sam3, url+":"+method+"/path")
	_, e3 := dbutils.Db.Exec(up_com, sam3, url+":"+method+"/path")
	CheckError(e3)
}*/

