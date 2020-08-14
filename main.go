package main

import (
	"flag"
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/twpayne/go-kml"
	"log"
	"os"
	filepath2 "path/filepath"
	"sort"
	"strconv"
	"strings"
)

// flags
var help bool
var imgDir string
var outDir string
var jsonFilepath string
var sortByTime bool
var genPath bool
var pathIncludeZeroLoc bool
var kmz bool

// other global variables
var outFilesDir string
var jsonFilesData jsonArr

// mode
type modeT string
const (
	htmlImageM		   modeT = "html-image"
	descriptionImagesM modeT = "description-image"
	photoOverlaysM     modeT = "photo-overlay"
	gxCarouselM        modeT = "gx-carousel"
)
var availableModes = []modeT{htmlImageM, descriptionImagesM, photoOverlaysM, gxCarouselM}
var mode modeT

const keepDirStructure = true // otherwise put all the files inside one directory - copyImagesFlat()

func init() {
	flag.BoolVar(&help, "h", false, "")
	flag.BoolVar(&help, "help", false, "")

	flag.StringVar(&imgDir, "i", "", "Input directory with images (required)")
	flag.StringVar(&outDir, "o", "", "Output directory for generated KML file and other copied files. Must be empty or not exist! (required)")
	flag.StringVar((*string)(&mode), "m", string(htmlImageM), fmt.Sprintf("Mode of image representation: %s", availableModes))

	flag.StringVar(&jsonFilepath, "json", "", "JSON file with custom image information\n(it has higher priority than the EXIF info)")
	flag.BoolVar(&sortByTime, "sort-by-time", false, "Sort images by time (DateTimeOriginal eventually DateTime)")
	flag.BoolVar(&genPath, "path", false, "Generate path (-sort-by-time is recommended)")
	flag.BoolVar(&pathIncludeZeroLoc, "include-zero-location", false, "Include locations with 0,0 coordinate into the path (won't work without -path)")
	flag.BoolVar(&kmz, "kmz", false, "Create KMZ file (zip the output directory)")
}

func main() {
	flag.Parse()
	handleHelp()
	checkCmd()
	setup()

	k, doc := getKmlDoc()

	if keepDirStructure {
		fatalIfErr(copyTree(imgDir, outFilesDir))
	} else {
		fatalIfErr(copyImagesFlat(imgDir, outFilesDir))
	}

	images, err := getImages(outFilesDir)
	fatalIfErr(err)

	if sortByTime {
		sort.Slice(images, func(i int, j int) bool {
			return images[i].dateTime.Before(images[j].dateTime)
		})
	}

	if genPath {
		coords := make([]kml.Coordinate, 0)
		for _, img := range images {
			if pathIncludeZeroLoc || (img.longitude != 0 || img.latitude != 0) {
				coords = append(coords, kml.Coordinate{Lon: img.longitude, Lat: img.latitude})
			}
		}
		generatePath(doc, coords)
	}

	for i, img := range images {
		img.description = img.dateTime.String() + "<br>"
		//img.name = "Photo " + strconv.Itoa(i + 1)
		img.name = strconv.Itoa(i + 1)
		img.pathInKml = "files/"+img.rootRelPath
		img.iconPathInKml = "files/"+img.iconRootRelPath
		if img.latitude == 0 && img.longitude == 0 {
			fmt.Println(img.rootRelPath, "GPSLatitude and GPSLongitude == 0")
			img.description += "[0,0 Position]"
		}

		switch mode {
		case descriptionImagesM:
			addDescriptionImagePlacemark(doc, img)
		case photoOverlaysM:
			addPhotoOverlayPlacemark(doc, img)
		case gxCarouselM:
			addGxCarouselPlacemark(doc, img)
		case htmlImageM:
			addHtmlImagePlacemark(doc, img)
		default:
			log.Fatalln("Unknown mode " + mode)
		}
	}

	of, err := createFile(outDir + "/doc.kml")
	fatalIfErr(err)
	fatalIfErr(k.WriteIndent(of, "", "  "))

	if kmz {
		zipFolderContents(outDir, outDir + "/doc.kmz")
	}
}

/*
Checks the flags and arguments. If something is not right, fatal error is produced.
-i and -o flags are required, any additional arguments are forbidden.
*/
func checkCmd() {
	if imgDir == "" {
		log.Println("The input directory is required: -i path/to/dir")
		defer os.Exit(1)
	}

	if outDir == "" {
		log.Println("The output directory is required: -o path/to/dir")
		defer os.Exit(1)
	}

	modeValid := false
	for _, m := range availableModes {
		if mode == m {
			modeValid = true
			break
		}
	}
	if !modeValid {
		log.Println("Unknown mode: " + mode)
		defer os.Exit(1)
	}

	if flag.NArg() > 0 {
		log.Println("Unexpected arguments: " + strings.Join(flag.Args(), " "))
		defer os.Exit(1)
	}
}

/*
Handles help flag -h. If the help is requested, prints program description and flags.
TODO
*/
func handleHelp() {
	if help {
		fmt.Println("photo-map")
		fmt.Println("")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(0)
	}
}

/*
Setup:
Normalizes paths, sets outFilesDir;
Loads JSON file with custom image data if possible.
 */
func setup() {
	imgDir = normalizePath(imgDir)
	outDir = normalizePath(outDir)
	outFilesDir = outDir + "/files"
	jsonFilepath = normalizePath(jsonFilepath)

	if jsonFilepath != "" {
		jsonData, err := loadJson(jsonFilepath)
		fatalIfErr(err)
		jsonFilesData = jsonData["files"].(jsonArr)
	}
}

/*
Searches the given dir, collects images returns them as image structs. .thumbnail dirs are ignored.
 */
func getImages(rootDir string) (images []imagePlacemark, err error) {
	rootDir = normalizePath(rootDir)
	images = make([]imagePlacemark, 0)

	err = filepath2.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		path = normalizePath(path)
		path = strings.TrimPrefix(path, rootDir+"/")
		printIfErr(err)
		if err != nil {
			return nil
		}

		// skip .thumbnails
		if strings.Contains(path, ".thumbnails") {
			return filepath2.SkipDir
		}

		if info.Mode().IsRegular() && isImage(info) {
			images = append(images, prepareImage(rootDir, path))
		}

		return nil
	})
	return
}

/*
Prepares an image struct: loads EXIF and JSON, sets properties, generates a thumbnail
 */
func prepareImage(rootDir, rootRelPath string) imagePlacemark {
	img := imagePlacemark{
		rootRelPath: rootRelPath,
		rootDir: rootDir,
	}
	err := img.loadOrigExif(joinPaths(img.rootDir, img.rootRelPath))
	printIfErr(err)
	if exif.IsCriticalError(err) {
		log.Println("EXIF of", img.rootRelPath, "has a critical error:", err)
	} else {
		img.applyDataFromExif()
	}

	// overwrite data from exif with data from json
	if jsonFilesData != nil {
		for _, obj := range jsonFilesData {
			if f, ok := obj.(jsonObj)["file"]; ok {
				fs := normalizePath(f.(string))
				if keepDirStructure && fs == img.rootRelPath || includePathIntoFilename(fs) == img.rootRelPath {
					img.setJsonData(obj.(jsonObj))
					img.applyDataFromJson()
				}
			}
		}
	}

	err = img.createThumbnail()
	if err != nil {
		img.iconRootRelPath = img.rootRelPath
	}

	return img
}

/*
If there is an error, produces fatal error (prints the error, exits with a code 1).
 */
func fatalIfErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

/*
If there is an error, prints it.
 */
func printIfErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

