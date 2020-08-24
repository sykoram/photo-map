package main

import (
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"log"
	"math"
	"os"
	"strings"
	"time"
)

type imagePlacemark struct {
	path       string // location of the file relative to the root dir (should be normalized) (empty if pure external image)
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
	hasDateTime  bool

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
		i.hasDateTime = true
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
		if err == nil {
			i.hasDateTime = true
		}
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
		float, err := getFloat64(lat)
		if err == nil {
			i.latitude = float
			i.hasLocation = true
		} else {
			i.latitude = 0
			i.hasLocation = false
		}
	}
	if lon, ok := i.customData["longitude"]; ok {
		float, err := getFloat64(lon)
		if err == nil {
			i.longitude = float
			i.hasLocation = true
		} else {
			i.longitude = 0
			i.hasLocation = false
		}
	}
}

/*
Sets paths of image and icon used in KML doc based on whether external or internal path is preferable.
 */
func (i *imagePlacemark) setKmlPaths(preferExternal, preferExternalIcon bool) {
	if preferExternal && i.externalPath != "" || i.path == "" {
		i.pathInKml = i.externalPath
	} else {
		i.pathInKml = joinPaths("files", i.path)
		i.isInternal = true
	}

	// icon
	if preferExternalIcon && i.iconExternalPath != "" || i.iconPath == "" {
		i.iconPathInKml = i.iconExternalPath
	} else {
		i.iconPathInKml = joinPaths("files", i.iconPath)
		i.isIconInternal = true
	}
}

/*
Tries to convert the interface{} to float64.
 */
func getFloat64(unk interface{}) (float64, error) {
	switch i := unk.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint:
		return float64(i), nil
	default:
		return math.NaN(), fmt.Errorf("non-numeric type could not be converted to float")
	}
}