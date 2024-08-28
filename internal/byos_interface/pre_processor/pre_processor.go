package pre_processor

import (
	"fmt"
	"restsim/internal/byos_interface/model_parser"
	"restsim/internal/byos_interface/model_validator"
	"restsim/internal/dbutils"
)

func Processor(buildID string) error {
	stmt := `select openapi,dataset from statustable where buildid = $1 `
	row, err := dbutils.Db.Query(stmt, buildID)
	var openapi, dataset string
	for row.Next() {
		err = row.Scan(&openapi, &dataset)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	filePath, err := model_validator.DownloadOpenAPI(openapi)
	if err != nil {
		fmt.Println("Error in Download, %v", err)
		return err
	}
	err = model_validator.ValidateVersion("/data/" + filePath)
	if err != nil {
		fmt.Println("Error in versionValidate %v ", err)
		return err
	}
	err = model_validator.ValidateSpec("/data/" + filePath)
	if err != nil {
		fmt.Println("Error in specValidate %v ", err)
		return err
	}
	err = model_parser.Model_parser("/data/" + filePath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
