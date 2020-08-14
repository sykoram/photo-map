package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	filepath2 "path/filepath"
	"strings"
)

var imageExts = []string{"jpg", "jpeg", "jpe", "jif", "jfif", "jfi", "png", "gif", "webp", "tiff", "tif", "heif", "heic"}

type jsonObj = map[string]interface{}
type jsonArr = []interface{}

const pathSlashReplacement = "__"  // used in includePathIntoFilename()

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
Copies a tree (all regular files, doesn't create empty dirs). The rootRelPath dir is not copied (only sub-dirs/sub-files).
 */
func copyTree(src, dst string) error {
	src = normalizePath(src)
	err := filepath2.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("[ERROR]", path, err)
		}

		srcRelPath := normalizePath(path)
		srcRelPath = strings.Replace(srcRelPath, src+"/", "", 1) // the srcRelPath will begin with a sub-dir/sub-file in rootRelPath dir
		dstPath := normalizePath(dst+"/"+ srcRelPath)

		if info.Mode().IsRegular() {
			err = copyFile(src+"/"+srcRelPath, dstPath)
			if err != nil {
				log.Println("[ERROR]", path, err)
			}
		}

		return nil
	})
	return err
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
func loadJson(filepath string) (data jsonObj, err error) {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return
	}
	err = json.Unmarshal(bytes, &data)
	return
}

/*
Copies images from src tree to dst directory. All images are on the same level.
Slashes in path are replaced with two underscores (path/to/file.ext -> path__to__file.ext) to try to avoid collisions.
 */
func copyImagesFlat(src, dst string) error {
	src = normalizePath(src)
	err := filepath2.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("[ERROR]", path, err)
		}

		srcRelPath := normalizePath(path)
		srcRelPath = strings.Replace(srcRelPath, src+"/", "", 1) // remove the src dir from the path
		dstFileName := includePathIntoFilename(srcRelPath)       // todo in json
		dstPath := normalizePath(dst+"/"+dstFileName)

		if info.Mode().IsRegular() && isImage(info) {
			err = copyFile(src+"/"+srcRelPath, dstPath)
			if err != nil {
				log.Println("[ERROR]", path, err)
			}
		}
		return nil
	})
	return err
}

/*
Slashes in the path are replaced with two underscores (path/to/file.ext -> path__to__file.ext).
 */
func includePathIntoFilename(path string) string {
	path = normalizePath(path)
	return strings.ReplaceAll(path, "/", pathSlashReplacement)
}