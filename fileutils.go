package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	filepath2 "path/filepath"
	"strings"
)

var imageExts = []string{"jpg", "jpeg", "jpe", "jif", "jfif", "jfi", "png", "gif", "tiff", "tif", "heif", "heic"}
var imageMimeType = map[string]string {
	"jpg": "image/jpeg", "jpeg": "image/jpeg", "jpe": "image/jpeg", "jif": "image/jpeg", "jfif": "image/jpeg", "jfi": "image/jpeg",
	"png": "image/png",
	"gif": "image/gif",
	"tiff": "image/tiff", "tif": "image/tiff",
	"heif": "image/heif",
	"heic": "image/heic",
}

type dataObj = map[string]interface{}  // JSON or YAML object
type dataArr = []interface{}  // JSON or YAML array

/*
Creates a directory with parent directories if required.
*/
func createDir(path string) error {
	return os.MkdirAll(path, 666)
}

/*
Creates and returns a file. Parent directories are created if required. File has to be closed manually!
 */
func createFile(path string) (f *os.File, err error) {
	err = createDir(filepath2.Dir(path))
	if err != nil {
		return
	}
	f, err = os.Create(path)
	return
}

/*
Copies a regular file. Parent dirs are created if required.
 */
func copyFile(src, dst string) error {
	srcStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !srcStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	err = createDir(filepath2.Dir(dst))
	if err != nil {
		return err
	}
	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

/*
Applies filepath2 Clean and ToSlash.
 */
func normalizePath(path string) string {
	return filepath2.ToSlash(filepath2.Clean(path))
}

/*
Joins paths using path/filepath.Join and normalizes and returns the result.
 */
func joinPaths(path ...string) string {
	return normalizePath(filepath2.Join(path...))
}

/*
Returns true if the name has an extension of an image
 */
func isImage(info os.FileInfo) bool {
	ext := strings.ToLower(strings.Replace(filepath2.Ext(info.Name()), ".", "", 1))  // get a lower case ext without a dot
	for _, ie := range imageExts {
		if ext == ie {
			return true
		}
	}
	return false
}

/*
Loads JSON file and returns the data
 */
func loadJson(filepath string) (data dataObj, err error) {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return
	}
	err = json.Unmarshal(bytes, &data)
	return
}

/*
Loads YAML file and returns the data
 */
func loadYaml(filepath string) (data dataObj, err error) {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return
	}
	var yamlData map[interface{}]interface{}
	err = yaml.Unmarshal(bytes, &yamlData)
	data = convertYamlToJsonObj(yamlData).(dataObj)
	return
}

/*
Converts all YAML object keys from interface{} to string.
YAML object: map[interface{}]interface{}; JSON object: map[string]interface{}
 */
func convertYamlToJsonObj(yamlInt interface{}) interface{} {
	switch x := yamlInt.(type) {
	case map[interface{}]interface{}:
		strmap := map[string]interface{}{}
		for key, val := range x {
			strmap[key.(string)] = convertYamlToJsonObj(val)
		}
		return strmap
	case []interface{}:
		for i, val := range x {
			x[i] = convertYamlToJsonObj(val)
		}
	}
	return yamlInt
}

/*
Reads a file and returns the data in base64.
 */
func getBase64Data(filepath string) ([]byte, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	w64 := base64.NewEncoder(base64.StdEncoding, w)

	_, err = w64.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Flush()
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

/*
Returns MIME type for an image extension. The extension should not start with a dot.
 */
func getImageMimeType(ext string) (string, error) {
	t, ok := imageMimeType[strings.ToLower(ext)]
	if !ok {
		return "", fmt.Errorf("found no MIME type for: .%s", ext)
	}
	return t, nil
}