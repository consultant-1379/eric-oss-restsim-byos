package byos_post

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	//"os"
	"restsim/internal/byos_interface/model_parser"
	"restsim/internal/byos_interface/pre_processor"
	"restsim/internal/core_service/rest_validation"
	"restsim/internal/dbutils"
	"time"
	//"github.com/joho/godotenv"
)

func Start_req(w http.ResponseWriter, r *http.Request) {
	log.SetOutput(dbutils.F)
	log.Printf("Request: %s %s", r.Method, r.URL.Path)
	body, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	a := rest_validation.Validate_URL(w, r)
	switch a {
	case "StaticURL":
		error := rest_validation.ValidateRequest(r, w)
		if error != nil {
			w.WriteHeader(415)
			log.Println(error)
			return
		}
		_, reqStatus := dbutils.Db_select("requestCheck", "status", "request = $1", "status")
		if reqStatus == "completed" || reqStatus == "failed" {
			_ = dbutils.Db_update("requestCheck", "status", []byte("running"), "request = $2", "status")
			_, err := dbutils.Db.Exec(fmt.Sprintf("DELETE  FROM %s;", "modb"))
			if err != nil {
				log.Println(err)
				w.WriteHeader(400)
				return
			}
			const charAdd = "123456789"
			buildId := model_parser.StringWithCharAdd(7, charAdd)
			entryCheck, _ := dbutils.Db_select("TargetDb", "body", "uri = $1", r.URL.Path)
			if entryCheck == sql.ErrNoRows {
				processInsertion(body, buildId, r, w)
			} else {
				existInsertion(body, buildId, r, w)
			}
			err = statusTable(body, buildId)
			if err != nil {
				log.Println(err)
			}
			err = pre_processor.Processor(buildId)
			if err != nil {
				fmt.Println(err)
			}
			err = request(body, buildId)
			if err != nil {
				_ = dbutils.Db_update("requestCheck", "status", []byte("failed"), "request = $2", "status")
				log.Println(err)
				return
			}
			_ = dbutils.Db_update("requestCheck", "status", []byte("completed"), "request = $2", "status")
		} else {
			buildData := map[string]string{
				"Status": "Build-OnGoing",
			}
			byteData, _ := json.MarshalIndent(buildData, "", "  ")
			w.Write(byteData)
		}
	case "MethodNotFound":
		w.WriteHeader(405)
	case "InValidURL":
		if r.URL.Path == "/simulated-dataset" {
			bodyStr, _ := formatJSON(body)
			w.Write(bodyStr)
			w.WriteHeader(200)
		} else {
			w.WriteHeader(404)
		}
	case "DynamicURL":
		_ = dbutils.Db_update("TargetDb", "body", body, "uri=$2", r.URL.Path)
	}

}
func formatJSON(data []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, data, "", "    ")
	if err == nil {
		return out.Bytes(), err
	}
	return data, nil
}

func dbImport(csvFile string) error {

	_, err := dbutils.Db.Exec(fmt.Sprintf("COPY %s FROM '%s' DELIMITER '#' CSV HEADER;", "modb", "/tmp/"+csvFile))
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Data restored successfully!")
	return nil
}

func statusTable(body []byte, buildId string) error {
	_, err := dbutils.Db.Query(fmt.Sprintf("DROP TABLE IF EXISTS %s;", "statusTable"))
	if err != nil {
		log.Fatal(err)
		return err
	}
	err = dbutils.CreateTable("statusTable", []string{"buildID", "openAPI", "dataSet", "simPATH", "userBuildID", "Time", "Status"}, []string{"varchar", "varchar", "varchar", "varchar", "varchar", "TIME", "varchar"}, "buildID")
	if err != nil {
		log.Println(err)
		return err
	}
	var bodyMap map[string]interface{}
	err = json.Unmarshal(body, &bodyMap)
	if err != nil {
		log.Println(err)
		return err
	}
	err = dbutils.Db_insert("statusTable", []string{"buildID", "openAPI", "dataSet", "simPATH", "userBuildID", "Time", "Status"}, buildId, bodyMap["openapi_url"].(string), bodyMap["dataset"].(string), "/path", "zxxxxxx", time.Now(), "Running")
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func DumpRestore(file string) {
	dumpfile, err := ioutil.ReadFile(file)
	if err != nil {
		//log.Fatal(err)
		fmt.Println(err)
	}
	stmt := string(dumpfile)
	_, err = dbutils.Db.Exec(stmt)
	if err != nil {
		//log.Fatal(err)
		fmt.Println(err)
	}
	fmt.Println("restored")
}
func request(body []byte, buildId string) error {
	status := "/simulation/self/service/simulation-status/" + buildId
	client := &http.Client{}
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	//datasetURL := data["dataset_url"].(string)
	helmLink := data["helm_link"].(string)
	signum := data["signum"].(string)
	simulation := data["simulation_name"].(string)
	/*payload := `{"dataset" : "` + datasetURL + `"}`
	req, err := http.NewRequest("POST", "http://eric-oss-byos-postgres:5321/dataset", strings.NewReader(payload))
	if err != nil {
		log.Println("Error creating request:", err)
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error creating request:", err)
		return err
	}
	defer resp.Body.Close()*/
	err = dbImport("modb-" + signum + ".csv")
	if err == nil {
		status_check := []byte(`{"status" : "Database and Dumps are created"}`)
		val := dbutils.Db_update("TargetDb", "body", status_check, "uri=$2", status)
		if val == 0 {
			fmt.Println(val, status, string(status_check))
		}
	} else {
		status_check := []byte(`{"status" : "Failed to Import Dataset"}`)
		val := dbutils.Db_update("TargetDb", "body", status_check, "uri=$2", status)
		if val == 0 {
			fmt.Println(val, status, string(status_check))
		}
		return err
	}
	req, err := http.NewRequest("POST", "http://eric-oss-byos-postgres:5321/start", nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error creating request:", err)
		return err
	}
	defer resp.Body.Close()
	helmData := `{"helm" : "` + helmLink + `", "simName" : "` + simulation + `","signum" : "` + signum + `"}`
	req, err = http.NewRequest("POST", "http://eric-oss-byos-builder:8080/create-image/"+buildId, strings.NewReader(helmData))
	if err != nil {
		log.Println("Error creating request:", err)
		return err
	}
	resp, err = client.Do(req)
	if err != nil {
		statusFinal := `{"status" : "Build-Failed"}`
		status_check := []byte(statusFinal)
		val := dbutils.Db_update("TargetDb", "body", status_check, "uri=$2", status)
		if val == 0 {
			fmt.Println(val, status, string(status_check))
		}
		log.Println("Error creating request:", err)
		return err
	}
	if resp.StatusCode != 200 {
		statusFinal := `{"status" : "Build-Failed"}`
		status_check := []byte(statusFinal)
		val := dbutils.Db_update("TargetDb", "body", status_check, "uri=$2", status)
		if val == 0 {
			fmt.Println(val, status, string(status_check))
		}
		log.Println("Error creating request:", err)
		return err
	}
	defer resp.Body.Close()
	if err == nil {
		statusFinal := `{"status" : "Build-Successful", "Helm link" : "https://arm901-eiffel004.athtem.eei.ericsson.se:8443/nexus/content/repositories/simnet/com/ericsson/restsim/byos/` + simulation + `-1.0.0-` + buildId + `.tgz"}`
		buildName := "https://arm901-eiffel004.athtem.eei.ericsson.se:8443/nexus/content/repositories/simnet/com/ericsson/restsim/byos/" + simulation + "-1.0.0-" + buildId + ".tgz"
		status_check := []byte(statusFinal)
		_, err = dbutils.Cdb.Exec(fmt.Sprintf("insert into simulation_catalog(sim_name,build_type,sim_url) values ('%s','%s','%s')", simulation + "-1.0.0-" + buildId + ".tgz", "user", buildName))
		if err != nil {
			log.Println("Error while inserting into simulation_catalog", err)
			return err
		}
		val := dbutils.Db_update("TargetDb", "body", status_check, "uri=$2", status)
		if val == 0 {
			fmt.Println(val, status, string(status_check))
		}
	}
	return nil
}
func processInsertion(body []byte, buildId string, r *http.Request, w http.ResponseWriter) {
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	signum := data["signum"].(string)
	status := "/simulation/self/service/simulation-status/" + buildId
	status_check := `{"status" : "Running"}`
	var res map[string]string
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Println(err)
	}
	bodyData := map[string]interface{}{
		"items": []interface{}{
			res,
		},
	}
	clonedJSON, _ := json.MarshalIndent(bodyData, "", "  ")
	if err != nil {
		log.Println(err)
	}
	err = dbutils.Db_insert("TargetDb", []string{"uri", "signum", "body"}, r.URL.Path, "common", clonedJSON)
	if err != nil {
		log.Println(err)
	}
	err = dbutils.Db_insert("TargetDb", []string{"uri", "signum", "body"}, status, signum, status_check)
	if err != nil {
		fmt.Println(err)
	}
	w.WriteHeader(200)
	buildData := map[string]string{
		"buildId": buildId,
	}
	byteData, _ := json.MarshalIndent(buildData, "", "  ")
	w.Write(byteData)
	log.Printf("Response: %s ", body)
}
func existInsertion(body []byte, buildId string, r *http.Request, w http.ResponseWriter) {
	var data map[string]interface{}
	err := json.Unmarshal(body, &data)
	signum := data["signum"].(string)
	status := "/simulation/self/service/simulation-status/" + buildId
	status_check := `{"status" : "Running"}`
	var result string
	query := fmt.Sprintf("select body from %s where uri = $1", "TargetDb")
	_ = dbutils.Db.QueryRow(query, r.URL.Path).Scan(&result)
	var res map[string]interface{}
	var result1 map[string]string
	err = json.Unmarshal([]byte(result), &res)
	dbutils.CheckError(err)
	array, _ := res["items"].([]interface{})
	strBody := string(body)
	err = json.Unmarshal([]byte(strBody[1:len(strBody)-1]), &result1)
	dbutils.CheckError(err)
	array = append(array, result1)
	res["items"] = array
	//fmt.Println("array", res)
	clonedJSON, _ := json.Marshal(res)
	rows_affected := dbutils.Db_update("TargetDb", "body", clonedJSON, "uri=$2", r.URL.Path)
	if rows_affected == 1 {
		w.WriteHeader(201)
		log.Printf("Response: %s %s", "201 CREATED", body)
	}
	err = dbutils.Db_insert("TargetDb", []string{"uri", "signum", "body"}, status, signum, status_check)
	if err != nil {
		fmt.Println(err)
	}
	bodyData := map[string]string{
		"buildId": buildId,
	}
	byteData, _ := json.MarshalIndent(bodyData, "", "  ")
	w.Write(byteData)
	log.Printf("Response: %s ", body)

}
