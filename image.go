package main

import (
	"github.com/rwcarlsen/goexif/exif"
	"log"
	"os"
	filepath2 "path/filepath"
	"strings"
	"time"
)

type imagePlacemark struct {
	path       string // location of the file either relative to the root dir (should be normalized) (empty if pure external image)
	iconPath   string // location of the thumbnail (or the actual image) relative to the root dir (empty if pure external image)
	rootDir    string // actual location of the root dir (should be normalized) (empty if pure external image)

	isInternal     bool
	isIconInternal bool

	externalPath     string
	iconExternalPath string

	pathInKml     string // path used in KML file
	iconPathInKml string // ~

	origExif   *exif.Exif
	customData dataObj

	name 		 string
	description  string
	dateTime	 time.Time
	latitude	 float64
	longitude	 float64

	hasLocation  bool

	width  int64
	length int64
}

/*
Decodes and returns the EXIF of the file.
*/
func (i *imagePlacemark) loadOrigExif(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	i.origExif, err = exif.Decode(file)
	return err
}

/*
Sets image properties according to the EXIF of the image.
 */
func (i *imagePlacemark) applyDataFromExif() {
	if i.origExif == nil {
		return
	}

	// dateTime
	t, err := i.origExif.DateTime()
	if err == nil {
		i.dateTime = t
	}

	// latitude & longitude
	lat, lon, err := i.origExif.LatLong()
	if err == nil {
		i.latitude = lat
		i.longitude = lon
		i.hasLocation = true
	}

	// width & height
	if w, err := i.origExif.Get(exif.ImageWidth); err == nil {
		i.width, _ = w.Int64(0)
	}
	if l, err := i.origExif.Get(exif.ImageLength); err == nil {
		i.length, _ = l.Int64(0)
	}
}

/*
Sets the field customData to the parameter
 */
func (i *imagePlacemark) setCustomData(data dataObj) {
	i.customData = data
}

/*
Sets image properties according to the customData object for the image
Used JSON/YAML fields/keys: "external" string, "dateTime" string, "timeZone" string, "latitude" float64, "longitude" float64
 */
func (i *imagePlacemark) applyCustomData() {
	if i.customData == nil {
		return
	}
	var err error

	// external path
	if ext, ok := i.customData["external"]; ok {
		i.externalPath = ext.(string)
		i.iconExternalPath = ext.(string)
	}

	// dateTime (+ timeZone)
	if dt, ok := i.customData["dateTime"]; ok {
		exifTimeLayout := "2006:01:02 15:04:05"
		dateStr := strings.Trim(dt.(string), "\x00 ")
		location := time.Local
		if tz, ok := i.customData["timeZone"]; ok {
			if loc, err := time.LoadLocation(tz.(string)); err == nil {
				location = loc
			} else {
				log.Println(err)
			}
		}
		i.dateTime, err = time.ParseInLocation(exifTimeLayout, dateStr, location)
		printIfErr(err)
	}

	// change timeZone only
	if tz, ok := i.customData["timeZone"]; ok {
		if _, dtExists := i.customData["dateTime"]; dtExists == false {
			if newLoc, err := time.LoadLocation(tz.(string)); err == nil {
				// calculate the difference between the new and the old timezone
				oldLoc := i.dateTime.Location()
				someTimeOldLoc := time.Date(2000,1,1,0,0,0,0, oldLoc)
				someTimeNewLoc := time.Date(2000,1,1,0,0,0,0, newLoc)
				timeZonesDiff := someTimeOldLoc.Sub(someTimeNewLoc)
				// add the difference to the dateTime to fix the time zone (change the timestamp)
				i.dateTime = i.dateTime.Add(timeZonesDiff)
			} else {
				log.Println(err)
			}
		}
	}

	// latitude & longitude
	if lat, ok := i.customData["latitude"]; ok {
		i.latitude = lat.(float64)
		i.hasLocation = true
	}
	if lon, ok := i.customData["longitude"]; ok {
		i.longitude = lon.(float64)
		i.hasLocation = true
	}
}

/*
Creates a JPEG thumbnail file from the EXIF thumbnail data and sets img.thumbnailSrc. The thumbnail file is located in .thumbnail dir.
*/
func (i *imagePlacemark) createThumbnail() (err error) {
	data, err := i.origExif.JpegThumbnail()
	if err != nil || len(data) == 0 {
		return
	}

	imgDir, imgName := filepath2.Split(i.path)
	tRelPath := imgDir + ".thumbnails/" + imgName
	f, err := createFile(joinPaths(i.rootDir, tRelPath))
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return
	}
	i.iconPath = tRelPath
	return
}

/*
Sets paths of image and icon used in KML doc based on whether external or internal path is preferable.
 */
func (i *imagePlacemark) setKmlPaths(preferExternal, preferExternalIcon bool) {
	if preferExternal && i.externalPath != "" || i.path == "" {
		i.pathInKml = i.externalPath
	} else {
		i.pathInKml = "files/" + i.path
		i.isInternal = true
	}

	// icon
	if preferExternalIcon && i.iconExternalPath != "" || i.iconPath == "" {
		i.iconPathInKml = i.iconExternalPath
	} else {
		i.iconPathInKml = "files/" + i.iconPath
		i.isIconInternal = true
	}
}