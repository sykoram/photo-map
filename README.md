# photo-map

A photo gallery placed on a map!

## Table of Contents
- [Features](#features)
- [Setup](#setup)
- [Usage](#usage)
  - [Arguments](#arguments)
  - [Modes](#modes)
  - [Custom data file](#custom-data-file)
  - [Viewing the results](#viewing-the-results)

## Features

- simple CLI program 
- generate a KML file with images placed on a map (can be opened in eg. Google Earth)
- location is automatically extracted from EXIF
- specify custom image information using a JSON or YAML file
- order images by time
- generate trip path
- zip KML and resources to KMZ file
- supports external images


## Setup

[Go](https://golang.org/) and [Git](https://git-scm.com/) have to be installed.

Clone or download this repo somewhere into `$GOPATH`, preferably into `$GOPATH/src/github.com/sykoram/photo-map`.

Download and install all dependencies:
```sh
go get ./...
```

And install the photo-map:
```sh
go install
```

This creates an executable inside `$GOBIN` (usually `$GOPATH/bin`), and photo-map should work now. You can try to run:

```sh
photo-map -h
```

## Usage

Use flag `-h` or `--help` to display the help.

The most basic command would be:
```sh
photo-map -i IMAGE_DIR -o OUTPUT_DIR
```


### Arguments

- `-i IMAGE_DIR`: Input directory with images (required)

- `-o OUTPUT_DIR`: Output directory

- `-mode MODE`: [Mode](#modes) of an image representation

- `-name NAME`: Project name

- `-data DATA_FILE`: Path to a [file with user-specified image data](#custom-data-file)

- `-timesort`: Order images by timestamp.

- `-path`: Draw a line between the images (`-timesort` is recommended).

- `-include-no-location`: Do not skip images without location. They are placed on \[0,0].

- `-base64`: Embed images in base64 into the KML document. \
  It may be a good idea to reduce size of the images; otherwise, the generated output KML/KMZ file might be large.

- `-kmz`: Zip the output directory into a one KMZ file.


### Modes

Different applications use different types of image representation. 

`g-earth-web` (**Google Earth Web**): `<gx:Carousel>` is used.

`g-earth-web-panel` (**Google Earth Web**): `<img>` tag inside HTML, `panel` balloon style.

`g-earth-pro` (**Google Earth Pro**): The image is an `<img>` tag inside a HTML balloon style (it is not in the description field).

`g-maps` (**Google Maps**): The image is a HTML `<img>` tag in the description.

`g-earth-photo-overlay` (**Google Earth Pro**): The image is placed above the map using PhotoOverlay.

**Google Earth mobile app** supports usually same modes as Google Earth Web


### Custom data file

Using the custom data file, you can specify some information about the images. This will overwrite information extracted from the EXIF. Both JSON and YAML files are supported, and they follow the same structure.

#### Structure

Inside the main JSON or YAML object, there has to be a key `items`, and its value is an array of objects. Each of these objects contains information about an image:

- `file` specifies the file (image). The path should be relative to the input directory (containing images), so it might be a good idea to put the JSON file also in this directory.

- `dateTime` sets the date and time using the EXIF format: `"2006:01:02 15:04:05"`. Any trailing spaces or null characters are trimmed.

- `timeZone` sets the time zone. It has to be either a valid [tz database name](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) or `"Local"`. If the `dateTime` is not specified in JSON for an image, the `dateTime` from EXIF will be recalculated using the difference between the EXIF time zone (if exists) and the specified zone. This can be used to correct the time and date.

- `latitude` and `longitude` define the GPS coordinations of the image. Positive for north and east, and negative for south and west.

- `external` specifies the absolute path to the corresponding image that is somewhere else (eg. on a website) and is not included in the KMZ file.

If a field is left out, the data from EXIF will not be overwritten. Unknown keys are ignored.

#### YAML example

```yaml
---
items:
- file: path/to/image.jpg
  dateTime: 2006:01:02 15:04:05
  timeZone: UTC
  latitude: 50.09
  longitude: 14.4
  external: https://example.com/path/to/image.jpg

- file: image2.jpg
  latitude: 50.087
  longitude: 14.42

- file: path/to/image3.jpg
  external: https://example.com/path/to/image3.jpg
```

#### JSON example

```json
{"items": [
  {
    "file": "path/to/image.jpg",
    "dateTime": "2006:01:02 15:04:05",
    "timeZone": "UTC",
    "latitude": 50.09,
    "longitude": 14.4,
    "external": "https://example.com/path/to/image.jpg"
  },
  {
    "file": "image2.jpg",
    "latitude": 50.087,
    "longitude": 14.42
  },
  {
    "file": "path/to/image3.jpg",
    "external": "https://example.com/path/to/image3.jpg"
  }
]}
```


### Viewing the results

#### Google Earth Web

1. open [Google Earth](https://earth.google.com/web/) in a browser
2. in menu (three horizontal lines in the left upper corner), click on "Projects"
3. click "Open" or "New project" > "Import KML file from computer" and select the generated KML or KMZ file > "Present"

#### Google Earth (mobile)

1. open Google Earth app
2. in menu (three horizontal lines in the left upper corner), tap on "Projects"
3. tap "Open" > "Import KML file" and select the KML or KMZ file > "Present"

#### Google Earth Pro (desktop)

1. start Google Earth Pro
2. click "File" > "Open" > select the generated KML or KMZ file

#### Google Maps

1. go to [Google My Maps](https://mymaps.google.com)
2. click "Create a new map"
3. under New layer or Untitled layer, click "Import", and select the generated KML or KMZ file


