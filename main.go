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
var includeNoLocation bool
var kmz bool
var mode string
var base64images bool

// other global variables
var dataFileItems dataArr
var isExternalPreferable = true
var isExternalIconPreferable = false

var availableModes = map[string]func (el *kml.CompoundElement, img imagePlacemark){
	"g-earth-web": addGxCarouselPlacemark,
	"g-earth-web-panel": addGxPanelHtmlImage,
	"g-earth-pro": addHtmlImagePlacemark,
	"g-maps": addDescriptionImagePlacemark,
	"g-earth-photo-overlay": addPhotoOverlayPlacemark,
}

func init() {
	flag.BoolVar(&help, "h", false, "")
	flag.BoolVar(&help, "help", false, "")

	flag.StringVar(&imgDir, "i", "", "Input directory with images (required)")
	flag.StringVar(&outDir, "o", "", "Output directory for generated KML file and other copied files. Must be empty or not exist! (required)")

	flag.StringVar(&mode, "mode", "g-earth-web", fmt.Sprintf("Different apps use different types of image representation: %s", getModesKeys()))
	flag.StringVar(&dataFilepath, "data", "", "JSON or YAML file with custom image information\n(it has higher priority than the EXIF info)")
	flag.BoolVar(&sortByTime, "timesort", false, "Sort images by time (DateTimeOriginal eventually DateTime)")
	flag.BoolVar(&genPath, "path", false, "Generate path (-timesort is recommended)")
	flag.BoolVar(&includeNoLocation, "include-no-location", false, "Do not skip images with no location (they are placed on [0,0])")
	flag.BoolVar(&kmz, "kmz", false, "Create KMZ file (zip the output directory)")
	flag.BoolVar(&base64images, "base64", false, "Embed images in base64 in the KML file")
}

func main() {
	flag.Parse()
	handleHelp()
	checkCmd()
	setup()

	fmt.Println("Indexing images...")
	images, err := indexImages(imgDir)
	fatalIfErr(err)

	if !base64images {
		fmt.Println("Collecting images...")
		collectImages(images)
	}

	fmt.Println("Generating KML document...")
	k, doc := getKmlDoc()

	if sortByTime {
		orderImagesByTime(images)
	}

	if genPath {
		generatePath(images, doc)
	}

	i := 1
	for _, img := range images {
		// todo collect image/icon here (only if used)
		if base64images {
			err := setBase64Images(&img)
			printIfErr(err)
		}
		warnIfNoLocation(img)
		if img.hasLocation || includeNoLocation {
			img.description = img.dateTime.String()
			img.name = strconv.Itoa(i)
			i++

			availableModes[mode](doc, img)
		} // todo else clean
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

	if _, ok := availableModes[mode]; !ok {
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
Orders images by their timestamp. Warns if an image has no dateTime.
 */
func orderImagesByTime(images []imagePlacemark) {
	for _, img := range images {
		if !img.hasDateTime {
			path := ""
			if img.isInternal {
				path = img.path
			} else {
				path = img.externalPath
			}
			log.Printf("%s has no dateTime", path)
		}
	}

	sort.Slice(images, func(i int, j int) bool {
		return images[i].dateTime.Before(images[j].dateTime)
	})
}

/*
Generates a path (line) that connects the images.
Images with no location are skipped.
 */
func generatePath(images []imagePlacemark, doc *kml.CompoundElement) {
	coords := make([]kml.Coordinate, 0)
	for _, img := range images {
		if img.hasLocation {
			ic := kml.Coordinate{Lon: img.longitude, Lat: img.latitude}
			if len(coords) == 0 || coords[len(coords)-1] != ic {  // ignore coordinates if same as previous
				coords = append(coords, ic)
			}
		}
	}
	createLine(doc, coords)
}

/*
Sets pathInKml to base64 data of the image file if the image is internal. Same for the icon.
 */
func setBase64Images(img *imagePlacemark) error {
	if img.isInternal {
		mimeType, err := getImageMimeType(strings.Replace(filepath2.Ext(img.pathInKml), ".", "", 1))
		if err != nil {
			return err
		}

		b64Data, err := getBase64Data(img.rootDir + "/" + img.path)
		if err != nil {
			return err
		}

		img.pathInKml = "data:" + mimeType + ";base64,"
		img.pathInKml += string(b64Data)
	}
	if img.isIconInternal {
		mimeType, err := getImageMimeType(strings.Replace(filepath2.Ext(img.iconPathInKml), ".", "", 1))
		if err != nil {
			return err
		}

		b64Data, err := getBase64Data(img.rootDir + "/" + img.iconPath)
		if err != nil {
			return err
		}

		img.iconPathInKml = "data:" + mimeType + ";base64,"
		img.iconPathInKml += string(b64Data)
	}
	return nil
}

/*
Warns if the image has no location.
 */
func warnIfNoLocation(img imagePlacemark) {
	if !img.hasLocation {
		path := ""
		if img.isInternal {
			path = img.path
		} else {
			path = img.externalPath
		}
		log.Println(path, "has no location")
	}
}

/*
Returns string keys of the modes
 */
func getModesKeys() []string {
	var sm []string
	for key := range availableModes {
		sm = append(sm, key)
	}
	return sm
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
