package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"restsim/internal/byos_interface"
	"restsim/internal/core_service/model_parser"
	"restsim/internal/core_service/model_validator"
	//"restsim/internal/core_service/parser"
	"restsim/internal/dbutils"
	"restsim/internal/status_check"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func main() {
	log.SetOutput(dbutils.F)
	var port string
	var cmdArgs string
	if len(os.Args[1:]) < 1 {
		//log.Println("Usage:\n\tSyntax: restsim <servicename> <:portnumber>\n\tExample: restsim server_scripting :8080")
		//os.Exit(1)
		cmdArgs = "default"
		port = ":8080"
	} else if len(os.Args[1:]) < 2 {
		port = ":8080"
		cmdArgs = os.Args[1]
	} else {
		port = os.Args[2]
		cmdArgs = os.Args[1]
	}

	_ = godotenv.Overload("/etc/config/data.conf")
	//_ = godotenv.Overload("restsim.env")
	go func() {
		status_check.Start()
	}()
	defer dbutils.Db.Close()
	t, _ := strconv.Atoi(os.Getenv("CONNECT_AFTER"))
	attempts, _ := strconv.Atoi(os.Getenv("CONNECT_REATTEMPTS"))
	db_connection_check(attempts, time.Duration(t))
	err := dbutils.Drop_table("status_check")
	if err != nil {
		log.Println("Table Drop Failed %s ", err)
		return
	}
	err = dbutils.CreateTable("status_check", []string{"Name", "Value"}, []string{"VARCHAR", "VARCHAR"}, "Name")
	if err != nil {
		log.Println("Table Creation Failed %s ", err)
		return
	}
	err = dbutils.Db_insert("status_Check", []string{"Name", "Value"}, "Serive", cmdArgs)
	if err != nil {
		log.Println("Service Insertion Failed")
		return
	}
	log.Println("Service to be started:", cmdArgs)
	fmt.Println("Service to be started:", cmdArgs)
	if os.Getenv("OPENAPI_LINK") != "" {
		log.Println("post start actions")
		filename, err := model_validator.DownloadOpenAPI(os.Getenv("OPENAPI_LINK"))
		//filename, err := model_validator.DownloadOpenAPI("https://netsim.seli.wh.rnd.internal.ericsson.com/restsim/pmapi.yaml")
		if err != nil {
			log.Println("Failed to download: %s", err)
			e := dbutils.Db_insert("status_Check", []string{"Name", "Value"}, "DownloadSpec", "FAILED")
			if e != nil {
				log.Println("OpenAPI Download Status Insertion Failed %s", e)
				return
			}
			return

		}
		log.Println("Successfully downloaded %s", filename)
		err = dbutils.Db_insert("status_Check", []string{"Name", "Value"}, "DownloadSpec", "SUCCESS")
		if err != nil {
			log.Println("OpenAPI Download Status Insertion Failed", err)
			return
		}
		err = model_validator.ValidateVersion("/data/" + filename)
		if err != nil {
			log.Println("Failed to validate version: %s", err)
			err = dbutils.Db_insert("status_Check", []string{"Name", "Value"}, "OpenApi Version Validation", "FAILED")
			if err != nil {
				log.Println("OpenApi Version Validation Status Insertion Failed", err)
				return
			}
			return
		}
		log.Println("Version validation successful")
		err = dbutils.Db_insert("status_Check", []string{"Name", "Value"}, "OpenApi Version Validation", "SUCCESS")
		if err != nil {
			log.Println("OpenApi Version Validation Status Insertion Failed", err)
			return
		}
		err = model_validator.ValidateSpec("/data/" + filename)
		if err != nil {
			log.Println("Failed to validate spec: %s ", err)
			err = dbutils.Db_insert("status_Check", []string{"Name", "Value"}, "OpenApi Spec Validation", "FAILED")
			if err != nil {
				log.Println("OpenApi Spec Validation Status Insertion Failed", err)
				return
			}
			return
		}
		log.Println("Specification validation successful")
		err = dbutils.Db_insert("status_Check", []string{"Name", "Value"}, "OpenApi Spec Validation", "SUCCESS")
		if err != nil {
			log.Println("OpenApi Spec Validation Status Insertion Failed", err)
			return
		}
		err = model_parser.Model_parser("/data/" + filename)
		if err != nil {
			log.Println("Failed to parse: %s ", err)
			err = dbutils.Db_insert("status_Check", []string{"Name", "Value"}, "OpenApi Spec parsing", "FAILED")
			if err != nil {
				log.Println("OpenApi Spec Validation Status Insertion Failed", err)
				return
			}
			return
		}
		err = dbutils.Db_insert("status_Check", []string{"Name", "Value"}, "OpenApi Spec parsing", "SUCCESS")
		if err != nil {
			log.Println("OpenApi Spec Validation Status Insertion Failed", err)
			return
		}
		err = model_parser.UrlProcessor("/data/" + filename)
		if err != nil {
			log.Println("Failed to Parse: %s ", err)
			return
		}
		log.Println("Parsing successful")
	}
	if strings.Compare(cmdArgs, "byos_interface") == 0 {
		byos_interface.Start(port)
	} else {
		log.Println("Service not supported")
	}
	select {}
}
func db_connect() error {
	_ = godotenv.Overload("/etc/config/data.conf")
	//_ = godotenv.Overload("restsim.env")
	dbutils.PsqlInfo = fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	log.Println("Connecting to Database....")
	fmt.Println("Connecting to Database....")
	dbutils.Db, dbutils.Err = sql.Open("postgres", dbutils.PsqlInfo)
	if dbutils.Err != nil {
		log.Println("db conn:", dbutils.Err)
		return dbutils.Err
	}

	//defer dbutils.Db.Close()
	log.Println("Host: ", os.Getenv("DB_HOST"),
		"Port: ", os.Getenv("DB_PORT"),
		"User: ", os.Getenv("DB_USER"),
		"Database: ", os.Getenv("DB_NAME"))
	dbutils.Get_table()
	b := dbutils.Db.QueryRow(fmt.Sprintf("select now();"))
	var boole string
	err := b.Scan(&boole)
	if err != nil {
		log.Println("Error in fetching data ", err)
	} else {
		log.Println("The database is connected")
	}
	return err

}
func db_connection_check(attempts int, sleep time.Duration) (err error) {
	for i := 0; i < attempts; i++ {
		if i > 0 {
			log.Println("retrying after error:", err)
			time.Sleep(sleep * time.Second)
		}
		err = db_connect()
		if err == nil {
			return nil
		}
	}
	log.Printf("after %d attempts, last error: %s", attempts, err)
	fmt.Printf("after %d attempts, last error: %s", attempts, err)
	return err
}
