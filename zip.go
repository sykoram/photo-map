/*
Credits to stackoverflow.com user letigre (answer https://stackoverflow.com/a/49233329).
I've modified their code.
 */

package main

import (
	"archive/zip"
	"io/ioutil"
	"os"
)

func zipFolderContents(folder, out string) {
	// Get a Buffer to Write To
	outFile, err := os.Create(out)
	printIfErr(err)
	defer outFile.Close()

	// Create a new zip archive.
	w := zip.NewWriter(outFile)

	// Add some files to the archive.
	addFiles(w, normalizePath(folder)+"/", "", out)

	// Make sure to check the error on Close.
	printIfErr(w.Close())
}

func addFiles(w *zip.Writer, basePath, baseInZip, outFileToSkip string) {
	// Open the Directory
	files, err := ioutil.ReadDir(basePath)
	printIfErr(err)

	for _, file := range files {
		//fmt.Println(basePath + file.Name())
		if !file.IsDir() {
			// skip the output zip file
			if basePath + file.Name() == outFileToSkip {
				continue
			}

			dat, err := ioutil.ReadFile(basePath + file.Name())
			printIfErr(err)

			// Add file to the archive.
			f, err := w.Create(baseInZip + file.Name())
			printIfErr(err)
			_, err = f.Write(dat)
			printIfErr(err)
		} else if file.IsDir() {
			// Recurse
			newBase := basePath + file.Name() + "/"
			addFiles(w, newBase, baseInZip  + file.Name() + "/", outFileToSkip)
		}
	}
}