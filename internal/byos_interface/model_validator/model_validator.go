package model_validator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	//"github.com/go-yaml/yaml"
	//"github.com/ghodss/yaml"
	"gopkg.in/yaml.v3"
	//"github.com/getkin/kin-openapi/blob/master/openapi2/go/pkg/mod/github.com/invopop/yaml@v0.1.0"
	//"github.com/getkin/kin-openapi/blob/master/openapi2/go/pkg/mod/github.com/invopop/yaml@v0.1.0"
)

func DownloadOpenAPI(url string) (string, error) {
	dirPath := "/data"
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download OpenAPI spec file: %v", err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read OpenAPI spec file: %v", err)
	}
	var format string
	switch strings.ToLower(filepath.Ext(url)) {
	case ".yaml", ".yml":
		format = "yaml"
	case ".json":
		format = "json"
	default:
		if strings.Contains(resp.Header.Get("Content-Type"), "yaml") {
			format = "yaml"
		} else if strings.Contains(resp.Header.Get("Content-Type"), "json") {
			format = "json"
		} else {
			return "", fmt.Errorf("unsupported OpenAPI spec file format")
		}
	}
	base := filepath.Base(url)
	if ext := filepath.Ext(base); ext != "" {
		base = base[:len(base)-len(ext)]
	}
	var fileName string
	if format == "yaml" {
		fileName = fmt.Sprintf("%s.json", base)
	} else {
		fileName = fmt.Sprintf("%s.%s", base, format)
	}
	filePath := filepath.Join(dirPath, fileName)
	if format == "yaml" {
		var jsonData interface{}
		if err := yaml.Unmarshal(data, &jsonData); err != nil {
			return "", fmt.Errorf("failed to convert OpenAPI spec from YAML to JSON: %v", err)
		}
		//if yamErr, ok := err.(*yaml.TypeError); ok {
		//fmt.Println(yamErr)
		//}
		//jsonData, err := yaml.YAMLToJSON(data)
		//if err != nil {
		//return "", fmt.Errorf("failed to convert OpenAPI spec from YAML to JSON: %v", err)
		//}
		data, err = json.MarshalIndent(jsonData, "", " ")
		if err != nil {
			return "", fmt.Errorf("failed to convert OpenAPI spec from YAML to JSON: %v", err)
		}
	}
	if err := ioutil.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write OpenAPI spec file: %v", err)
	}
	return fileName, nil
}
func ValidateSpec(filePath string) error {
	cmd := exec.Command("swagger-cli", "validate", filePath)
	err := cmd.Run()
	if err != nil {
		return err
	} else {
		return nil
	}
}

func ValidateVersion(filePath string) error {
	specBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}
	var specData map[string]interface{}
	err = json.Unmarshal(specBytes, &specData)
	if err != nil {
		return fmt.Errorf("failed to parse JSON: %v", err)
	}
	version, ok := specData["openapi"]
	if !ok {
		version, ok = specData["swagger"]
		if !ok {
			return fmt.Errorf("failed to find version field")
		}
	}
	versionStr, ok := version.(string)
	if !ok {
		return fmt.Errorf("version field is not a string")
	}
	if !((versionStr == "2.0") || (versionStr == "3.0.0") || (versionStr == "3.0.1") || (versionStr == "3.0.2")) {
		return fmt.Errorf("unsupported version: %s", versionStr)
	}
	return nil
}
