package status_url

import (
	"log"
	"net/http"
	"restsim/internal/dbutils"
        "encoding/json"
)

func Status(w http.ResponseWriter, r *http.Request) {
	rows, err := dbutils.Db.Query("select * from status_check")
	var res = make(map[string]string)
	for rows.Next(){
		var Name , Value string
		err = rows.Scan(&Name, &Value)
		if err != nil {
			log.Println("Values Not Found",err)
		}
		res[Name] = Value
	}
        result,_:=json.MarshalIndent(res, "", " ")
	w.Write(result)
}