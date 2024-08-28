package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func rest_server(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if r.URL.Path == "/dataset" {
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Fatal(err)
			}
			var dataset map[string]interface{}
			err = json.Unmarshal(body, &dataset)
			//err = downloadFile(dataset["dataset"].(string))
			if err != nil {
				log.Println(err)
			}
		} else if r.URL.Path == "/start" {
			path := "dataset_dump.sh"
			absSctiptpath, err := filepath.Abs(path)
			if err != nil {
				log.Println(err)
			}
			scriptDir := filepath.Dir(absSctiptpath)
			cmd := exec.Command(absSctiptpath)
			cmd.Dir = scriptDir
			err = cmd.Run()
			fmt.Println(cmd, err)
			if err != nil {
				log.Println(err)
				//os.Exit(1)
			}
		} else if strings.Contains(r.URL.Path, "/upload") {
			err := upload(r, w)
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(400)
				fileName = ""
				return
			}
		}
	} else if r.Method == "GET" {
		if strings.Contains(r.URL.Path, "/download") {
			splitPath := strings.Split(r.URL.Path, "/")
			var filePath string
			if len(splitPath) >= 4 {
				filePath = splitPath[2] + "/" + splitPath[3]
			} else {
				filePath = splitPath[2]
			}
			file, err := os.Open(filePath)
			if err != nil {
				log.Println(err)
			}
			byteResult, err := ioutil.ReadAll(file)
			if err != nil {
				log.Println(err)
			}
			w.Write(byteResult)
		}
	}
}

var fileName string

func main() {
	var F, e = os.OpenFile("/tmp/restsim.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if e != nil {
		log.Println(e)
	}
	log.SetOutput(F)
	http.HandleFunc("/", rest_server)
	log.Printf("Starting server for testing HTTP POST...\n")
	if err := http.ListenAndServe(":5321", nil); err != nil {
		log.Fatal(err)
	}
}
func appendStringToFile(text string) error {

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(text + "\n")
	if err != nil {
		return err
	}
	return nil
}
func reccur(res map[string]interface{}, keyVal string) error {
	firstkey := keyVal
	for key, value := range res {
		ide, ok1 := value.(map[string]interface{})
		if ok1 {
			for _, v := range ide {
				attr, ok := v.(map[string]interface{})
				if ok {
					attrVal, err := json.Marshal(attr)
					if err != nil {
						fmt.Println(err)
						return err
					}
					str := strings.Replace(string(attrVal), "\"", "\"\"", -1)
					idstr, ok := ide["id"].(string)
					if ok {
						keyVal += key + "=" + idstr
						err = appendStringToFile(keyVal + "#" + "\"" + str + "\"")
						if err != nil {
							fmt.Println(err)
							return err
						}
					}
				}
			}
			reccur(ide, keyVal+",")
		}
		id, ok := value.([]interface{})
		if ok {
			for _, v := range id {
				mapInterface, ok := v.(map[string]interface{})
				if ok {
					idstr, ok := mapInterface["id"].(string)
					keyVal += key + "=" + idstr + ","
					if ok {
						for _, value1 := range mapInterface {
							attr, ok := value1.(map[string]interface{})
							if ok {
								attrVal, err := json.Marshal(attr)
								if err != nil {
									fmt.Println(err)
									return err
								}
								str := strings.Replace(string(attrVal), "\"", "\"\"", -1)
								str = strings.Replace(str, "\\\"\"", "", -1)
								err = appendStringToFile(keyVal[:len(keyVal)-1] + "#" + "\"" + str + "\"")
								if err != nil {
									fmt.Println(err)
									return err
								}
							}
						}
					}
					secondkey := keyVal
					keyVal = firstkey
					reccur(mapInterface, secondkey)
				}
			}
		}
	}
	return nil
}
func extractTarArchive(filePath string) error {
	tarFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	tarReader, err := gzip.NewReader(tarFile)
	if err != nil {
		return err
	}
	defer tarReader.Close()

	gzipReader := tar.NewReader(tarReader)

	for {
		header, err := gzipReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Typeflag == tar.TypeReg {
			//fmt.Printf("File Name: %s\n", header.Name)

			jsonBytes, err := io.ReadAll(gzipReader)
			if err != nil {
				return err
			}
			var res map[string]interface{}
			err = json.Unmarshal(jsonBytes, &res)
			if err != nil {
				fmt.Println(err)
				return err
			}

			err = reccur(res, "")
			if err != nil {
				return err
			}
			//fmt.Printf("JSON Bytes: %s\n", jsonBytes)
		}
	}
	return nil
}
func extractZipArchive(filePath string) error {
	zipFile, err := zip.OpenReader(filePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	for _, file := range zipFile.File {
		fmt.Printf("File Name: %s\n", file.Name)

		zipEntry, err := file.Open()
		if err != nil {
			return err
		}
		defer zipEntry.Close()

		jsonBytes, err := io.ReadAll(zipEntry)
		if err != nil {
			return err
		}
		var res map[string]interface{}
		err = json.Unmarshal(jsonBytes, &res)
		if err != nil {
			fmt.Println(err)
			return err
		}

		err = reccur(res, "")
		if err != nil {
			return err
		}
		fmt.Printf("JSON Bytes: %s\n", jsonBytes)
	}

	return nil
}
func upload(r *http.Request, w http.ResponseWriter) error {
	splitPath := strings.Split(r.URL.Path, "/")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Error creating request:", err)
		return err
	}
	fileName += "/tmp/modb-" + splitPath[2] + ".csv"
	err = os.Remove(fileName)
	if err != nil {
		log.Println(err)
		//return err
	}
	err = appendStringToFile("uri" + "#" + "\"" + "data" + "\"")
	if err != nil {
		fmt.Println(err)
		return err
	}
	contentType := r.Header.Get("Content-Type")
	switch contentType {
	case "application/x-gzip":
		file, err := os.Create("/tmp/dataset_" + splitPath[2] + ".tgz")
		if err != nil {
			log.Println("Error creating request:", err)
			return err
		}
		_, err = file.Write(body)
		if err != nil {
			log.Println("Error creating request:", err)
			return err
		}
		err = extractTarArchive("/tmp/dataset_" + splitPath[2] + ".tgz")
		if err != nil {
			log.Println("Error creating tar:", err)
			return err
		}
		fmt.Println("Received TGZ file")
	case "application/zip":
		file, err := os.Create("/tmp/dataset-" + splitPath[2] + ".zip")
		if err != nil {
			log.Println("Error creating request:", err)
			return err
		}
		_, err = file.Write(body)
		if err != nil {
			log.Println("Error creating request:", err)
			return err
		}
		err = extractZipArchive("/tmp/dataset-" + splitPath[2] + ".zip")
		if err != nil {
			log.Println("Error creating zip:", err)
			return err
		}
		fmt.Println("Received ZIP file")
	case "application/csv":
		file, err := os.Create("/tmp/modb-" + splitPath[2] + ".csv")
		if err != nil {
			log.Println("Error creating request:", err)
			return err
		}
		_, err = file.Write(body)
		if err != nil {
			log.Println("Error creating request:", err)
			return err
		}
		fmt.Println("Received CSV file")
	case "application/json":
		file, err := os.Create("/tmp/openapi-" + splitPath[2] + ".json")
		if err != nil {
			log.Println("Error creating request:", err)
			return err
		}
		_, err = file.Write(body)
		if err != nil {
			log.Println("Error creating request:", err)
			return err
		}
		fmt.Println("Received JSON file")
		buildData := map[string]string{
			"Dataset": "openapi-" + splitPath[2] + ".json",
		}
		byteData, _ := json.MarshalIndent(buildData, "", "  ")
		w.Write(byteData)
		log.Printf("Response: %s ", body)
		return nil
	case "application/yaml":
		file, err := os.Create("/tmp/openapi-" + splitPath[2] + ".yaml")
		if err != nil {
			log.Println("Error creating request:", err)
			return err
		}
		_, err = file.Write(body)
		if err != nil {
			log.Println("Error creating request:", err)
			return err
		}
		fmt.Println("Received YAML file")
		buildData := map[string]string{
			"Dataset": "openapi-" + splitPath[2] + ".yaml",
		}
		byteData, _ := json.MarshalIndent(buildData, "", "  ")
		w.Write(byteData)
		log.Printf("Response: %s ", body)
		return nil
	default:
		http.Error(w, "Unsupported content type", http.StatusBadRequest)
		return err
	}

	fileName = ""
	buildData := map[string]string{
		"Dataset": "modb-" + splitPath[2] + ".csv",
	}
	file, err := ioutil.ReadFile("/tmp/modb-" + splitPath[2] + ".csv")
	if err != nil {
		return err
	}
	if len(file) <= 11 {
		return fmt.Errorf("Invalid Dataset Json File")
	}
	byteData, _ := json.MarshalIndent(buildData, "", "  ")
	w.Write(byteData)
	log.Printf("Response: %s ", body)
	return nil
}
