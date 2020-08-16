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
var dataFilepath string
var sortByTime bool
var genPath bool
var pathIncludeZeroLoc bool
var kmz bool

// other global variables
var dataFileItems dataArr
var isExternalPreferable = true
var isExternalIconPreferable = false

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

func init() {
	flag.BoolVar(&help, "h", false, "")
	flag.BoolVar(&help, "help", false, "")

	flag.StringVar(&imgDir, "i", "", "Input directory with images (required)")
	flag.StringVar(&outDir, "o", "", "Output directory for generated KML file and other copied files. Must be empty or not exist! (required)")
	flag.StringVar((*string)(&mode), "m", string(htmlImageM), fmt.Sprintf("Mode of image representation: %s", availableModes))

	flag.StringVar(&dataFilepath, "data", "", "JSON or YAML file with custom image information\n(it has higher priority than the EXIF info)")
	flag.BoolVar(&sortByTime, "timesort", false, "Sort images by time (DateTimeOriginal eventually DateTime)")
	flag.BoolVar(&genPath, "path", false, "Generate path (-timesort is recommended)")
	flag.BoolVar(&pathIncludeZeroLoc, "include-zero-location", false, "Include locations with 0,0 coordinate into the path (won't work without -path)")
	flag.BoolVar(&kmz, "kmz", false, "Create KMZ file (zip the output directory)")
}

func main() {
	flag.Parse()
	handleHelp()
	checkCmd()
	setup()

	fmt.Println("Indexing images...")
	images, err := indexImages(imgDir)
	fatalIfErr(err)

	fmt.Println("Collecting images...")
	collectImages(images)

	fmt.Println("Generating KML document...")
	k, doc := getKmlDoc()

	if sortByTime {
		orderImagesByTime(images)
	}

	if genPath {
		generatePath(images, doc)
	}

	for i, img := range images {
		img.description = img.dateTime.String() + "<br>"
		//img.name = "Photo " + strconv.Itoa(i + 1)
		img.name = strconv.Itoa(i + 1)

		if img.latitude == 0 && img.longitude == 0 {
			fmt.Println(img.path, "Warning: GPSLatitude and GPSLongitude == 0")
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
		fmt.Println("Creating KMZ file...")
		zipFolderContents(outDir, outDir+"/doc.kmz")
	}
	fmt.Println("Done!")
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
Handles help flag -h. If the help is requested, prints program description and usage, and exits.
*/
func handleHelp() {
	if help {
		fmt.Println("photo-map")
		fmt.Println("An image gallery placed on a map!")
		fmt.Println("\nSee https://github.com/sykoram/photo-map for documentation and more information.")
		fmt.Println("\nUsage:")
		flag.PrintDefaults()
		os.Exit(0)
	}
}

/*
Setup:
Normalizes paths, sets outFilesDir;
Loads JSON or YAML file with custom image data if possible.
 */
func setup() {
	imgDir = normalizePath(imgDir)
	outDir = normalizePath(outDir)
	dataFilepath = normalizePath(dataFilepath)

	if dataFilepath != "" {
		var data dataObj
		var err error

		switch strings.ToLower(filepath2.Ext(dataFilepath)) {
		case ".json":
			data, err = loadJson(dataFilepath)
			fatalIfErr(err)
		case ".yaml":
			data, err = loadYaml(dataFilepath)
			fatalIfErr(err)
		}

		if data["items"] == nil {
			log.Fatalln("Cannot find key 'items' in the data file.")
		} else {
			dataFileItems = data["items"].(dataArr)
		}
	}
}

/*
Returns imagePlacemarks created using both internal and external images.
Internal images are collected from the rootDir.
Purely external images are loaded from the JSON or YAML data file.
The returned structs have kmlPaths already set.
 */
func indexImages(rootDir string) (images []imagePlacemark, err error) {
	images, err = getInternalImages(rootDir)
	if err != nil {
		return
	}
	externalImages, err := getExternalImages()
	if err != nil {
		return
	}
	images = append(images, externalImages...)

	for i := range images {
		images[i].setKmlPaths(isExternalPreferable, isExternalIconPreferable)
	}
	return
}

/*
Searches the given dir, collects images returns them as image structs. .thumbnail dirs are ignored.
 */
func getInternalImages(rootDir string) (images []imagePlacemark, err error) {
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
			images = append(images, prepareInternalImage(rootDir, path))
		}

		return nil
	})
	return
}

/*
Prepares an internal image struct: loads EXIF and JSON, sets properties, generates a thumbnail
 */
func prepareInternalImage(rootDir, rootRelPath string) imagePlacemark {
	img := imagePlacemark{
		path:    rootRelPath,
		rootDir: rootDir,
	}
	err := img.loadOrigExif(joinPaths(img.rootDir, img.path))
	if err != nil && exif.IsCriticalError(err) {
		log.Println("EXIF of", img.path, "has a critical error:", err)
	} else {
		img.applyDataFromExif()
	}

	// overwrite data from exif with data from json
	if dataFileItems != nil {
		for _, obj := range dataFileItems {
			if f, ok := obj.(dataObj)["file"]; ok {
				fs := normalizePath(f.(string))
				if fs == img.path {
					img.setCustomData(obj.(dataObj))
					img.applyCustomData()
				}
			}
		}
	}

	err = img.createThumbnail()
	if err != nil {
		img.iconPath = img.path
	}

	return img
}

/*
Returns imagePlacemarks with purely external images loaded from the JSON/YAML file.
 */
func getExternalImages() (images []imagePlacemark, err error) {
	images = make([]imagePlacemark, 0)
	if dataFileItems == nil {
		return
	}
	for _, obj := range dataFileItems {
		_, isExt := obj.(dataObj)["external"]
		if _, isInt := obj.(dataObj)["file"]; isExt && !isInt { // only pure external images without local files
			img := imagePlacemark{}
			img.setCustomData(obj.(dataObj))
			img.applyCustomData() // sets also externalPath
			images = append(images, img)
			fmt.Println(img.externalPath)
		}
	}
	return
}

/*
Copies necessary internal images or thumbnails to the output directory.
 */
func collectImages(images []imagePlacemark) {
	for _, img := range images {
		if img.isInternal {
			printIfErr(copyFile(img.rootDir + "/" + img.path, outDir + "/" + img.pathInKml))
		}
		if img.isIconInternal {
			printIfErr(copyFile(img.rootDir + "/" + img.iconPath, outDir + "/" + img.iconPathInKml))
		}
	}
}

/*
Orders images by its timestamp.
 */
func orderImagesByTime(images []imagePlacemark) {
	sort.Slice(images, func(i int, j int) bool {
		return images[i].dateTime.Before(images[j].dateTime)
	})
}

/*
Generates a path (line) that connects the images
 */
func generatePath(images []imagePlacemark, doc *kml.CompoundElement) {
	coords := make([]kml.Coordinate, 0)
	for _, img := range images {
		if pathIncludeZeroLoc || (img.longitude != 0 || img.latitude != 0) {
			coords = append(coords, kml.Coordinate{Lon: img.longitude, Lat: img.latitude})
		}
	}
	createLine(doc, coords)
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

