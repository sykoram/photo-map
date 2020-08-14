package main

import (
	"encoding/xml"
	"github.com/twpayne/go-kml"
	"image/color"
)

var iconScale = 2.0

var descImgMaxWidth = "800px"
var descImgMaxHeight = "800px"

var pathName = "Path"
var pathLineColor = color.RGBA{R: 0x00, G: 0xff, B: 0x7f, A: 0xff}
var pathLineWidth = 4.0

/*
Returns a KML element and its Document element.
 */
func getKmlDoc() (kmlEl *kml.CompoundElement, docEl *kml.CompoundElement) {
	docEl = kml.Document()
	kmlEl = kml.KML(docEl)
	kmlEl.Attr = append(kmlEl.Attr,
		xml.Attr{Name: xml.Name{Local: "xmlns:gx"}, Value: "http://www.google.com/kml/ext/2.2"},
		xml.Attr{Name: xml.Name{Local: "xmlns:kml"}, Value: "http://www.opengis.net/kml/2.2"},
		xml.Attr{Name: xml.Name{Local: "xmlns:atom"}, Value: "http://www.w3.org/2005/Atom"},
	)
	return kmlEl, docEl
}

/*
Add a image placemark into the given element (usually Document or Folder).
The description image placemark has a HTML img tag in the description.
*/
func addDescriptionImagePlacemark(el *kml.CompoundElement, img imagePlacemark) {
	el.Add(
		kml.Placemark(
			kml.Name(img.name),
			kml.Description(`<!DOCTYPE html><html><head></head><body>
<p>`+img.description+`</p>
<img src="`+img.pathInKml+`" style="display: block; max-width:`+descImgMaxWidth+`; max-height:`+descImgMaxHeight+`; width: auto; height: auto;" />
</body></html>`),
			kml.Point(
				kml.Coordinates(kml.Coordinate{Lat: img.latitude, Lon: img.longitude}),
			),
			kml.Style(
				kml.Scale(iconScale),
				kml.IconStyle(
					kml.Icon(
						kml.Href(img.iconPathInKml),
					),
				),
			),
		),
	)
}

/*
Add a image placemark into the given element (usually Document or Folder).
The HTML image placemark has a HTML balloon style with a img tag.
 */
func addHtmlImagePlacemark(el *kml.CompoundElement, img imagePlacemark) {
	el.Add(
		kml.Placemark(
			kml.Name(img.name),
			kml.Description(img.description),
			kml.Point(
				kml.Coordinates(kml.Coordinate{Lat: img.latitude, Lon: img.longitude}),
			),
			kml.Style(
				kml.Scale(iconScale),
				kml.IconStyle(
					kml.Icon(
						kml.Href(img.iconPathInKml),
					),
				),
				kml.BalloonStyle(
					kml.Text(`
<!DOCTYPE html>
<html>
<head>
	<style>
		img {display: block; max-width:`+descImgMaxWidth+`; max-height:`+descImgMaxHeight+`; width: auto; height: auto;}
	</style>
</head>
<body>
	<p><b>$[name]</b></p>
	<p>$[description]</p>
	<img src="`+img.pathInKml+`"/>
</body>
</html>
`),
				),
			),
		),
	)
}

/*
Add a image placemark into the given element (usually Document or Folder).
The photo overlay placemark uses PhotoOverlay - the image is not in the description/HTML, but placed above the map.
 */
func addPhotoOverlayPlacemark(el *kml.CompoundElement, img imagePlacemark) {
	w, l := img.width, img.length
	if w == 0 || l == 0 { // prevent divide-by-zero exception
		w = 1
		l = 1
	}
	id := img.name 	// todo better id
	coordinate := kml.Coordinate{Lat: img.latitude, Lon: img.longitude}
	photoOverlay := kml.PhotoOverlay(
		kml.Name(img.name),
		kml.Description(`<!DOCTYPE html><html><head></head><body>
<a href="#`+id+`">Click here to fly into photo</a><br>
</body></html>`),
		kml.Open(false),
		kml.Visibility(true),
		kml.Icon(
			kml.Href(img.pathInKml),
			//kml.Href(iconSrc),
		),
		// todo tilt 45 deg
		kml.Camera(
			kml.Latitude(coordinate.Lat),
			kml.Longitude(coordinate.Lon),
			kml.Altitude(10),
			kml.Tilt(90),
		),
		kml.Point(
			kml.Coordinates(coordinate),
		),
		kml.Rotation(0),
		kml.ViewVolume(
			kml.Near(10),
			kml.LeftFOV(float64(w/l*-20)),
			kml.RightFOV(float64(w/l*20)),
			kml.BottomFOV(-20),
			kml.TopFOV(20),
		),
		kml.Shape(kml.ShapeRectangle),
		// todo ImagePyramid
		kml.Style(
			kml.Scale(iconScale),
			kml.IconStyle(
				kml.Icon(
					kml.Href(img.iconPathInKml),
				),
			),
			kml.BalloonStyle(
				kml.DisplayMode(kml.DisplayModeHide),
			),
		),
	)
	photoOverlay.Attr = append(photoOverlay.Attr, xml.Attr{
		Name:  xml.Name{
			Space: "",
			Local: "id",
		},
		Value: id,
	})
	el.Add(photoOverlay)
}

/*
Add a image placemark into the given element (usually Document or Folder).
This placemark uses gx:Carousel.
fixme
 */
func addGxCarouselPlacemark(el *kml.CompoundElement, img imagePlacemark) {
	el.Add(
		kml.Placemark(
			kml.Name(img.name),
			kml.Description(`<!DOCTYPE html><html><head></head><body>
<p>`+img.description+`</p>
</body></html>`),
			kml.Point(
				kml.Coordinates(kml.Coordinate{Lat: img.latitude, Lon: img.longitude}),
			),
			kml.Style(
				kml.Scale(iconScale),
				kml.IconStyle(
					kml.Icon(
						kml.Href(img.iconPathInKml),
					),
				),
			),
			newCompoundEl("gx:Carousel").Add(
				newCompoundEl("gx:Image").Add(
					newSimpleEl("gx:ImageUrl", img.pathInKml),
				),
			),
		),
	)
}

/*
Returns a new KML compound element.
 */
func newCompoundEl(name string) *kml.CompoundElement {
	el := new(kml.CompoundElement)
	el.StartElement = xml.StartElement{
		Name: xml.Name{Local: name},
		Attr: nil,
	}
	return el
}

/*
Returns a new KML simple element <space>:<local>
 */
func newSimpleEl(name, value string) *kml.SimpleElement {
	// hack - rename value element
	el := kml.Value(value)
	el.StartElement = xml.StartElement{
		Name: xml.Name{Local: name},
		Attr: nil,
	}
	return el
}

/*
Generates a line connecting the given coordinates.
 */
func generatePath(el *kml.CompoundElement, coordinates []kml.Coordinate) {
	el.Add(
		kml.Placemark(
			kml.Name(pathName),
			kml.Style(
				kml.LineStyle(
					kml.Color(pathLineColor),
					kml.Width(pathLineWidth),
				),
			),
			kml.LineString(
				kml.Extrude(true),
				kml.Tessellate(true),
				kml.Coordinates(coordinates...),
			),
		),
	)
}